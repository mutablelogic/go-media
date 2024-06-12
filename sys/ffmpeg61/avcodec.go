package ffmpeg

import (
	"encoding/json"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVPacket          C.AVPacket
	AVCodec           C.AVCodec
	AVCodecContext    C.AVCodecContext
	AVCodecParameters C.AVCodecParameters
	AVCodecID         C.enum_AVCodecID
)

type jsonAVPacket struct {
	Pts           int64 `json:"pts,omitempty"`
	Dts           int64 `json:"dts,omitempty"`
	Size          int   `json:"size,omitempty"`
	StreamIndex   int   `json:"stream_index"` // Stream index starts at 0
	Flags         int   `json:"flags,omitempty"`
	SideDataElems int   `json:"side_data_elems,omitempty"`
	Duration      int64 `json:"duration,omitempty"`
	Pos           int64 `json:"pos,omitempty"`
}

type jsonAVCodecParameters struct {
	CodecType AVMediaType `json:"codec_type,omitempty"`
	CodecID   AVCodecID   `json:"codec_id,omitempty"`
	CodecTag  uint32      `json:"codec_tag,omitempty"`
	Format    int         `json:"format,omitempty"`
	BitRate   int64       `json:"bit_rate,omitempty"`
}

type jsonAVCodec struct {
	Name         string      `json:"name,omitempty"`
	LongName     string      `json:"long_name,omitempty"`
	Type         AVMediaType `json:"type,omitempty"`
	ID           AVCodecID   `json:"id,omitempty"`
	Capabilities int         `json:"capabilities,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_CODEC_ID_NONE AVCodecID = C.AV_CODEC_ID_NONE
	AV_CODEC_ID_MP2  AVCodecID = C.AV_CODEC_ID_MP2
)

////////////////////////////////////////////////////////////////////////////////
// JSON OUTPUT

func (ctx *AVPacket) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVPacket{
		Pts:           int64(ctx.pts),
		Dts:           int64(ctx.dts),
		Size:          int(ctx.size),
		StreamIndex:   int(ctx.stream_index),
		Flags:         int(ctx.flags),
		SideDataElems: int(ctx.side_data_elems),
		Duration:      int64(ctx.duration),
		Pos:           int64(ctx.pos),
	})
}

func (ctx *AVCodecParameters) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVCodecParameters{
		CodecType: AVMediaType(ctx.codec_type),
		CodecID:   AVCodecID(ctx.codec_id),
		CodecTag:  uint32(ctx.codec_tag),
		Format:    int(ctx.format),
		BitRate:   int64(ctx.bit_rate),
	})
}

func (ctx *AVCodec) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVCodec{
		Name:         C.GoString(ctx.name),
		LongName:     C.GoString(ctx.long_name),
		Type:         AVMediaType(ctx._type),
		ID:           AVCodecID(ctx.id),
		Capabilities: int(ctx.capabilities),
	})
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVPacket) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

func (ctx *AVCodecParameters) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

func (ctx *AVCodec) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

////////////////////////////////////////////////////////////////////////////////
// AVCodecParameters

func (ctx *AVCodecParameters) CodecType() AVMediaType {
	return AVMediaType(ctx.codec_type)
}

func (ctx *AVCodecParameters) CodecID() AVCodecID {
	return AVCodecID(ctx.codec_id)
}

func (ctx *AVCodecParameters) CodecTag() uint32 {
	return uint32(ctx.codec_tag)
}

func (ctx *AVCodecParameters) SetCodecTag(tag uint32) {
	ctx.codec_tag = C.uint32_t(tag)
}

////////////////////////////////////////////////////////////////////////////////
// AVCodec

func (c *AVCodec) SampleFormats() []AVSampleFormat {
	var result []AVSampleFormat
	ptr := uintptr(unsafe.Pointer(c.sample_fmts))
	if ptr == 0 {
		return nil
	}
	for {
		v := AVSampleFormat(*(*C.enum_AVSampleFormat)(unsafe.Pointer(ptr)))
		if v == AV_SAMPLE_FMT_NONE {
			break
		}
		result = append(result, v)
		ptr += unsafe.Sizeof(AV_SAMPLE_FMT_NONE)
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////
// AVPacket

func (ctx *AVPacket) StreamIndex() int {
	return int(ctx.stream_index)
}

func (ctx *AVPacket) Pts() int64 {
	return int64(ctx.pts)
}

func (ctx *AVPacket) Dts() int64 {
	return int64(ctx.dts)
}

func (ctx *AVPacket) Duration() int64 {
	return int64(ctx.duration)
}

func (ctx *AVPacket) Pos() int64 {
	return int64(ctx.pos)
}

func (ctx *AVPacket) SetPos(pos int64) {
	ctx.pos = C.int64_t(pos)
}

////////////////////////////////////////////////////////////////////////////////
// AVCodecContext

func (ctx *AVCodecContext) BitRate() int64 {
	return int64(ctx.bit_rate)
}

func (ctx *AVCodecContext) SetBitRate(bit_rate int64) {
	ctx.bit_rate = C.int64_t(bit_rate)
}

func (ctx *AVCodecContext) SampleFormat() AVSampleFormat {
	return AVSampleFormat(ctx.sample_fmt)
}

func (ctx *AVCodecContext) SetSampleFormat(sample_fmt AVSampleFormat) {
	ctx.sample_fmt = C.enum_AVSampleFormat(sample_fmt)
}

func (ctx *AVCodecContext) SampleRate() int {
	return int(ctx.sample_rate)
}

func (ctx *AVCodecContext) SetSampleRate(sample_rate int) {
	ctx.sample_rate = C.int(sample_rate)
}
