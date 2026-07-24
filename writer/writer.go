package writer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"net/url"
	"sort"
	"sync"

	// Image imports for decoding
	_ "image/gif"  // Register GIF decoder for artwork.DecodeConfig
	_ "image/jpeg" // Register JPEG decoder for artwork.DecodeConfig
	_ "image/png"  // Register PNG decoder for artwork.DecodeConfig

	_ "golang.org/x/image/bmp"  // Register BMP decoder for artwork.DecodeConfig
	_ "golang.org/x/image/webp" // Register WebP decoder for artwork.DecodeConfig

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	profile "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	types "github.com/mutablelogic/go-server/pkg/types"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// Writer is a wrapper around an AVFormatContext that provides a higher-level
// interface for writing media files. It embeds an Encoder, giving it Add,
// Encode, Flush and FrameSize for free — one codec context per stream,
// opened alongside each stream's AVStream in open().
type Writer struct {
	sync.Mutex
	opts
	*Encoder
	output  *ff.AVFormatContext
	header  bool           // Track if header was successfully written (for Close)
	artwork map[int][]byte // Map of stream index to artwork data
	once    sync.Once
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new writer with a URL and options
func Create(url *url.URL, output *profile.Output, opts ...Opt) (*Writer, error) {
	self := new(Writer)

	// Set writer options
	if err := self.opts.apply(opts...); err != nil {
		return nil, err
	}

	// Create the encoder
	encoder, err := NewEncoder(self.writePacket)
	if err != nil {
		return nil, err
	} else {
		self.Encoder = encoder
	}

	// Check URL and output
	if url == nil || output == nil || output.Context() == nil {
		return nil, gomedia.ErrBadParameter.Withf("url and output must be non-nil")
	} else if len(self.streams) == 0 {
		return nil, gomedia.ErrBadParameter.Withf("at least one stream must be provided")
	}

	// Allocate the output media context
	ctx, err := ff.AVFormat_create_file(url.String(), output.Context())
	if err != nil {
		return nil, err
	} else {
		self.output = ctx
	}

	// Continue to open the stream
	return self.open(output)
}

// Create a new writer with an io.Writer and options
func NewWriter(w io.Writer, output *profile.Output, opts ...Opt) (*Writer, error) {
	self := new(Writer)

	// Set writer options
	if err := self.opts.apply(opts...); err != nil {
		return nil, err
	}

	// Create the encoder
	encoder, err := NewEncoder(self.writePacket)
	if err != nil {
		return nil, err
	} else {
		self.Encoder = encoder
	}

	// Check writer and output
	if w == nil || output == nil || output.Context() == nil {
		return nil, gomedia.ErrBadParameter.Withf("writer and output must be non-nil")
	} else if len(self.streams) == 0 {
		return nil, gomedia.ErrBadParameter.Withf("at least one stream must be provided")
	}

	// Get filename from writer
	var filename string
	if w_, ok := w.(gomedia.NamedWriter); ok {
		filename = w_.Name()
	}

	// Allocate the AVIO context
	avio := ff.AVFormat_avio_alloc_context(ioBufferSize, true, &io_callback{w})
	if avio == nil {
		return nil, gomedia.ErrInternalError.With("failed to allocate avio context")
	} else if ctx, err := ff.AVFormat_open_writer(avio, output.Context(), filename); err != nil {
		return nil, err
	} else {
		self.output = ctx
	}

	// Continue with open
	return self.open(output)
}

// Close a writer and release resources
func (w *Writer) Close() error {
	var result error

	// Mutex lock to ensure thread safety
	w.Lock()
	defer w.Unlock()

	// Write the trailer only if header was successfully written and the
	// context hasn't already been freed by a prior Close call.
	if w.header && w.output != nil {
		// Flush any internally buffered/interleaving packets.
		result = errors.Join(result, ff.AVFormat_interleaved_write_frame(w.output, nil))

		// Write the trailer to the output file
		result = errors.Join(result, ff.AVFormat_write_trailer(w.output))
	}

	// Close the encoder, freeing every codec context it owns. This is
	// independent of the AVFormatContext freed below: Encoder's contexts
	// were never attached to an AVStream, so nothing else could free them.
	if w.Encoder != nil {
		result = errors.Join(result, w.Encoder.Close())
	}

	// Free output resources
	if w.output != nil {
		result = errors.Join(result, ff.AVFormat_close_writer(w.output))
	}

	// Free resources and clear artwork data
	w.output = nil
	w.header = false
	w.artwork = nil

	// Return any errors
	return result
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (writer *Writer) open(output *profile.Output) (*Writer, error) {
	var result error

	// Initialize the artwork map
	writer.artwork = make(map[int][]byte)

	// Create streams in a deterministic order (map iteration order is
	// randomized, but stream creation order determines each stream's
	// physical position in the output container)
	ids := make([]int, 0, len(writer.streams))
	for id := range writer.streams {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	// Some formats (e.g. mp4/mov) want the codec to emit extradata for the
	// muxer to store in the header, rather than repeating it in-band in
	// every keyframe packet.
	var codecFlags []ff.AVCodecFlag
	if writer.output.Output().Flags().Is(ff.AVFMT_GLOBALHEADER) {
		codecFlags = append(codecFlags, ff.AV_CODEC_FLAG_GLOBAL_HEADER)
	}

	for _, id := range ids {
		profile := writer.streams[id]

		// Create stream
		stream := ff.AVFormat_new_stream(writer.output, profile.Codec().Context())
		if stream == nil {
			result = errors.Join(result, gomedia.ErrInternalError.Withf("failed to allocate stream for profile %q", profile.UUID()))
			continue
		} else {
			// Set stream index 0...n-1
			stream.SetId(id)
		}

		// Open a codec context for this stream so frames can actually be
		// encoded. This must happen before copying codec parameters below:
		// opening the codec is what generates extradata some codecs need
		// (e.g. OpusHead for libopus, AudioSpecificConfig for aac), and the
		// muxer needs that extradata on the stream, not just the raw
		// parameters the profile requested.
		if err := writer.Encoder.Add(id, profile, codecFlags...); err != nil {
			result = errors.Join(result, err)
			continue
		}

		// Copy codec parameters from the now-opened codec context
		par, err := writer.Encoder.Par(id)
		if err != nil {
			result = errors.Join(result, err)
			continue
		}
		err = ff.AVCodec_parameters_copy(stream.CodecPar(), par)
		ff.AVCodec_parameters_free(par)
		if err != nil {
			result = errors.Join(result, err)
			continue
		}

		// Set timebase if specified. Subtitle profiles have none of their
		// own (no sample/frame rate to derive one from) - give the stream
		// the same default Encoder.Add gave the codec context, so packets
		// built against that timebase need no further rescaling here.
		if timebase := profile.TimeBase(); timebase != nil {
			stream.SetTimeBase(types.Value(timebase))
		} else if ff.AVMediaType(profile.Type()) == ff.AVMEDIA_TYPE_SUBTITLE {
			stream.SetTimeBase(subtitleTimeBase)
		}
	}

	// Bail out if any stream failed to be created
	if result != nil {
		return nil, errors.Join(result, writer.Close())
	}

	// Build the format-level options dictionary from the output profile
	dict, err := dictFromOpts(output.Opts)
	if err != nil {
		return nil, errors.Join(err, writer.Close())
	}
	defer ff.AVUtil_dict_free(dict)

	// Allocate metadata dictionary
	metadata := ff.AVUtil_dict_alloc()
	if metadata == nil {
		return nil, errors.Join(gomedia.ErrInternalError.With("unable to allocate metadata dictionary"), writer.Close())
	}
	// Note: No defer free - ownership transferred to output context via SetMetadata

	// Add metadata entries (but store artwork for later)
	for _, entry := range writer.metadata {
		// Add artwork streams
		if entry.Key() == gomedia.MetaArtwork {
			// Create stream
			stream := ff.AVFormat_new_stream(writer.output, nil)
			if stream == nil {
				return nil, errors.Join(gomedia.ErrInternalError.Withf("failed to allocate stream for artwork"), writer.Close())
			} else if config, _, err := image.DecodeConfig(bytes.NewReader(entry.Bytes())); err != nil {
				return nil, errors.Join(gomedia.ErrBadParameter.Withf("failed to decode artwork image: %w", err), writer.Close())
			} else {
				stream.CodecPar().SetCodecType(ff.AVMEDIA_TYPE_VIDEO)
				stream.CodecPar().SetCodecID(codecFromImageData(entry.Bytes()))
				stream.CodecPar().SetWidth(config.Width)
				stream.CodecPar().SetHeight(config.Height)
				stream.SetDisposition(ff.AV_DISPOSITION_ATTACHED_PIC)
			}

			// Store the artwork data, keyed by the stream's physical index
			// (its id is never set for artwork streams, so it stays 0)
			key := stream.Index()
			if _, exists := writer.artwork[key]; exists {
				return nil, errors.Join(gomedia.ErrBadParameter.Withf("stream %d already has artwork", key), writer.Close())
			} else {
				writer.artwork[key] = entry.Bytes()
			}

			// Continue without adding this entry to the metadata dictionary
			continue
		}

		// Ignore empty keys and values
		if entry.Key() == "" || entry.Value() == "" {
			continue
		}

		// Set dictionary entry
		if err := ff.AVUtil_dict_set(metadata, entry.Key(), entry.Value(), ff.AV_DICT_APPEND); err != nil {
			ff.AVUtil_dict_free(metadata)
			return nil, errors.Join(err, writer.Close())
		}
	}

	// Write the header, consuming recognized options from the dictionary
	writer.output.SetMetadata(metadata)
	if err := ff.AVFormat_write_header(writer.output, dict); err != nil {
		return nil, errors.Join(err, writer.Close())
	} else {
		writer.header = true
	}

	// Any keys left in the dictionary were not recognized by the muxer
	if keys := ff.AVUtil_dict_keys(dict); len(keys) > 0 {
		return nil, errors.Join(gomedia.ErrBadParameter.Withf("invalid output options: %v", keys), writer.Close())
	}

	// Return the writer
	return writer, nil
}

// writePacket is the embedded Encoder's packet callback: it muxes every
// packet the encoder produces straight into the output. A nil packet (the
// encoder's end-of-batch signal) is a no-op.
func (w *Writer) writePacket(packet *ff.AVPacket) error {
	w.Lock()
	defer w.Unlock()

	var result error
	w.once.Do(func() {
		pkt := ff.AVCodec_packet_alloc()
		if pkt == nil {
			result = errors.New("failed to allocate artwork packet")
			return
		}

		// Write artwork packets in a deterministic order (map iteration
		// order is randomized, but stream index order is not)
		indices := make([]int, 0, len(w.artwork))
		for index := range w.artwork {
			indices = append(indices, index)
		}
		sort.Ints(indices)

		for _, index := range indices {
			data := w.artwork[index]

			// Copy artwork data to packet
			if err := ff.AVCodec_packet_from_data(pkt, data); err != nil {
				result = errors.Join(result, err)
			} else {
				pkt.SetStreamIndex(index)
				pkt.SetFlags(ff.AV_PKT_FLAG_KEY)
				if err := ff.AVFormat_write_frame(w.output, pkt); err != nil {
					result = errors.Join(result, err)
				}
			}

			// Release packet memory immediately after writing
			ff.AVCodec_packet_unref(pkt)
		}
		ff.AVCodec_packet_free(pkt)
	})

	if result != nil {
		return result
	} else if packet == nil {
		return nil
	} else if w.output == nil {
		return gomedia.ErrInternalError.With("writer is closed")
	} else {
		return ff.AVFormat_interleaved_write_frame(w.output, packet)
	}
}

// dictFromOpts unmarshals a JSON object of output options into an AVDictionary
// suitable for AVFormat_write_header. Always returns a non-nil dictionary on
// success (empty if raw is empty), which the caller must free.
func dictFromOpts(raw json.RawMessage) (*ff.AVDictionary, error) {
	dict := ff.AVUtil_dict_alloc()
	if len(raw) == 0 {
		return dict, nil
	}

	var opts map[string]any
	if err := json.Unmarshal(raw, &opts); err != nil {
		ff.AVUtil_dict_free(dict)
		return nil, gomedia.ErrBadParameter.Withf("invalid output options: %w", err)
	}

	for key, value := range opts {
		if err := ff.AVUtil_dict_set(dict, key, fmt.Sprint(value), ff.AV_DICT_APPEND); err != nil {
			ff.AVUtil_dict_free(dict)
			return nil, gomedia.ErrBadParameter.Withf("invalid output option %q: %w", key, err)
		}
	}

	return dict, nil
}

// Detect codec ID from image data using content type detection
func codecFromImageData(data []byte) ff.AVCodecID {
	contentType := http.DetectContentType(data)

	switch contentType {
	case "image/jpeg":
		return ff.AV_CODEC_ID_MJPEG
	case "image/png":
		return ff.AV_CODEC_ID_PNG
	case "image/gif":
		return ff.AV_CODEC_ID_GIF
	case "image/bmp":
		return ff.AV_CODEC_ID_BMP
	case "image/webp":
		return ff.AV_CODEC_ID_WEBP
	default:
		// Default to JPEG for unknown image types
		return ff.AV_CODEC_ID_MJPEG
	}
}
