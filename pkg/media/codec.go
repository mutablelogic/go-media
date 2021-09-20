package media

import (
	"fmt"

	// Packages
	ffmpeg "github.com/djthorpe/go-media/sys/ffmpeg"

	// Namespace imports
	. "github.com/djthorpe/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Codec struct {
	ctx   *ffmpeg.AVCodecParameters
	codec *ffmpeg.AVCodec
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewCodec(ctx *ffmpeg.AVCodec) *Codec {
	if ctx == nil {
		return nil
	}
	return &Codec{nil, ctx}
}

func NewCodecWithParameters(ctx *ffmpeg.AVCodecParameters) *Codec {
	if ctx == nil {
		return nil
	}
	return &Codec{ctx, ffmpeg.FindCodecById(ctx.Id())}
}

func (c *Codec) Release() error {
	c.ctx = nil
	c.codec = nil
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c *Codec) String() string {
	str := "<codec"
	if name := c.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if description := c.Description(); description != "" {
		str += fmt.Sprintf(" desc=%q", description)
	}
	if flags := c.Flags(); flags != MEDIA_FLAG_NONE {
		str += fmt.Sprint(" flags=", flags)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (c *Codec) Name() string {
	if c.codec == nil {
		return ""
	}
	return c.codec.Name()
}

func (c *Codec) Description() string {
	if c.codec == nil {
		return ""
	}
	return c.codec.Description()
}

func (c *Codec) Flags() MediaFlag {
	flags := MEDIA_FLAG_NONE

	switch {
	case c.ctx != nil:
		switch c.ctx.Type() {
		case ffmpeg.AVMEDIA_TYPE_VIDEO:
			if c.ctx.BitRate() > 0 {
				flags |= MEDIA_FLAG_VIDEO
			}
		case ffmpeg.AVMEDIA_TYPE_AUDIO:
			flags |= MEDIA_FLAG_AUDIO
		case ffmpeg.AVMEDIA_TYPE_SUBTITLE:
			flags |= MEDIA_FLAG_SUBTITLE
		case ffmpeg.AVMEDIA_TYPE_UNKNOWN, ffmpeg.AVMEDIA_TYPE_DATA:
			flags |= MEDIA_FLAG_DATA
		case ffmpeg.AVMEDIA_TYPE_ATTACHMENT:
			flags |= MEDIA_FLAG_ATTACHMENT
		}
	case c.codec != nil:
		switch c.codec.Type() {
		case ffmpeg.AVMEDIA_TYPE_VIDEO:
			flags |= MEDIA_FLAG_VIDEO
		case ffmpeg.AVMEDIA_TYPE_AUDIO:
			flags |= MEDIA_FLAG_AUDIO
		case ffmpeg.AVMEDIA_TYPE_SUBTITLE:
			flags |= MEDIA_FLAG_SUBTITLE
		case ffmpeg.AVMEDIA_TYPE_UNKNOWN, ffmpeg.AVMEDIA_TYPE_DATA:
			flags |= MEDIA_FLAG_DATA
		case ffmpeg.AVMEDIA_TYPE_ATTACHMENT:
			flags |= MEDIA_FLAG_ATTACHMENT
		}
	}

	// Encode and decode flags
	if c.codec.IsEncoder() {
		flags |= MEDIA_FLAG_ENCODER
	}
	if c.codec.IsDecoder() {
		flags |= MEDIA_FLAG_DECODER
	}

	// Return flags
	return flags
}
