package media

import (
	"fmt"

	// Packages
	ffmpeg "github.com/djthorpe/go-media/sys/ffmpeg"
	multierror "github.com/hashicorp/go-multierror"

	// Namespace imports
	. "github.com/djthorpe/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Stream struct {
	ctx   *ffmpeg.AVStream
	codec *Codec
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewStream(ctx *ffmpeg.AVStream, source *Stream) *Stream {
	this := new(Stream)
	this.ctx = ctx
	if source == nil {
		this.codec = NewCodecWithParameters(ctx.CodecPar())
	} else {
		this.codec = NewCodecWithParameters(source.ctx.CodecPar())
	}

	// Return success
	return this
}

func (s *Stream) Release() error {
	var result error
	if s.codec != nil {
		if err := s.codec.Release(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Set instance variables to nil
	s.ctx = nil
	s.codec = nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s *Stream) String() string {
	str := "<stream"
	if i := s.Index(); i >= 0 {
		str += fmt.Sprint(" index=", i)
	}
	if codec := s.Codec(); codec != nil {
		str += fmt.Sprint(" codec=", codec)
	}
	if flags := s.Flags(); flags != MEDIA_FLAG_NONE {
		str += fmt.Sprint(" flags=", flags)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

func (s *Stream) Index() int {
	if s.ctx == nil {
		return -1
	}
	return s.ctx.Index()
}

func (s *Stream) Codec() *Codec {
	if s.ctx == nil {
		return nil
	}
	return s.codec
}

func (s *Stream) Artwork() []byte {
	if s.ctx == nil {
		return nil
	}
	if s.ctx.Disposition()&ffmpeg.AV_DISPOSITION_ATTACHED_PIC == 0 {
		return nil
	}
	if pkt := s.ctx.AttachedPicture(); pkt == nil {
		return nil
	} else {
		return pkt.Bytes()
	}
}

func (s *Stream) Flags() MediaFlag {
	flags := MEDIA_FLAG_NONE
	if s.ctx == nil {
		return flags
	}

	// Codec flags
	if s.codec != nil {
		flags |= s.codec.Flags()
	}

	// Remove encoder/decoder flags
	flags &^= (MEDIA_FLAG_ENCODER | MEDIA_FLAG_DECODER)

	// Disposition flags
	if s.ctx.Disposition()&ffmpeg.AV_DISPOSITION_ATTACHED_PIC != 0 {
		flags |= MEDIA_FLAG_ARTWORK
	}
	if s.ctx.Disposition()&ffmpeg.AV_DISPOSITION_CAPTIONS != 0 {
		flags |= MEDIA_FLAG_CAPTIONS
	}

	// Return flags
	return flags
}
