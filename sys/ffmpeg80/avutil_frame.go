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
// TYPES

type (
	AVFrame C.struct_AVFrame
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVFrame) MarshalJSON() ([]byte, error) {
	type jsonAVAudioFrame struct {
		SampleFormat   AVSampleFormat  `json:"sample_format"`
		NumSamples     int             `json:"num_samples"`
		SampleRate     int             `json:"sample_rate"`
		ChannelLayout  AVChannelLayout `json:"channel_layout,omitempty"`
		BytesPerSample int             `json:"bytes_per_sample,omitempty"`
	}

	type jsonAVVideoFrame struct {
		PixelFormat  AVPixelFormat `json:"pixel_format"`
		Width        int           `json:"width"`
		Height       int           `json:"height"`
		SampleAspect AVRational    `json:"sample_aspect_ratio,omitempty"`
		PictureType  AVPictureType `json:"picture_type,omitempty"`
		Stride       []int         `json:"plane_stride,omitempty"`
	}

	type jsonAVFrame struct {
		*jsonAVAudioFrame
		*jsonAVVideoFrame
		NumPlanes  int         `json:"num_planes,omitempty"`
		PlaneBytes []int       `json:"plane_bytes,omitempty"`
		Pts        AVTimestamp `json:"pts"`
		TimeBase   AVRational  `json:"time_base,omitempty"`
	}
	if ctx.sample_rate > 0 && ctx.SampleFormat() != AV_SAMPLE_FMT_NONE {
		// Audio
		return json.Marshal(jsonAVFrame{
			jsonAVAudioFrame: &jsonAVAudioFrame{
				NumSamples:     int(ctx.nb_samples),
				SampleFormat:   AVSampleFormat(ctx.format),
				SampleRate:     int(ctx.sample_rate),
				ChannelLayout:  AVChannelLayout(ctx.ch_layout),
				BytesPerSample: AVUtil_get_bytes_per_sample(AVSampleFormat(ctx.format)),
			},
			Pts:        AVTimestamp(ctx.pts),
			TimeBase:   AVRational(ctx.time_base),
			NumPlanes:  AVUtil_frame_get_num_planes(ctx),
			PlaneBytes: ctx.planesizes(),
		})
	} else if ctx.width != 0 && ctx.height != 0 && ctx.PixFmt() != AV_PIX_FMT_NONE {
		// Video
		return json.Marshal(jsonAVFrame{
			jsonAVVideoFrame: &jsonAVVideoFrame{
				PixelFormat:  AVPixelFormat(ctx.format),
				Width:        int(ctx.width),
				Height:       int(ctx.height),
				SampleAspect: AVRational(ctx.sample_aspect_ratio),
				PictureType:  AVPictureType(ctx.pict_type),
				Stride:       ctx.linesizes(),
			},
			Pts:        AVTimestamp(ctx.pts),
			TimeBase:   AVRational(ctx.time_base),
			NumPlanes:  AVUtil_frame_get_num_planes(ctx),
			PlaneBytes: ctx.planesizes(),
		})
	} else {
		// Other
		return json.Marshal(jsonAVFrame{
			Pts:        AVTimestamp(ctx.pts),
			TimeBase:   AVRational(ctx.time_base),
			NumPlanes:  AVUtil_frame_get_num_planes(ctx),
			PlaneBytes: ctx.planesizes(),
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

func AVUtil_frame_is_allocated(frame *AVFrame) bool {
	return frame.data[0] != nil
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
	} else if frame.width != 0 && frame.height != 0 {
		// Video
		return AVUtil_pix_fmt_count_planes(AVPixelFormat(frame.format))
	}

	// Other
	return 0
}

// Copy frame data
func AVUtil_frame_copy(dst, src *AVFrame) error {
	if ret := AVError(C.av_frame_copy((*C.struct_AVFrame)(dst), (*C.struct_AVFrame)(src))); ret < 0 {
		return ret
	}
	return nil
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

func (ctx *AVFrame) SampleAspectRatio() AVRational {
	return AVRational(ctx.sample_aspect_ratio)
}

func (ctx *AVFrame) SetSampleAspectRatio(aspect_ratio AVRational) {
	ctx.sample_aspect_ratio = C.struct_AVRational(aspect_ratio)
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

func (ctx *AVFrame) SetPts(pts int64) {
	ctx.pts = C.int64_t(pts)
}

func (ctx *AVFrame) TimeBase() AVRational {
	return AVRational(ctx.time_base)
}

func (ctx *AVFrame) SetTimeBase(timeBase AVRational) {
	ctx.time_base = C.struct_AVRational(timeBase)
}

// Return stride of a plane for images, or plane size for audio.
func (ctx *AVFrame) Linesize(plane int) int {
	if plane < 0 || plane >= int(C.AV_NUM_DATA_POINTERS) {
		return 0
	}
	return int(ctx.linesize[plane])
}

// Return size of a plane in bytes
func (ctx *AVFrame) Planesize(plane int) int {
	if plane < 0 || plane >= int(C.AV_NUM_DATA_POINTERS) {
		return 0
	}
	if ctx.NumSamples() > 0 && ctx.SampleFormat() != AV_SAMPLE_FMT_NONE {
		// For audio: planar formats have one channel per plane, packed formats have all channels in one plane
		bytesPerSample := AVUtil_get_bytes_per_sample(AVSampleFormat(ctx.format))
		if AVUtil_sample_fmt_is_planar(AVSampleFormat(ctx.format)) {
			// Planar: each plane contains one channel
			return bytesPerSample * ctx.NumSamples()
		} else {
			// Packed: plane 0 contains all channels interleaved
			if plane == 0 {
				return bytesPerSample * ctx.NumSamples() * ctx.ChannelLayout().NumChannels()
			} else {
				return 0
			}
		}
	} else if ctx.Height() > 0 && ctx.PixFmt() != AV_PIX_FMT_NONE {
		return ctx.Linesize(plane) * ctx.Height()
	} else {
		return 0
	}
}

// Return all strides.
func (ctx *AVFrame) linesizes() []int {
	var linesizes []int

	if !AVUtil_frame_is_allocated(ctx) {
		return nil
	}
	for i := 0; i < AVUtil_frame_get_num_planes(ctx); i++ {
		linesizes = append(linesizes, ctx.Linesize(i))
	}
	return linesizes
}

// Return all planes sizes
func (ctx *AVFrame) planesizes() []int {
	var planesizes []int

	if !AVUtil_frame_is_allocated(ctx) {
		return nil
	}
	for i := 0; i < AVUtil_frame_get_num_planes(ctx); i++ {
		planesizes = append(planesizes, ctx.Planesize(i))
	}
	return planesizes
}

// Returns a plane as a byte array (same as uint8).
func (ctx *AVFrame) Bytes(plane int) []byte {
	return ctx.Uint8(plane)
}

// Returns a plane as a uint8 array.
func (ctx *AVFrame) Uint8(plane int) []uint8 {
	if !AVUtil_frame_is_allocated(ctx) {
		return nil
	}
	return cUint8Slice(unsafe.Pointer(ctx.data[plane]), C.int(ctx.Planesize(plane)))
}

// Returns a plane as a int8 array.
func (ctx *AVFrame) Int8(plane int) []int8 {
	if !AVUtil_frame_is_allocated(ctx) {
		return nil
	}
	return cInt8Slice(unsafe.Pointer(ctx.data[plane]), C.int(ctx.Planesize(plane)))
}

// Returns a plane as a uint16 array.
func (ctx *AVFrame) Uint16(plane int) []uint16 {
	if !AVUtil_frame_is_allocated(ctx) {
		return nil
	}
	return cUint16Slice(unsafe.Pointer(ctx.data[plane]), C.int(ctx.Planesize(plane)>>1))
}

// Returns a plane as a int16 array.
func (ctx *AVFrame) Int16(plane int) []int16 {
	if !AVUtil_frame_is_allocated(ctx) {
		return nil
	}
	return cInt16Slice(unsafe.Pointer(ctx.data[plane]), C.int(ctx.Planesize(plane)>>1))
}

// Returns a plane as a uint32 array.
func (ctx *AVFrame) Uint32(plane int) []uint32 {
	if !AVUtil_frame_is_allocated(ctx) {
		return nil
	}
	return cUint32Slice(unsafe.Pointer(ctx.data[plane]), C.int(ctx.Planesize(plane)>>2))
}

// Returns a plane as a int32 array.
func (ctx *AVFrame) Int32(plane int) []int32 {
	if !AVUtil_frame_is_allocated(ctx) {
		return nil
	}
	return cInt32Slice(unsafe.Pointer(ctx.data[plane]), C.int(ctx.Planesize(plane)>>2))
}

// Returns a plane as a float32 array.
func (ctx *AVFrame) Float32(plane int) []float32 {
	if !AVUtil_frame_is_allocated(ctx) {
		return nil
	}
	return cFloat32Slice(unsafe.Pointer(ctx.data[plane]), C.int(ctx.Planesize(plane)>>2))
}

// Returns a plane as a float64 array.
func (ctx *AVFrame) Float64(plane int) []float64 {
	if !AVUtil_frame_is_allocated(ctx) {
		return nil
	}
	return cFloat64Slice(unsafe.Pointer(ctx.data[plane]), C.int(ctx.Planesize(plane)>>3))
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
