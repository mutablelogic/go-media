package media

import (
	"encoding/json"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type codec struct {
	ctx *ff.AVCodec
}

var _ Codec = (*codec)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newCodec(ctx *ff.AVCodec) *codec {
	return &codec{ctx}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

type jsonCodec struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
}

func (codec *codec) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonCodec{
		Name:        codec.Name(),
		Description: codec.Description(),
		Type:        codec.Type().String(),
	})
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (codec *codec) Name() string {
	return codec.ctx.Name()
}

// Return the codec description
func (codec *codec) Description() string {
	return codec.ctx.LongName()
}

// Return the codec type (AUDIO, VIDEO, SUBTITLE, DATA, INPUT, OUTPUT)
func (codec *codec) Type() MediaType {
	t := NONE
	switch codec.ctx.Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		t = AUDIO
	case ff.AVMEDIA_TYPE_VIDEO:
		t = VIDEO
	case ff.AVMEDIA_TYPE_SUBTITLE:
		t = SUBTITLE
	default:
		t = DATA
	}
	if ff.AVCodec_is_encoder(codec.ctx) {
		t |= OUTPUT
	}
	if ff.AVCodec_is_decoder(codec.ctx) {
		t |= INPUT
	}
	return t
}

// TODO: Supported sample formats, channel layouts, pixel formats, etc.
