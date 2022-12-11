package media

import (

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	//. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type codec struct {
	ctx *ffmpeg.AVCodecContext
}

// Ensure *stream complies with Stream interface
var _ Codec = (*codec)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewCodec(c *ffmpeg.AVCodec) *codec {
	if c == nil {
		return nil
	}
	ctx := ffmpeg.AVCodec_alloc_context3(c)
	if ctx == nil {
		return nil
	}
	return &codec{
		ctx: ctx,
	}
}

func (codec *codec) Close() error {
	if codec.ctx != nil {
		ffmpeg.AVCodec_free_context(&codec.ctx)
		codec.ctx = nil
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (codec *codec) String() string {
	str := "<media.codec"
	return str + ">"
}

// //////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS
// Name returns the unique name for the codec
func (codec *codec) Name() string {
	if codec.ctx == nil {
		return ""
	}
	return codec.ctx.Codec().Name()
}

// Description returns the long description for the codec
func (codec *codec) Description() string {
	if codec.ctx == nil {
		return ""
	}
	return codec.ctx.Codec().Description()
}

// Flags for the codec (Audio, Video, Encoder, Decoder)
func (codec *codec) Flags() MediaFlag {
	if codec.ctx == nil {
		return MEDIA_FLAG_NONE
	}
	return MEDIA_FLAG_NONE
}
