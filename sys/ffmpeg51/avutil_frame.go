package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/frame.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Allocate an AVFrame and set its fields to default values.
func AVUtil_frame_alloc() *AVFrame {
	return (*AVFrame)(C.av_frame_alloc())
}

// Free the frame and any dynamically allocated objects in it
func AVUtil_frame_free(frame **AVFrame) {
	C.av_frame_free((**C.AVFrame)(unsafe.Pointer(frame)))
}

// Free the frame and any dynamically allocated objects in it
func AVUtil_frame_free_ptr(frame *AVFrame) {
	C.av_frame_free((**C.AVFrame)(unsafe.Pointer(&frame)))
}

// Unreference all the buffers referenced by frame and reset the frame fields.
func AVUtil_frame_unref(frame *AVFrame) {
	C.av_frame_unref((*C.AVFrame)(frame))
}

/*
func NewAudioFrame(f AVSampleFormat, rate int, layout AVChannelLayout) *AVFrame {
	frame := NewAVFrame()
	if frame == nil {
		return nil
	}
	ctx := (*C.AVFrame)(frame)
	ctx.format = C.int(f)
	ctx.sample_rate = C.int(rate)
	ctx.channel_layout = C.uint64_t(layout)
	ctx.channels = C.av_get_channel_layout_nb_channels(C.uint64_t(layout))
	return frame
}
*/
