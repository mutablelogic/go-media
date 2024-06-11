package ffmpeg

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Allocate an AVPacket and set its fields to default values.
func AVCodec_av_packet_alloc() *AVPacket {
	return (*AVPacket)(C.av_packet_alloc())
}

// Free the packet, if the packet is reference counted, it will be unreferenced first.
func AVCodec_av_packet_free(pkt *AVPacket) {
	C.av_packet_free((**C.struct_AVPacket)(unsafe.Pointer(&pkt)))
}

// Create a new packet that references the same data as src.
func AVCodec_av_packet_clone(src *AVPacket) *AVPacket {
	return (*AVPacket)(C.av_packet_clone((*C.struct_AVPacket)(src)))
}

// Allocate the payload of a packet and initialize its fields with default values.
func AVCodec_av_new_packet(pkt *AVPacket, size int) error {
	if err := AVError(C.av_new_packet((*C.struct_AVPacket)(pkt), C.int(size))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Reduce packet size, correctly zeroing padding.
func AVCodec_av_shrink_packet(pkt *AVPacket, size int) {
	C.av_shrink_packet((*C.struct_AVPacket)(pkt), C.int(size))
}

// Increase packet size, correctly zeroing padding.
func AVCodec_av_grow_packet(pkt *AVPacket, size int) error {
	if err := AVError(C.av_grow_packet((*C.struct_AVPacket)(pkt), C.int(size))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Convert valid timing fields (timestamps / durations) in a packet from one timebase to another.
func AVCodec_av_packet_rescale_ts(pkt *AVPacket, tb_src, tb_dst AVRational) {
	C.av_packet_rescale_ts((*C.AVPacket)(pkt), (C.AVRational)(tb_src), (C.AVRational)(tb_dst))
}

// Unreference the packet to release the data
func AVCodec_av_packet_unref(pkt *AVPacket) {
	C.av_packet_unref((*C.struct_AVPacket)(pkt))
}
