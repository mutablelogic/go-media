package ffmpeg

import (
	"io"
	"syscall"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
#include <stdlib.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// DECODING

// Return decoded output data from a decoder.
// Returns syscall.EAGAIN if more input is needed to produce output.
// Returns io.EOF when the decoder has been flushed and no more output is available.
// Returns syscall.EINVAL if the codec is not opened or requires encoder parameters.
func AVCodec_receive_frame(ctx *AVCodecContext, frame *AVFrame) error {
	if err := AVError(C.avcodec_receive_frame((*C.AVCodecContext)(ctx), (*C.AVFrame)(frame))); err != 0 {
		if err == AVERROR_EOF {
			return io.EOF
		} else if err.IsErrno(syscall.EAGAIN) {
			return syscall.EAGAIN
		} else if err.IsErrno(syscall.EINVAL) {
			return syscall.EINVAL
		} else {
			return err
		}
	}
	return nil
}

// Supply raw packet data as input to a decoder.
// Pass nil to flush the decoder and signal end of stream.
// Returns syscall.EAGAIN if the decoder cannot accept more input at this time.
// Returns syscall.EINVAL if the codec is not opened or requires decoder parameters.
func AVCodec_send_packet(ctx *AVCodecContext, pkt *AVPacket) error {
	if err := AVError(C.avcodec_send_packet((*C.AVCodecContext)(ctx), (*C.AVPacket)(pkt))); err != 0 {
		if err == AVERROR_EOF {
			return io.EOF
		} else if err.IsErrno(syscall.EAGAIN) {
			return syscall.EAGAIN
		} else if err.IsErrno(syscall.EINVAL) {
			return syscall.EINVAL
		} else {
			return err
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// SUBTITLE DECODING

// Decode a subtitle message. Return a negative value on error, otherwise
// return the number of bytes used. If no subtitle could be decompressed,
// got_sub_ptr is zero. Otherwise, the subtitle is stored in *sub.
// Note: Subtitles use a legacy API and don't support the send/receive pattern.
func AVCodec_decode_subtitle2(ctx *AVCodecContext, sub *AVSubtitle, got_sub_ptr *int, pkt *AVPacket) error {
	var got_sub C.int
	if err := AVError(C.avcodec_decode_subtitle2(
		(*C.AVCodecContext)(ctx),
		(*C.AVSubtitle)(sub),
		&got_sub,
		(*C.AVPacket)(pkt),
	)); err != 0 {
		return err
	}
	if got_sub_ptr != nil {
		*got_sub_ptr = int(got_sub)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// FLUSH

// Reset the internal codec state and flush internal buffers.
// Should be called before seeking or when switching to a different stream.
// Note: This does not flush the AVCodecContext, only the internal codec buffers.
func AVCodec_flush_buffers(ctx *AVCodecContext) {
	C.avcodec_flush_buffers((*C.AVCodecContext)(ctx))
}
