package ffmpeg

import (
	"encoding/json"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Stream struct {
	ctx *ff.AVStream
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new stream
func newStream(ctx *ff.AVStream) *Stream {
	return &Stream{
		ctx: ctx,
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s *Stream) MarshalJSON() ([]byte, error) {
	type j struct {
		Index int                   `json:"index"`
		Type  media.Type            `json:"type"`
		Codec *ff.AVCodecParameters `json:"codec,omitempty"`
	}
	return json.Marshal(j{
		Index: s.Index(),
		Type:  s.Type(),
		Codec: s.ctx.CodecPar(),
	})
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

// Return the stream index
func (s *Stream) Index() int {
	return int(s.ctx.Index())
}

// Return the stream type
func (s *Stream) Type() media.Type {
	if s.ctx.Disposition()&ff.AV_DISPOSITION_ATTACHED_PIC != 0 {
		return media.DATA
	}
	switch s.ctx.CodecPar().CodecType() {
	case ff.AVMEDIA_TYPE_VIDEO:
		return media.VIDEO
	case ff.AVMEDIA_TYPE_AUDIO:
		return media.AUDIO
	case ff.AVMEDIA_TYPE_DATA:
		return media.DATA
	case ff.AVMEDIA_TYPE_SUBTITLE:
		return media.SUBTITLE
	default:
		return media.UNKNOWN
	}
}

// Return the codec parameters
func (s *Stream) CodecPar() *ff.AVCodecParameters {
	return s.ctx.CodecPar()
}
