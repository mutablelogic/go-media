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

type stream struct {
	ctx *ffmpeg.AVStream
}

// Ensure *stream complies with Stream interface
var _ Stream = (*stream)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewStream(ctx *ffmpeg.AVStream) *stream {
	if ctx == nil {
		return nil
	}
	return &stream{
		ctx: ctx,
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (stream *stream) String() string {
	str := "<media.stream"
	str += fmt.Sprint(" index=", stream.Index())
	if flags := stream.Flags(); flags != MEDIA_FLAG_NONE {
		str += fmt.Sprint(" flags=", flags)
	}
	if stream.ctx != nil {
		str += fmt.Sprint(" ctx=", stream.ctx)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (stream *stream) Index() int {
	if stream.ctx == nil {
		return -1
	}
	return stream.ctx.Index()
}

func (stream *stream) Flags() MediaFlag {
	flags := MEDIA_FLAG_NONE
	if stream.ctx == nil {
		return flags
	}

	// TODO: Add codec flags
	//if stream.ctx.CodecPar()codec != nil {
	//	flags |= s.codec.Flags()
	//}

	// Remove encoder/decoder flags
	flags &^= (MEDIA_FLAG_ENCODER | MEDIA_FLAG_DECODER)

	// Disposition flags
	if stream.ctx.Disposition()&ffmpeg.AV_DISPOSITION_ATTACHED_PIC != 0 {
		flags |= MEDIA_FLAG_ARTWORK
	}
	if stream.ctx.Disposition()&ffmpeg.AV_DISPOSITION_CAPTIONS != 0 {
		flags |= MEDIA_FLAG_CAPTIONS
	}

	// Return flags
	return flags
}

func (stream *stream) Artwork() []byte {
	if stream.ctx == nil {
		return nil
	}
	if stream.ctx.Disposition()&ffmpeg.AV_DISPOSITION_ATTACHED_PIC == 0 {
		return nil
	}
	if pkt := stream.ctx.AttachedPic(); pkt.Size() == 0 {
		return nil
	} else {
		return pkt.Bytes()
	}
}

/*
func (s *Stream) Codec() *Codec {
	if s.ctx == nil {
		return nil
	}
	return s.codec
}
*/
