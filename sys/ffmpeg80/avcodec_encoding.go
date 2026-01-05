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
