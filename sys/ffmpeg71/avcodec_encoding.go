package ffmpeg

import (
	"io"
	"syscall"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec libavformat
#include <libavcodec/avcodec.h>
#include <libavformat/avformat.h>
#include <stdlib.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Copy codec parameters from input stream to output codec context.
// Supply a raw video or audio frame to the encoder.
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

// Read encoded data from the encoder.
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

// Write a packet to an output media file ensuring correct interleaving.
// This function will buffer the packets internally as needed to make sure the packets in the output file are
// properly interleaved, usually ordered by increasing dts. Callers doing their own interleaving should
// call av_write_frame() instead of this function.
func AVCodec_interleaved_write_frame(ctx *AVFormatContext, pkt *AVPacket) error {
	if err := AVError(C.av_interleaved_write_frame((*C.AVFormatContext)(ctx), (*C.AVPacket)(pkt))); err != 0 {
		return err
	}
	// Return success
	return nil
}
