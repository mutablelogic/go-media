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
// ENCODING

// Supply a raw video or audio frame to the encoder.
// Pass nil to flush the encoder and signal end of stream.
// Returns syscall.EAGAIN if the encoder cannot accept more input at this time.
// Returns syscall.EINVAL if the codec is not opened or requires encoder parameters.
func AVCodec_send_frame(ctx *AVCodecContext, frame *AVFrame) error {
	if err := AVError(C.avcodec_send_frame((*C.AVCodecContext)(ctx), (*C.AVFrame)(frame))); err != 0 {
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

// Return encoded output data from an encoder.
// Returns syscall.EAGAIN if more input is needed to produce output.
// Returns io.EOF when the encoder has been flushed and no more output is available.
// Returns syscall.EINVAL if the codec is not opened or requires decoder parameters.
func AVCodec_receive_packet(ctx *AVCodecContext, pkt *AVPacket) error {
	if err := AVError(C.avcodec_receive_packet((*C.AVCodecContext)(ctx), (*C.AVPacket)(pkt))); err != 0 {
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
// SUBTITLE ENCODING

// Encode subtitles to a buffer. Returns the number of bytes written to buf
// on success, or a negative error code on failure.
// Note: Subtitles use a legacy API and don't support the send/receive pattern.
func AVCodec_encode_subtitle(ctx *AVCodecContext, buf []byte, sub *AVSubtitle) (int, error) {
	if len(buf) == 0 {
		return 0, syscall.EINVAL
	}
	ret := int(C.avcodec_encode_subtitle(
		(*C.AVCodecContext)(ctx),
		(*C.uint8_t)(&buf[0]),
		C.int(len(buf)),
		(*C.AVSubtitle)(sub),
	))
	if ret < 0 {
		return 0, AVError(ret)
	}
	return ret, nil
}
