package ffmpeg

import (
	"errors"
	"fmt"
	"io"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// Create media from io.Writer
type Writer struct {
	output *ff.AVFormatContext
}

type writer_callback struct {
	w io.Writer
}

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	bufSize = 4096
)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewWriter(w io.Writer, opt ...Opt) (*Writer, error) {
	var o opts

	writer := new(Writer)

	// Apply options
	for _, opt := range opt {
		if err := opt(&o); err != nil {
			return nil, err
		}
	}

	// Check output
	if o.oformat == nil {
		return nil, ErrBadParameter.Withf("invalid output format")
	}

	// Allocate the AVIO context
	avio := ff.AVFormat_avio_alloc_context(bufSize, false, &writer_callback{w})
	if avio == nil {
		return nil, errors.New("failed to allocate avio context")
	} else if ctx, err := ff.AVFormat_open_writer(avio, o.oformat, ""); err != nil {
		return nil, err
	} else {
		writer.output = ctx
	}

	fmt.Println("WRITER", writer.output)

	// Return success
	return writer, nil
}

func (w *Writer) Close() error {
	var result error

	// Free output resources
	if w.output != nil {
		// This calls avio_close(w.avio)
		fmt.Println("TODO: AVFormat_close_writer")
		result = errors.Join(result, ff.AVFormat_close_writer(w.output))
		w.output = nil
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (w *writer_callback) Reader(buf []byte) int {
	return 0
}

func (w *writer_callback) Seeker(offset int64, whence int) int64 {
	whence = whence & ^ff.AVSEEK_FORCE
	seeker, ok := w.w.(io.ReadSeeker)
	if !ok {
		return -1
	}
	switch whence {
	case io.SeekStart, io.SeekCurrent, io.SeekEnd:
		n, err := seeker.Seek(offset, whence)
		if err != nil {
			return -1
		}
		return n
	}
	return -1
}

func (w *writer_callback) Writer(buf []byte) int {
	if n, err := w.w.Write(buf); err != nil {
		return -1
	} else {
		return n
	}
}
