package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavfilter
#include <libavfilter/buffersrc.h>
#include <libavfilter/buffersink.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

type AVBufferSrcFlag C.int

const (
	AV_BUFFERSRC_FLAG_NONE            AVBufferSrcFlag = 0
	AV_BUFFERSRC_FLAG_NO_CHECK_FORMAT AVBufferSrcFlag = C.AV_BUFFERSRC_FLAG_NO_CHECK_FORMAT
	AV_BUFFERSRC_FLAG_PUSH            AVBufferSrcFlag = C.AV_BUFFERSRC_FLAG_PUSH
	AV_BUFFERSRC_FLAG_KEEP_REF        AVBufferSrcFlag = C.AV_BUFFERSRC_FLAG_KEEP_REF
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - BUFFERSRC

// Add a frame to the buffer source.
func AVBufferSrc_add_frame_flags(ctx *AVFilterContext, frame *AVFrame, flags AVBufferSrcFlag) error {
	var cFrame *C.AVFrame
	if frame != nil {
		cFrame = (*C.AVFrame)(frame)
	}
	if err := AVError(C.av_buffersrc_add_frame_flags((*C.AVFilterContext)(ctx), cFrame, C.int(flags))); err != 0 {
		return err
	}
	return nil
}

// Add a frame to the buffer source (simplified version without flags).
func AVBufferSrc_add_frame(ctx *AVFilterContext, frame *AVFrame) error {
	return AVBufferSrc_add_frame_flags(ctx, frame, AV_BUFFERSRC_FLAG_NONE)
}

// Write a frame to the buffer source.
func AVBufferSrc_write_frame(ctx *AVFilterContext, frame *AVFrame) error {
	var cFrame *C.AVFrame
	if frame != nil {
		cFrame = (*C.AVFrame)(frame)
	}
	if err := AVError(C.av_buffersrc_write_frame((*C.AVFilterContext)(ctx), cFrame)); err != 0 {
		return err
	}
	return nil
}

// Close the buffer source.
func AVBufferSrc_close(ctx *AVFilterContext, pts int64, flags AVBufferSrcFlag) error {
	if err := AVError(C.av_buffersrc_close((*C.AVFilterContext)(ctx), C.int64_t(pts), C.uint(flags))); err != 0 {
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - BUFFERSINK

// Get a frame from the buffer sink.
func AVBufferSink_get_frame(ctx *AVFilterContext, frame *AVFrame) error {
	if err := AVError(C.av_buffersink_get_frame((*C.AVFilterContext)(ctx), (*C.AVFrame)(frame))); err != 0 {
		return err
	}
	return nil
}

// Get a frame with flags from the buffer sink.
func AVBufferSink_get_frame_flags(ctx *AVFilterContext, frame *AVFrame, flags int) error {
	if err := AVError(C.av_buffersink_get_frame_flags((*C.AVFilterContext)(ctx), (*C.AVFrame)(frame), C.int(flags))); err != 0 {
		return err
	}
	return nil
}

// Set the frame size for the buffer sink.
func AVBufferSink_set_frame_size(ctx *AVFilterContext, frame_size uint) {
	C.av_buffersink_set_frame_size((*C.AVFilterContext)(ctx), C.uint(frame_size))
}

// Get the frame rate from the buffer sink.
func AVBufferSink_get_frame_rate(ctx *AVFilterContext) AVRational {
	return AVRational(C.av_buffersink_get_frame_rate((*C.AVFilterContext)(ctx)))
}

// Get the sample aspect ratio from the buffer sink.
func AVBufferSink_get_sample_aspect_ratio(ctx *AVFilterContext) AVRational {
	return AVRational(C.av_buffersink_get_sample_aspect_ratio((*C.AVFilterContext)(ctx)))
}

// Get the width from the buffer sink.
func AVBufferSink_get_w(ctx *AVFilterContext) int {
	return int(C.av_buffersink_get_w((*C.AVFilterContext)(ctx)))
}

// Get the height from the buffer sink.
func AVBufferSink_get_h(ctx *AVFilterContext) int {
	return int(C.av_buffersink_get_h((*C.AVFilterContext)(ctx)))
}

// Get the pixel format from the buffer sink.
func AVBufferSink_get_format(ctx *AVFilterContext) AVPixelFormat {
	return AVPixelFormat(C.av_buffersink_get_format((*C.AVFilterContext)(ctx)))
}

// Get the time base from the buffer sink.
func AVBufferSink_get_time_base(ctx *AVFilterContext) AVRational {
	return AVRational(C.av_buffersink_get_time_base((*C.AVFilterContext)(ctx)))
}

// Get the sample rate from the buffer sink (audio).
func AVBufferSink_get_sample_rate(ctx *AVFilterContext) int {
	return int(C.av_buffersink_get_sample_rate((*C.AVFilterContext)(ctx)))
}

// Get the channel layout from the buffer sink (audio).
func AVBufferSink_get_ch_layout(ctx *AVFilterContext) AVChannelLayout {
	var layout C.AVChannelLayout
	C.av_buffersink_get_ch_layout((*C.AVFilterContext)(ctx), &layout)
	return AVChannelLayout(layout)
}
