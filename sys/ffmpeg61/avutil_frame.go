package ffmpeg

import (
	"encoding/json"
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
// STRINGIFY

type jsonAVFrame struct {
	SampleFormat  AVSampleFormat  `json:"sample_format,omitempty"`
	NumSamples    int             `json:"num_samples,omitempty"`
	SampleRate    int             `json:"sample_rate,omitempty"`
	ChannelLayout AVChannelLayout `json:"channel_layout,omitempty"`
	PixelFormat   AVPixelFormat   `json:"pixel_format"`
	Width         int             `json:"width,omitempty"`
	Height        int             `json:"height,omitempty"`
	PictureType   AVPictureType   `json:"picture_type,omitempty"`
	Pts           AVTimestamp     `json:"pts,omitempty"`
	BestEffortTs  AVTimestamp     `json:"best_effort_timestamp,omitempty"`
	TimeBase      AVRational      `json:"time_base,omitempty"`
	NumPlanes     int             `json:"num_planes,omitempty"`
}

func (ctx *AVFrame) MarshalJSON() ([]byte, error) {
	if ctx.nb_samples > 0 {
		// Audio
		return json.Marshal(jsonAVFrame{
			NumSamples:    int(ctx.nb_samples),
			SampleFormat:  AVSampleFormat(ctx.format),
			SampleRate:    int(ctx.sample_rate),
			ChannelLayout: AVChannelLayout(ctx.ch_layout),
			Pts:           AVTimestamp(ctx.pts),
			BestEffortTs:  AVTimestamp(ctx.best_effort_timestamp),
			TimeBase:      AVRational(ctx.time_base),
			NumPlanes:     AVUtil_frame_get_num_planes(ctx),
		})
	} else {
		// Video
		return json.Marshal(jsonAVFrame{
			PixelFormat: AVPixelFormat(ctx.format),
			Width:       int(ctx.width),
			Height:      int(ctx.height),
			Pts:         AVTimestamp(ctx.pts),
			TimeBase:    AVRational(ctx.time_base),
			NumPlanes:   AVUtil_frame_get_num_planes(ctx),
		})
	}
}

func (ctx *AVFrame) String() string {
	data, _ := json.MarshalIndent(ctx, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

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
// format, width and height for video,
// format, nb_samples and ch_layout for audio
func AVUtil_frame_get_buffer(frame *AVFrame, align bool) error {
	if ret := AVError(C.av_frame_get_buffer((*C.struct_AVFrame)(frame), boolToInt(align))); ret != 0 {
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

// Return the number of planes in the frame data.
func AVUtil_frame_get_num_planes(frame *AVFrame) int {
	if frame.nb_samples > 0 {
		// Audio
		if AVUtil_sample_fmt_is_planar(AVSampleFormat(frame.format)) {
			return int(frame.ch_layout.nb_channels)
		} else {
			return 1
		}
	} else {
		// Video
		desc := AVUtil_get_pix_fmt_desc(AVPixelFormat(frame.format))
		return int(desc.nb_components)
	}
}

// Copy only "metadata" fields from src to dst, those fields that do not affect the data layout in the buffers.
// E.g. pts, sample rate (for audio) or sample aspect ratio (for video), but not width/height or channel layout.
// Side data is also copied.
func AVUtil_frame_copy_props(dst, src *AVFrame) error {
	if ret := AVError(C.av_frame_copy_props((*C.struct_AVFrame)(dst), (*C.struct_AVFrame)(src))); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

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

func (ctx *AVFrame) SampleRate() int {
	return int(ctx.sample_rate)
}

func (ctx *AVFrame) SetSampleRate(sample_rate int) {
	ctx.sample_rate = C.int(sample_rate)
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

func (ctx *AVFrame) Pts() int64 {
	return int64(ctx.pts)
}

func (ctx *AVFrame) BestEffortTs() int64 {
	return int64(ctx.best_effort_timestamp)
}

func (ctx *AVFrame) SetPts(pts int64) {
	ctx.pts = C.int64_t(pts)
}

func (ctx *AVFrame) TimeBase() AVRational {
	return AVRational(ctx.time_base)
}

// Return stride of a plane for images, or plane size for audio.
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

// Returns the data as a set of planes and strides
func (ctx *AVFrame) Data() ([][]byte, []int) {
	planes := make([][]byte, int(C.AV_NUM_DATA_POINTERS))
	strides := make([]int, int(C.AV_NUM_DATA_POINTERS))
	for i := 0; i < int(C.AV_NUM_DATA_POINTERS); i++ {
		planes[i] = ctx.Uint8(i)
		strides[i] = ctx.Linesize(i)
	}
	return planes, strides
}
