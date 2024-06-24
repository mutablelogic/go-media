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

type writerstream struct {
	t     MediaType
	codec *ff.AVCodec
}

var _ Stream = (*stream)(nil)

//var _ Stream = (*writerstream)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Stream wrapper for decoding
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
	switch stream.Type() {
	case AUDIO:
		return newCodecAudioParameters(stream.CodecPar())
	case VIDEO:
		return newCodecVideoParameters(stream.CodecPar())
	}

	// Other types not yet supported
	return nil
}
