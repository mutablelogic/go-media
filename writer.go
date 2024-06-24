package media

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type writer struct {
	t        MediaType
	output   *ff.AVFormatContext
	avio     *ff.AVIOContextEx
	metadata *ff.AVDictionary
	header   bool
	encoder  map[int]*encoder
}

type writer_callback struct {
	w io.Writer
}

var _ Media = (*writer)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create media from a url or device
func createMedia(url string, format Format, metadata []Metadata, params ...Parameters) (*writer, error) {
	writer := new(writer)
	writer.t = OUTPUT
	writer.encoder = make(map[int]*encoder, len(params))

	// If there are no streams, then return an error
	if len(params) == 0 {
		return nil, ErrBadParameter.With("no streams specified for encoder")
	}

	// Guess the output format
	var ofmt *ff.AVOutputFormat
	if format == nil && url != "" {
		ofmt = ff.AVFormat_guess_format("", url, "")
	} else if format != nil {
		ofmt = format.(*outputformat).ctx
	}
	if ofmt == nil {
		return nil, ErrBadParameter.With("unable to guess the output format")
	}

	// Allocate the output media context
	ctx, err := ff.AVFormat_create_file(url, ofmt)
	if err != nil {
		return nil, err
	} else {
		writer.output = ctx
	}

	// Add encoders and streams
	var result error
	for i, param := range params {
		encoder, err := newEncoder(ctx, i, param)
		if err != nil {
			result = errors.Join(result, err)
		} else {
			writer.encoder[i] = encoder
		}
	}

	// Return any errors from creating the streams
	if result != nil {
		return nil, errors.Join(result, writer.Close())
	}

	// Open the output file, if needed
	if !ctx.Flags().Is(ff.AVFMT_NOFILE) {
		w, err := ff.AVFormat_avio_open(url, ff.AVIO_FLAG_WRITE)
		if err != nil {
			return nil, errors.Join(err, writer.Close())
		} else {
			ctx.SetPb(w)
			writer.avio = w
		}
	}

	// Set metadata
	if len(metadata) > 0 {
		writer.metadata = ff.AVUtil_dict_alloc()
		if writer.metadata == nil {
			return nil, errors.Join(errors.New("unable to allocate metadata dictionary"), writer.Close())
		}
		for _, m := range metadata {
			// Ignore duration and artwork fields
			key := m.Key()
			if key == MetaArtwork || key == MetaDuration {
				continue
			}
			// Set dictionary entry
			if err := ff.AVUtil_dict_set(writer.metadata, key, fmt.Sprint(m.Value()), ff.AV_DICT_APPEND); err != nil {
				return nil, errors.Join(err, writer.Close())
			}
		}
		// TODO: Create artwork streams
	}

	// Write the header
	if err := ff.AVFormat_write_header(ctx, nil); err != nil {
		return nil, errors.Join(err, writer.Close())
	} else {
		writer.header = true
	}

	// Return success
	return writer, nil
}

// Create media from io.Writer
// TODO
func createWriter(w io.Writer, format Format, metadata []Metadata, params ...Parameters) (*writer, error) {
	return nil, ErrNotImplemented
}

func (w *writer) Close() error {
	var result error

	// Write the trailer if the header was written
	if w.header {
		if err := ff.AVFormat_write_trailer(w.output); err != nil {
			result = errors.Join(result, err)
		}
	}

	// Close encoders
	for _, encoder := range w.encoder {
		result = errors.Join(result, encoder.Close())
	}

	// Free resources
	if w.metadata != nil {
		ff.AVUtil_dict_free(w.metadata)
	}
	if w.output != nil {
		result = errors.Join(result, ff.AVFormat_close_writer(w.output))
	}
	if w.avio != nil {
		fmt.Println("TODO AVIO")
		//		result = errors.Join(result, ff.AVFormat_avio_close(w.avio))
	}

	// Release resources
	w.encoder = nil
	w.metadata = nil
	w.avio = nil
	w.output = nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// Display the reader as a string
func (w *writer) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.output)
}

// Display the reader as a string
func (w *writer) String() string {
	data, _ := json.MarshalIndent(w, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (w *writer) Decoder(DecoderMapFunc) (Decoder, error) {
	return nil, ErrNotImplemented
}

// Return OUTPUT and combination of DEVICE and STREAM
func (w *writer) Type() MediaType {
	return OUTPUT
}

// Return the metadata for the media.
func (w *writer) Metadata(...string) []Metadata {
	// Not yet implemented
	return nil
}
