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
	*ff.AVStream
}

var _ Stream = (*stream)(nil)

//var _ Stream = (*writerstream)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Stream wrapper for decoding
func newStream(ctx *ff.AVStream) *stream {
	return &stream{ctx}
}

/*
// Stream wrapper for encoding
func newWriterStream(ctx *ff.AVFormatContext, param Parameters) (*writerstream, error) {
	// Parameters - Codec
	var codec_id ff.AVCodecID
	if param.Type().Is(CODEC) {
		codec_id = param.Codec().ID()
	} else if param.Type().Is(VIDEO) {
		codec_id = ctx.Input().VideoCodec()
	} else if param.Type().Is(AUDIO) {
		codec_id = ctx.Input().AudioCodec()
	} else {
		return nil, ErrBadParameter.With("invalid stream parameters")

	}

	return nil, ErrNotImplemented
}
*/
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
