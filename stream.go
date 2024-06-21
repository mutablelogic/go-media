package media

import (
	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type stream struct {
	*ff.AVStream
}

var _ Stream = (*stream)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Open media from a url, file path or device
func newStream(ctx *ff.AVStream) *stream {
	return &stream{ctx}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (stream *stream) Type() MediaType {
	switch stream.CodecPar().CodecType() {
	case ff.AVMEDIA_TYPE_AUDIO:
		return AUDIO
	case ff.AVMEDIA_TYPE_VIDEO:
		return VIDEO
	case ff.AVMEDIA_TYPE_SUBTITLE:
		return SUBTITLE
	}
	return DATA
}

func (stream *stream) Parameters() Parameters {
	// TODO
	return nil
}
