package mmal

import (
	"fmt"

	"github.com/djthorpe/mmal"
)

func (this *streamformat) Type() mmal.StreamType {
	return mmal.StreamType(this.handle._type)
}

func (this *streamformat) Encoding() mmal.EncodingType {
	return mmal.EncodingType(this.handle.encoding)
}

func (this *streamformat) EncodingVariant() mmal.EncodingType {
	return mmal.EncodingType(this.handle.encoding_variant)
}

func (this *streamformat) Bitrate() uint32 {
	return uint32(this.handle.bitrate)
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *streamformat) String() string {
	if this.Type() == mmal.MMAL_ES_TYPE_NONE || this.Type() == mmal.MMAL_ES_TYPE_CONTROL {
		return fmt.Sprintf("<mmal.StreamFormat>{ type=%v }", this.Type())
	} else {
		return fmt.Sprintf("<mmal.StreamFormat>{ type=%v encoding=%v variant=%v bitrate=%vbps }", this.Type(), this.Encoding(), this.EncodingVariant(), this.Bitrate())
	}
}
