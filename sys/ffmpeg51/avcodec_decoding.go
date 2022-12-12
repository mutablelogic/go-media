package ffmpeg

import (
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
// PUBLIC METHODS

// Copy codec parameters from input stream to output codec context.
func AVCodec_parameters_to_context(ctx *AVCodecContext, par *AVCodecParameters) error {
	if err := AVError(C.avcodec_parameters_to_context((*C.AVCodecContext)(ctx), (*C.AVCodecParameters)(par))); err != 0 {
		return err
	}
	return nil
}

// Return decoded output data from a decoder or encoder. Error return of
// EAGAIN means that more input is needed to produce output, while EINVAL
// means that the decoder has been flushed and no more output is available.
func AVCodec_receive_frame(ctx *AVCodecContext, frame *AVFrame) error {
	if err := AVError(C.avcodec_receive_frame((*C.AVCodecContext)(ctx), (*C.AVFrame)(frame))); err != 0 {
		if err.IsErrno(syscall.EAGAIN) {
			return syscall.EAGAIN
		} else if err.IsErrno(syscall.EINVAL) {
			return syscall.EINVAL
		} else {
			return err
		}
	}
	return nil
}

// Send a packet with a compressed frame to a decoder. Error return of
// EAGAIN means that more input is needed to produce output, while EINVAL
// means that the decoder has been flushed and no more output is available.
func AVCodec_send_packet(ctx *AVCodecContext, pkt *AVPacket) error {
	if err := AVError(C.avcodec_send_packet((*C.AVCodecContext)(ctx), (*C.AVPacket)(pkt))); err != 0 {
		if err.IsErrno(syscall.EAGAIN) {
			return syscall.EAGAIN
		} else if err.IsErrno(syscall.EINVAL) {
			return syscall.EINVAL
		} else {
			return err
		}
	}
	return nil
}
