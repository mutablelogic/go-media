package main

import (
	// Packages

	"errors"
	"fmt"
	"io"
	"log"
	"mime"

	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type reader struct {
	*ff.AVFormatContext
	ctx *ffmpeg.AVIOContextEx
}

type reader_callback struct {
	r io.Reader
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	bufSize = 4096
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewReader(r io.Reader, mimetype string) (*reader, error) {
	reader := new(reader)

	// Get input format from mimetype
	if mimetype != "" {
		if t, _, err := mime.ParseMediaType(mimetype); err != nil {
			return nil, err
		} else {
			mimetype = t
		}
	}

	// If there is a mimetype, then use it to find the input format
	var opaque uintptr
	for {
		format := ff.AVFormat_demuxer_iterate(&opaque)
		if format == nil {
			break
		}
		log.Println(format)
	}

	// Allocate the AVIO context
	reader.ctx = ffmpeg.AVFormat_avio_alloc_context(bufSize, false, &reader_callback{r})
	if reader.ctx == nil {
		return nil, errors.New("failed to allocate avio context")
	}

	// Open the stream
	if ctx, err := ff.AVFormat_open_reader(reader.ctx, nil, nil); err != nil {
		ffmpeg.AVFormat_avio_context_free(reader.ctx)
		fmt.Println("TODO: AVFormat_open_reader", err)
		return nil, err
	} else {
		reader.AVFormatContext = ctx
	}

	// Find stream information
	if err := ff.AVFormat_find_stream_info(reader.AVFormatContext, nil); err != nil {
		ff.AVFormat_free_context(reader.AVFormatContext)
		ffmpeg.AVFormat_avio_context_free(reader.ctx)
		fmt.Println("TODO: AVFormat_open_reader", err)
		return nil, err
	}

	// Return success
	return reader, nil
}

func (r *reader) Close() {
	ff.AVFormat_free_context(r.AVFormatContext)
	ffmpeg.AVFormat_avio_context_free(r.ctx)
	r.AVFormatContext = nil
	r.ctx = nil
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

// Read reads data from the reader
func (r *reader) Read(buf []byte) (int, error) {
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
func (r *reader) Seek(offset int64, whence int) (int64, error) {
	n := ffmpeg.AVFormat_avio_seek(r.ctx, offset, whence)
	if n < 0 {
		return 0, io.EOF
	}
	return n, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (r *reader_callback) Reader(buf []byte) int {
	n, err := r.r.Read(buf)
	if err != nil {
		return ffmpeg.AVERROR_EOF
	}
	return n
}

func (r *reader_callback) Seeker(offset int64, whence int) int64 {
	whence = whence & ^ff.AVSEEK_FORCE
	seeker, ok := r.r.(io.ReadSeeker)
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

func (r *reader_callback) Writer([]byte) int {
	return ffmpeg.AVERROR_EOF
}
