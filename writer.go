package media

import (
	"errors"
	"io"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type writer struct {
	t      MediaType
	output *ff.AVFormatContext
	avio   *ff.AVIOContextEx
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

	// Add streams
	/*
		for _, param := range params {
			stream, err := newWriterStream(ctx, param)
			if err != nil {
				return nil, errors.Join(err, writer.Close())
			} else {
				fmt.Println("TODO: STREAM", stream)
			}
		}
	*/
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

	// TODO: Metadata

	// TODO: Write the header

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

	// TODO: Write the trailer

	// Free resources
	if w.avio != nil {
		result = errors.Join(result, ff.AVFormat_avio_close(w.avio))
	}
	result = errors.Join(result, ff.AVFormat_close_writer(w.output))

	// Return any errors
	return result
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
