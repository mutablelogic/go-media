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
	"time"

	// Image imports for decoding
	_ "image/gif"  // Register GIF decoder for artwork.DecodeConfig
	_ "image/jpeg" // Register JPEG decoder for artwork.DecodeConfig
	_ "image/png"  // Register PNG decoder for artwork.DecodeConfig

	_ "golang.org/x/image/bmp"  // Register BMP decoder for artwork.DecodeConfig
	_ "golang.org/x/image/webp" // Register WebP decoder for artwork.DecodeConfig

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	frame "github.com/mutablelogic/go-media/frame"
	profile "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	types "github.com/mutablelogic/go-server/pkg/types"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// Writer is a wrapper around an AVFormatContext that provides a higher-level
// interface for writing media files. It embeds an Encoder, giving it Add and
// FrameSize for free — one codec context per stream, opened alongside each
// stream's AVStream in open(). Encode and Flush are overridden (see below) to
// resample/rescale a frame into the exact format its stream's codec expects
// before handing it to the embedded Encoder.
type Writer struct {
	sync.Mutex
	opts
	*Encoder
	output     *ff.AVFormatContext
	header     bool               // Track if header was successfully written (for Close)
	artwork    map[int][]byte     // Map of stream index to artwork data
	resamplers map[int]*Resampler // Map of stream index to its resampler, if any (audio/video only)
	once       sync.Once
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

	// Flush every stream first, so any samples not yet forming a full chunk
	// (via its resampler) and any packets still buffered inside its codec
	// (e.g. B-frame reordering delay) are written before the trailer, even
	// if the caller never called Flush itself. A stream the caller already
	// flushed returns io.EOF here - that just means there was nothing left,
	// not a real error.
	//
	// This must happen before the lock below is taken: Flush's packets flow
	// through the embedded Encoder's callback (writePacket), which takes
	// the same lock itself on every call - holding it here too would
	// deadlock.
	if w.header && w.output != nil {
		for stream := range w.streams {
			if err := w.Flush(stream); err != nil && !errors.Is(err, io.EOF) {
				result = errors.Join(result, err)
			}
		}
	}

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

	// Close every stream's resampler, freeing the swr/sws contexts they own -
	// already flushed above, so nothing left to drain here.
	for _, r := range w.resamplers {
		result = errors.Join(result, r.Close())
	}
	w.resamplers = nil

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
// PUBLIC METHODS

// Encode converts f via its stream's resampler (built in open(), if the
// stream is audio/video) into the exact format its codec was opened with,
// then hands the result to the embedded Encoder. Frames for a stream with no
// registered resampler (subtitles) pass straight through.
func (w *Writer) Encode(f frame.Frame) error {
	if f == nil {
		return gomedia.ErrBadParameter.With("nil frame")
	}
	r, ok := w.resamplers[f.Stream()]
	if !ok {
		return w.Encoder.Encode(f)
	}
	return r.Process(f, w.Encoder.Encode)
}

// Flush drains stream's resampler (if any) of any samples not yet forming a
// full frame, then flushes the encoder. Resampler.Process(nil, ...) is
// documented as a no-op for video/passthrough, so this only does real work
// for audio. Safe to call more than once for the same stream - the encoder
// returns io.EOF for a stream already flushed, rather than an error.
func (w *Writer) Flush(stream int) error {
	if r, ok := w.resamplers[stream]; ok {
		if err := r.Process(nil, w.Encoder.Encode); err != nil {
			return err
		}
	}
	return w.Encoder.Flush(stream)
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (writer *Writer) open(output *profile.Output) (*Writer, error) {
	var result error

	// Initialize the artwork and resampler maps
	writer.artwork = make(map[int][]byte)
	writer.resamplers = make(map[int]*Resampler)

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

		// Build a resampler for audio/video streams, so Writer.Encode can
		// convert an incoming frame into the exact format this stream's
		// codec was opened with, rather than requiring the caller to do it.
		// profile.Par() (Go-managed) is used here rather than the codec
		// parameters just opened above (C-allocated, freed further down) -
		// the resampler holds onto this pointer for its entire lifetime, so
		// it must not be one that gets freed out from under it.
		switch ff.AVMediaType(profile.Type()) {
		case ff.AVMEDIA_TYPE_AUDIO:
			r, err := newAudioResampler(profile.Par(), writer.Encoder.FrameSize(id))
			if err != nil {
				result = errors.Join(result, err)
				continue
			}
			writer.resamplers[id] = r
		case ff.AVMEDIA_TYPE_VIDEO:
			r, err := newVideoRescaler(profile.Par(), types.Value(profile.TimeBase()))
			if err != nil {
				result = errors.Join(result, err)
				continue
			}
			writer.resamplers[id] = r
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

	// Add metadata entries (but store artwork/chapters for later - artwork
	// needs its own attached-pic stream, and chapters their own array on
	// the container, so neither belongs in the metadata dict)
	var chapters []gomedia.Chapter
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

		// Collect chapter markers, written to the container below, after
		// this loop - AVFormat_new_chapters needs the final count up front
		if entry.Key() == gomedia.MetaChapter {
			chapter, ok := entry.Any().(gomedia.Chapter)
			if !ok {
				return nil, errors.Join(gomedia.ErrBadParameter.With("chapter metadata entry's Any() is not a gomedia.Chapter"), writer.Close())
			}
			chapters = append(chapters, chapter)
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

	// Write chapter markers to the container - this must happen before
	// AVFormat_write_header, for muxers (mov/mp4, mkv, ...) that write
	// chapters in the header rather than the trailer.
	if len(chapters) > 0 {
		cchapters, err := ff.AVFormat_new_chapters(writer.output, len(chapters))
		if err != nil {
			return nil, errors.Join(err, writer.Close())
		}
		for i, chapter := range chapters {
			cchapters[i].SetId(int64(i))
			// A time base of 1/time.Second lets Start/End be set directly
			// from a time.Duration's nanosecond count, with no rescaling.
			cchapters[i].SetTimeBase(ff.AVUtil_rational(1, int(time.Second)))
			cchapters[i].SetStart(int64(chapter.Start))
			cchapters[i].SetEnd(int64(chapter.End))

			if len(chapter.Metadata) == 0 {
				continue
			}
			dict := ff.AVUtil_dict_alloc()
			for key, value := range chapter.Metadata {
				if err := ff.AVUtil_dict_set(dict, key, value, ff.AV_DICT_APPEND); err != nil {
					ff.AVUtil_dict_free(dict)
					return nil, errors.Join(err, writer.Close())
				}
			}
			// Note: No defer free - ownership transferred via SetMetadata
			cchapters[i].SetMetadata(dict)
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
