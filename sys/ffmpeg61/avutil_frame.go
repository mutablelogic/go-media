package ffmpeg

import (
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/frame.h>
#include <stdlib.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Allocate an AVFrame and set its fields to default values.
func AVUtil_frame_alloc() *AVFrame {
	return (*AVFrame)(C.av_frame_alloc())
}

// Free the frame and any dynamically allocated objects in it
func AVUtil_frame_free(frame *AVFrame) {
	C.av_frame_free((**C.AVFrame)(unsafe.Pointer(&frame)))
}

// Unreference all the buffers referenced by frame and reset the frame fields.
func AVUtil_frame_unref(frame *AVFrame) {
	C.av_frame_unref((*C.AVFrame)(frame))
}

// Allocate new buffer(s) for audio or video data.
// The following fields must be set on frame before calling this function:
// format (pixel format for video, sample format for audio), width and height for video, nb_samples and ch_layout for audio
func AVUtil_frame_get_buffer(frame *AVFrame, align int) error {
	if ret := AVError(C.av_frame_get_buffer((*C.struct_AVFrame)(frame), C.int(align))); ret != 0 {
		return ret
	}
	return nil
}

// Ensure that the frame data is writable, avoiding data copy if possible.
// Do nothing if the frame is writable, allocate new buffers and copy the data if it is not.
// Non-refcounted frames behave as non-writable, i.e. a copy is always made.
func AVUtil_frame_make_writable(frame *AVFrame) error {
	if ret := AVError(C.av_frame_make_writable((*C.struct_AVFrame)(frame))); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (ctx *AVFrame) NumSamples() int {
	return int(ctx.nb_samples)
}

func (ctx *AVFrame) SetNumSamples(nb_samples int) {
	ctx.nb_samples = C.int(nb_samples)
}
func (ctx *AVFrame) SampleFormat() AVSampleFormat {
	return AVSampleFormat(ctx.format)
}

func (ctx *AVFrame) SetSampleFormat(format AVSampleFormat) {
	ctx.format = C.int(format)
}

func (ctx *AVFrame) ChannelLayout() AVChannelLayout {
	return AVChannelLayout(ctx.ch_layout)
}

func (ctx *AVFrame) SetChannelLayout(src AVChannelLayout) error {
	if ret := AVError(C.av_channel_layout_copy((*C.struct_AVChannelLayout)(&ctx.ch_layout), (*C.struct_AVChannelLayout)(&src))); ret != 0 {
		return ret
	}
	return nil
}

// Returns a plane as a uint8 array.
func (ctx *AVFrame) Uint8(plane int) []byte {
	return cByteSlice(unsafe.Pointer(&ctx.data[plane]), C.int(ctx.linesize[plane]))
}

// Returns a plane as a uint16 array.
func (ctx *AVFrame) Uint16(plane int) []uint16 {
	return cUint16Slice(unsafe.Pointer(&ctx.data[plane]), C.int(ctx.linesize[plane]>>1))
}

// Returns a plane as a int16 array.
func (ctx *AVFrame) Int16(plane int) []int16 {
	return cInt16Slice(unsafe.Pointer(&ctx.data[plane]), C.int(ctx.linesize[plane]>>1))
}
