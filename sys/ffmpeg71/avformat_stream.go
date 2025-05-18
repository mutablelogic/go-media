package ffmpeg

import (
	"encoding/json"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type jsonAVStream struct {
	Index       int                `json:"index"`
	Id          int                `json:"id"`
	CodecPar    *AVCodecParameters `json:"codec_par,omitempty"`
	StartTime   AVTimestamp        `json:"start_time"`
	Duration    AVTimestamp        `json:"duration"`
	NumFrames   int64              `json:"num_frames,omitempty"`
	TimeBase    AVRational         `json:"time_base,omitempty"`
	Disposition AVDisposition      `json:"disposition,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVStream) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVStream{
		Index:       int(ctx.index),
		Id:          int(ctx.id),
		CodecPar:    (*AVCodecParameters)(ctx.codecpar),
		StartTime:   AVTimestamp(ctx.start_time),
		Duration:    AVTimestamp(ctx.duration),
		NumFrames:   int64(ctx.nb_frames),
		TimeBase:    AVRational(ctx.time_base),
		Disposition: AVDisposition(ctx.disposition),
	})
}

func (ctx *AVStream) String() string {
	data, _ := json.MarshalIndent(ctx, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (ctx *AVStream) Index() int {
	return int(ctx.index)
}

func (ctx *AVStream) Id() int {
	return int(ctx.id)
}

func (ctx *AVStream) SetId(id int) {
	ctx.id = C.int(id)
}

func (ctx *AVStream) CodecPar() *AVCodecParameters {
	return (*AVCodecParameters)(ctx.codecpar)
}

func (ctx *AVStream) TimeBase() AVRational {
	return AVRational(ctx.time_base)
}

func (ctx *AVStream) SetTimeBase(time_base AVRational) {
	ctx.time_base = C.AVRational(time_base)
}

func (ctx *AVStream) Disposition() AVDisposition {
	return AVDisposition(ctx.disposition)
}

func (ctx *AVStream) AttachedPic() *AVPacket {
	if ctx.disposition&C.AV_DISPOSITION_ATTACHED_PIC == 0 {
		return nil
	} else {
		return (*AVPacket)(&ctx.attached_pic)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func AVFormat_new_stream(ctx *AVFormatContext, c *AVCodec) *AVStream {
	return (*AVStream)(C.avformat_new_stream((*C.struct_AVFormatContext)(ctx), (*C.struct_AVCodec)(c)))
}

// Find the best stream given the media type, wanted stream number, and related stream number.
func AVFormat_find_best_stream(ctx *AVFormatContext, t AVMediaType, wanted int, related int) (int, *AVCodec, error) {
	var codec *C.struct_AVCodec
	ret := int(C.av_find_best_stream((*C.struct_AVFormatContext)(ctx), (C.enum_AVMediaType)(t), C.int(wanted), C.int(related), (**C.struct_AVCodec)(&codec), 0))
	if ret < 0 {
		return 0, nil, AVError(ret)
	} else {
		return ret, (*AVCodec)(codec), nil
	}
}
