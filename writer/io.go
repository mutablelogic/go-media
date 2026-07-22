package writer

import (
	"io"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type io_callback struct {
	w io.Writer
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	ioBufferSize = 4096
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - Writer

func (w *io_callback) Reader(buf []byte) int {
	if r, ok := w.w.(io.Reader); ok {
		if n, err := r.Read(buf); err != nil {
			return -1
		} else {
			return n
		}
	}
	return -1
}

func (w *io_callback) Seeker(offset int64, whence int) int64 {
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

func (w *io_callback) Writer(buf []byte) int {
	if n, err := w.w.Write(buf); err != nil {
		return -1
	} else {
		return n
	}
}
