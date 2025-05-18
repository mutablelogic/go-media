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

type jsonAVPacket struct {
	Pts           int64      `json:"pts,omitempty"`
	Dts           int64      `json:"dts,omitempty"`
	Size          int        `json:"size,omitempty"`
	StreamIndex   int        `json:"stream_index"` // Stream index starts at 0
	Flags         int        `json:"flags,omitempty"`
	SideDataElems int        `json:"side_data_elems,omitempty"`
	Duration      int64      `json:"duration,omitempty"`
	TimeBase      AVRational `json:"time_base,omitempty"`
	Pos           int64      `json:"pos,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVPacket) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVPacket{
		Pts:           int64(ctx.pts),
		Dts:           int64(ctx.dts),
		Size:          int(ctx.size),
		StreamIndex:   int(ctx.stream_index),
		Flags:         int(ctx.flags),
		SideDataElems: int(ctx.side_data_elems),
		Duration:      int64(ctx.duration),
		TimeBase:      AVRational(ctx.time_base),
		Pos:           int64(ctx.pos),
	})
}

func (ctx *AVPacket) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Allocate an AVPacket and set its fields to default values.
func AVCodec_packet_alloc() *AVPacket {
	return (*AVPacket)(C.av_packet_alloc())
}

// Free the packet, if the packet is reference counted, it will be unreferenced first.
func AVCodec_packet_free(pkt *AVPacket) {
	C.av_packet_free((**C.struct_AVPacket)(unsafe.Pointer(&pkt)))
}

// Create a new packet that references the same data as src.
func AVCodec_packet_clone(src *AVPacket) *AVPacket {
	return (*AVPacket)(C.av_packet_clone((*C.struct_AVPacket)(src)))
}

// Allocate the payload of a packet and initialize its fields with default values.
func AVCodec_new_packet(pkt *AVPacket, size int) error {
	if err := AVError(C.av_new_packet((*C.struct_AVPacket)(pkt), C.int(size))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Reduce packet size, correctly zeroing padding.
func AVCodec_shrink_packet(pkt *AVPacket, size int) {
	C.av_shrink_packet((*C.struct_AVPacket)(pkt), C.int(size))
}

// Increase packet size, correctly zeroing padding.
func AVCodec_grow_packet(pkt *AVPacket, size int) error {
	if err := AVError(C.av_grow_packet((*C.struct_AVPacket)(pkt), C.int(size))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Convert valid timing fields (timestamps / durations) in a packet from one timebase to another.
func AVCodec_packet_rescale_ts(pkt *AVPacket, tb_src, tb_dst AVRational) {
	C.av_packet_rescale_ts((*C.AVPacket)(pkt), (C.AVRational)(tb_src), (C.AVRational)(tb_dst))
}

// Unreference the packet to release the data
func AVCodec_packet_unref(pkt *AVPacket) {
	C.av_packet_unref((*C.struct_AVPacket)(pkt))
}

////////////////////////////////////////////////////////////////////////////////
// AVPacket

func (ctx *AVPacket) StreamIndex() int {
	return int(ctx.stream_index)
}

func (ctx *AVPacket) SetStreamIndex(index int) {
	ctx.stream_index = C.int(index)
}

func (ctx *AVPacket) TimeBase() AVRational {
	if ctx == nil {
		return AVRational{}
	}
	return AVRational(ctx.time_base)
}

func (ctx *AVPacket) SetTimeBase(tb AVRational) {
	ctx.time_base = C.AVRational(tb)
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

func (ctx *AVPacket) Bytes() []byte {
	return C.GoBytes(unsafe.Pointer(ctx.data), C.int(ctx.size))
}

func (ctx *AVPacket) Size() int {
	return int(ctx.size)
}
