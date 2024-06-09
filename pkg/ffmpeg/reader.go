package ffmpeg

import (
	"io"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Reader struct {
	r   io.Reader
	ctx *ffmpeg.AVIOContextEx
}

var _ io.ReadCloser = (*Reader)(nil)
var _ io.Seeker = (*Reader)(nil)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	bufSize = 1024 * 64
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewReader creates a new reader from an io.Reader
// You should call Close() on the reader when you are done with it
// The reader will not close the underlying io.Reader
func NewReader(r io.Reader) *Reader {
	reader := new(Reader)
	reader.r = r
	reader.ctx = ffmpeg.AVFormat_avio_alloc_context(bufSize, false, reader)
	if reader.ctx == nil {
		return nil
	}
	return reader
}

// Close closes the reader
func (r *Reader) Close() error {
	ffmpeg.AVFormat_avio_context_free(r.ctx)
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// READ

// Read reads data from the reader
func (r *Reader) Read(buf []byte) (int, error) {
	// Read data
	n := ffmpeg.AVFormat_avio_read(r.ctx, buf)
	if n == ffmpeg.AVERROR_EOF {
		return 0, io.EOF
	} else if n < 0 {
		return 0, io.EOF
	}
	return n, nil
}

// Seek seeks to a position in the reader
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	n := ffmpeg.AVFormat_avio_seek(r.ctx, offset, whence)
	if n < 0 {
		return 0, io.EOF
	}
	return n, nil
}

////////////////////////////////////////////////////////////////////////////////
// CALLBACKS

func (r *Reader) Reader(buf []byte) int {
	n, err := r.r.Read(buf)
	if err != nil {
		return ffmpeg.AVERROR_EOF
	}
	return n
}

func (r *Reader) Seeker(offset int64, whence int) int64 {
	if _, ok := r.r.(io.ReadSeeker); ok {
		n, err := r.r.(io.ReadSeeker).Seek(offset, whence)
		if err != nil {
			return -1
		}
		return n
	}
	return -1
}

func (r *Reader) Writer([]byte) int {
	return ffmpeg.AVERROR_EOF
}
