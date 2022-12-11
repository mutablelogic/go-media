package media

import (

	// Packages
	"fmt"

	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	//. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type codec struct {
	ctx   *ffmpeg.AVCodecContext
	codec *ffmpeg.AVCodec
}

// Ensure *stream complies with Stream interface
var _ Codec = (*codec)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a decoder for a stream
func NewDecoder(stream *ffmpeg.AVStream) *codec {
	this := new(codec)
	if stream == nil {
		return nil
	}

	// Find the decoder
	decoder := ffmpeg.AVCodec_find_decoder(stream.CodecPar().CodecID())
	if decoder == nil {
		return nil
	} else {
		this.codec = decoder
	}

	// Allocate a codec context for the decoder
	ctx := ffmpeg.AVCodec_alloc_context3(decoder)
	if ctx == nil {
		return nil
	} else {
		this.ctx = ctx
	}

	// Copy codec parameters from input stream to output codec context#
	if err := ffmpeg.AVCodec_parameters_to_context(this.ctx, stream.CodecPar()); err != nil {
		ffmpeg.AVCodec_free_context_ptr(this.ctx)
		return nil
	}

	// Init the decoders
	if err := ffmpeg.AVCodec_open2(this.ctx, decoder, nil); err != nil {
		ffmpeg.AVCodec_free_context_ptr(this.ctx)
		return nil
	}

	// Return success
	return this
}

func (codec *codec) Close() error {
	if codec.ctx != nil {
		ffmpeg.AVCodec_free_context_ptr(codec.ctx)
		codec.ctx = nil
	}
	return nil
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
	return codec.ctx.Codec().Name()
}

// Description returns the long description for the codec
func (codec *codec) Description() string {
	if codec.ctx == nil {
		return ""
	}
	return codec.ctx.Codec().Description()
}

// Flags for the codec (Audio, Video, Encoder, Decoder, ...)
func (codec *codec) Flags() MediaFlag {
	flags := MEDIA_FLAG_NONE
	if codec.codec == nil {
		return flags
	}
	if codec.codec.AVCodec_is_decoder() {
		flags |= MEDIA_FLAG_DECODER
	}
	if codec.codec.AVCodec_is_encoder() {
		flags |= MEDIA_FLAG_ENCODER
	}
	switch codec.codec.MediaType() {
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
