package writer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sync"

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
	*Encoder
	output *ff.AVFormatContext
	header bool // Track if header was successfully written (for Close)
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new writer with a URL and options
func Create(url *url.URL, output *profile.Output, streams ...profile.Profile) (*Writer, error) {
	self := new(Writer)
	encoder, err := NewEncoder(self.writePacket)
	if err != nil {
		return nil, err
	} else {
		self.Encoder = encoder
	}

	// Check URL and output
	if url == nil || output == nil || output.Context() == nil {
		return nil, gomedia.ErrBadParameter.Withf("url and output must be non-nil")
	} else if len(streams) == 0 {
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
	return self.open(output, streams...)
}

// Create a new writer with an io.Writer and options
func NewWriter(w io.Writer, output *profile.Output, streams ...profile.Profile) (*Writer, error) {
	self := new(Writer)
	encoder, err := NewEncoder(self.writePacket)
	if err != nil {
		return nil, err
	} else {
		self.Encoder = encoder
	}

	// Check writer and output
	if w == nil || output == nil || output.Context() == nil {
		return nil, gomedia.ErrBadParameter.Withf("writer and output must be non-nil")
	} else if len(streams) == 0 {
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
	return self.open(output, streams...)
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

	// Return any errors
	return result
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (writer *Writer) open(output *profile.Output, streams ...profile.Profile) (*Writer, error) {
	var result error

	// Create streams
	for i, profile := range streams {
		// Create stream
		stream := ff.AVFormat_new_stream(writer.output, profile.Codec().Context())
		if stream == nil {
			result = errors.Join(result, gomedia.ErrInternalError.Withf("failed to allocate stream for profile %q", profile.UUID()))
			continue
		} else {
			// Set stream index 0...n-1
			stream.SetId(i)
		}

		// Open a codec context for this stream so frames can actually be
		// encoded. This must happen before copying codec parameters below:
		// opening the codec is what generates extradata some codecs need
		// (e.g. OpusHead for libopus, AudioSpecificConfig for aac), and the
		// muxer needs that extradata on the stream, not just the raw
		// parameters the profile requested.
		if err := writer.Encoder.Add(i, profile); err != nil {
			result = errors.Join(result, err)
			continue
		}

		// Copy codec parameters from the now-opened codec context
		par, err := writer.Encoder.Par(i)
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

		// Set timebase if specified
		if timebase := profile.TimeBase(); timebase != nil {
			stream.SetTimeBase(types.Value(timebase))
		}
	}

	// Bail out if any stream failed to be created
	if result != nil {
		return writer, result
	}

	// Build the format-level options dictionary from the output profile
	dict, err := dictFromOpts(output.Opts)
	if err != nil {
		return writer, err
	}
	defer ff.AVUtil_dict_free(dict)

	// Write the header, consuming recognized options from the dictionary
	if err := ff.AVFormat_write_header(writer.output, dict); err != nil {
		return writer, err
	}
	writer.header = true

	// Any keys left in the dictionary were not recognized by the muxer
	if keys := ff.AVUtil_dict_keys(dict); len(keys) > 0 {
		return writer, gomedia.ErrBadParameter.Withf("invalid output options: %v", keys)
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
	if packet == nil {
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
