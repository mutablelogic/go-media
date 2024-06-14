package ffmpeg

import (
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#include <libavutil/buffer.h>
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

func (ctx *AVFrame) Width() int {
	return int(ctx.width)
}

func (ctx *AVFrame) SetWidth(width int) {
	ctx.width = C.int(width)
}

func (ctx *AVFrame) Height() int {
	return int(ctx.height)
}

func (ctx *AVFrame) SetHeight(height int) {
	ctx.height = C.int(height)
}

func (ctx *AVFrame) PixFmt() AVPixelFormat {
	return AVPixelFormat(ctx.format)
}

func (ctx *AVFrame) SetPixFmt(format AVPixelFormat) {
	ctx.format = C.int(format)
}

func (ctx *AVFrame) Linesize(plane int) int {
	if plane < 0 || plane >= int(C.AV_NUM_DATA_POINTERS) {
		return 0
	}
	return int(ctx.linesize[plane])
}

// Return a buffer reference to the data for a plane.
func (ctx *AVFrame) BufferRef(plane int) *AVBufferRef {
	return (*AVBufferRef)(C.av_frame_get_plane_buffer((*C.AVFrame)(ctx), C.int(plane)))
}

func (ctx *AVFrame) Pts() int64 {
	return int64(ctx.pts)
}

func (ctx *AVFrame) SetPts(pts int64) {
	ctx.pts = C.int64_t(pts)
}

// Returns a plane as a uint8 array.
func (ctx *AVFrame) Uint8(plane int) []uint8 {
	if buf := ctx.BufferRef(plane); buf == nil {
		return nil
	} else {
		return cUint8Slice(unsafe.Pointer(buf.data), C.int(buf.size))
	}
}

// Returns a plane as a int16 array.
func (ctx *AVFrame) Int16(plane int) []int16 {
	if buf := ctx.BufferRef(plane); buf == nil {
		return nil
	} else {
		return cInt16Slice(unsafe.Pointer(buf.data), C.int(buf.size)>>1)
	}
}
