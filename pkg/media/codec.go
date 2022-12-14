package media

import (
	"fmt"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	//. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type codec struct {
	ctx *ffmpeg.AVCodec
}

// Ensure *stream complies with Stream interface
var _ Codec = (*codec)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a codec from a AVCodecID for a stream
func NewCodecEncoder(id ffmpeg.AVCodecID) *codec {
	this := new(codec)
	if id == ffmpeg.AV_CODEC_ID_NONE {
		return nil
	}

	// Find the decoder
	encoder := ffmpeg.AVCodec_find_encoder(id)
	if encoder == nil {
		return nil
	} else {
		this.ctx = encoder
	}

	// Return success
	return this
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (codec *codec) String() string {
	str := "<media.codec"
	if name := codec.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if desc := codec.Description(); desc != "" {
		str += fmt.Sprintf(" desc=%q", desc)
	}
	if flags := codec.Flags(); flags != MEDIA_FLAG_NONE {
		str += fmt.Sprint(" flags=", flags)
	}
	return str + ">"
}

// //////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS
// Name returns the unique name for the codec
func (codec *codec) Name() string {
	if codec.ctx == nil {
		return ""
	}
	return codec.ctx.Name()
}

// Description returns the long description for the codec
func (codec *codec) Description() string {
	if codec.ctx == nil {
		return ""
	}
	return codec.ctx.Description()
}

// Flags for the codec (Audio, Video, Encoder, Decoder, ...)
func (codec *codec) Flags() MediaFlag {
	flags := MEDIA_FLAG_NONE
	if codec.ctx == nil {
		return flags
	}
	if codec.ctx.AVCodec_is_decoder() {
		flags |= MEDIA_FLAG_DECODER
	}
	if codec.ctx.AVCodec_is_encoder() {
		flags |= MEDIA_FLAG_ENCODER
	}
	switch codec.ctx.MediaType() {
	case ffmpeg.AVMEDIA_TYPE_AUDIO:
		flags |= MEDIA_FLAG_AUDIO
	case ffmpeg.AVMEDIA_TYPE_VIDEO:
		flags |= MEDIA_FLAG_VIDEO
	case ffmpeg.AVMEDIA_TYPE_SUBTITLE:
		flags |= MEDIA_FLAG_SUBTITLE
	case ffmpeg.AVMEDIA_TYPE_DATA:
		flags |= MEDIA_FLAG_DATA
	case ffmpeg.AVMEDIA_TYPE_ATTACHMENT:
		flags |= MEDIA_FLAG_ATTACHMENT
	}
	return flags
}
