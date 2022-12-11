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

// Return decoded output data from a decoder or encoder.
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

// Copy codec parameters from input stream to output codec context.
func AVCodec_parameters_to_context(ctx *AVCodecContext, par *AVCodecParameters) error {
	if err := AVError(C.avcodec_parameters_to_context((*C.AVCodecContext)(ctx), (*C.AVCodecParameters)(par))); err != 0 {
		return err
	}
	return nil
}
