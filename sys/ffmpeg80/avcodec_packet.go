package ffmpeg

import (
	"encoding/json"
	"errors"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec
#include <libavcodec/packet.h>
#include <string.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVPacket C.struct_AVPacket
)

////////////////////////////////////////////////////////////////////////////////
// JSON OUTPUT

func (ctx *AVPacket) MarshalJSON() ([]byte, error) {
	type jsonAVPacket struct {
		Pts           int64      `json:"pts,omitempty"`
		Dts           int64      `json:"dts,omitempty"`
		Size          int        `json:"size,omitempty"`
		StreamIndex   int        `json:"stream_index"`
		Flags         int        `json:"flags,omitempty"`
		SideDataElems int        `json:"side_data_elems,omitempty"`
		Duration      int64      `json:"duration,omitempty"`
		TimeBase      AVRational `json:"time_base,omitempty"`
		Pos           int64      `json:"pos,omitempty"`
	}
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

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVPacket) String() string {
	return marshalToString(ctx)
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_PKT_FLAG_KEY     = C.AV_PKT_FLAG_KEY     // Packet contains a keyframe
	AV_PKT_FLAG_CORRUPT = C.AV_PKT_FLAG_CORRUPT // Packet is corrupted
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Allocate an AVPacket and set its fields to default values.
func AVCodec_packet_alloc() *AVPacket {
	return (*AVPacket)(C.av_packet_alloc())
}

// Free the packet, if the packet is reference counted, it will be unreferenced first.
func AVCodec_packet_free(packet *AVPacket) {
	C.av_packet_free((**C.struct_AVPacket)(unsafe.Pointer(&packet)))
}

// Free the packet and set the caller's pointer to nil.
//
// Prefer this over AVCodec_packet_free when you want to avoid accidental
// use-after-free of a stale Go pointer.
func AVCodec_packet_freep(packet **AVPacket) {
	if packet == nil {
		return
	}
	C.av_packet_free((**C.struct_AVPacket)(unsafe.Pointer(packet)))
}

// Wipe the packet and unreference any buffer.
func AVCodec_packet_unref(packet *AVPacket) {
	C.av_packet_unref((*C.struct_AVPacket)(packet))
}

// Setup dst as a reference to src (increment reference count without copying data).
// More efficient than clone when you just need another reference to the same data.
func AVCodec_packet_ref(dst, src *AVPacket) error {
	if err := AVError(C.av_packet_ref((*C.struct_AVPacket)(dst), (*C.struct_AVPacket)(src))); err != 0 {
		return err
	}
	return nil
}

// Create a new packet that references the same data as src.
func AVCodec_packet_clone(src *AVPacket) *AVPacket {
	return (*AVPacket)(C.av_packet_clone((*C.struct_AVPacket)(src)))
}

// Allocate the payload of a packet and initialize its fields with default values.
func AVCodec_new_packet(pkt *AVPacket, size int) error {
	if err := AVError(C.av_new_packet((*C.struct_AVPacket)(pkt), C.int(size))); err != 0 {
		return err
	}
	return nil
}

// Reduce packet size, correctly zeroing padding.
func AVCodec_shrink_packet(pkt *AVPacket, size int) {
	C.av_shrink_packet((*C.struct_AVPacket)(pkt), C.int(size))
}

// Increase packet size, correctly zeroing padding.
func AVCodec_grow_packet(pkt *AVPacket, size int) error {
	if err := AVError(C.av_grow_packet((*C.struct_AVPacket)(pkt), C.int(size))); err != 0 {
		return err
	}
	return nil
}

// Convert valid timing fields (timestamps / durations) in a packet from one timebase to another.
func AVCodec_packet_rescale_ts(pkt *AVPacket, tb_src, tb_dst AVRational) {
	C.av_packet_rescale_ts((*C.struct_AVPacket)(pkt), (C.struct_AVRational)(tb_src), (C.struct_AVRational)(tb_dst))
}

// Create a new packet with a copy of the provided data
func AVCodec_packet_from_data(pkt *AVPacket, data []byte) error {
	if pkt == nil {
		return errors.New("nil packet")
	}
	if len(data) == 0 {
		return errors.New("empty data")
	}
	// Allocate buffer and copy data
	if err := C.av_new_packet((*C.struct_AVPacket)(pkt), C.int(len(data))); err < 0 {
		return AVError(err)
	}
	C.memcpy(unsafe.Pointer(pkt.data), unsafe.Pointer(&data[0]), C.size_t(len(data)))
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// AVPacket METHODS

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
	ctx.time_base = C.struct_AVRational(tb)
}

func (ctx *AVPacket) Pts() int64 {
	return int64(ctx.pts)
}

func (ctx *AVPacket) SetPts(pts int64) {
	ctx.pts = C.int64_t(pts)
}

func (ctx *AVPacket) Dts() int64 {
	return int64(ctx.dts)
}

func (ctx *AVPacket) SetDts(dts int64) {
	ctx.dts = C.int64_t(dts)
}

func (ctx *AVPacket) Duration() int64 {
	return int64(ctx.duration)
}

func (ctx *AVPacket) SetDuration(duration int64) {
	ctx.duration = C.int64_t(duration)
}

func (ctx *AVPacket) Pos() int64 {
	return int64(ctx.pos)
}

func (ctx *AVPacket) SetPos(pos int64) {
	ctx.pos = C.int64_t(pos)
}

func (ctx *AVPacket) Flags() int {
	return int(ctx.flags)
}

func (ctx *AVPacket) SetFlags(flags int) {
	ctx.flags = C.int(flags)
}

func (ctx *AVPacket) Bytes() []byte {
	if ctx == nil || ctx.size <= 0 || ctx.data == nil {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(ctx.data), C.int(ctx.size))
}

func (ctx *AVPacket) Size() int {
	if ctx == nil {
		return 0
	}
	return int(ctx.size)
}
