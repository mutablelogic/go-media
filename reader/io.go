package reader

import (
	"io"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type io_callback struct {
	r io.Reader
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	ioBufferSize = 4096
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - Reader

func (r *io_callback) Reader(buf []byte) int {
	n, err := r.r.Read(buf)
	if n > 0 {
		return n
	}
	if err != nil {
		return ff.AVERROR_EOF
	}
	return n
}

func (r *io_callback) Seeker(offset int64, whence int) int64 {
	whence = whence & ^ff.AVSEEK_FORCE
	seeker, ok := r.r.(io.ReadSeeker)
	if !ok {
		return -1
	}
	if whence == ff.AVSEEK_SIZE {
		current, err := seeker.Seek(0, io.SeekCurrent)
		if err != nil {
			return -1
		}
		size, err := seeker.Seek(0, io.SeekEnd)
		if err != nil {
			return -1
		}
		if _, err := seeker.Seek(current, io.SeekStart); err != nil {
			return -1
		}
		return size
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

func (r *io_callback) Writer([]byte) int {
	return ff.AVERROR_EOF
}
