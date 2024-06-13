package ffmpeg

import "io"

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

// Return decoded output data from a decoder or encoder. Error return of
// EAGAIN means that more input is needed to produce output, while EINVAL
// means that the decoder has been flushed and no more output is available.
func AVCodec_receive_frame(ctx *AVCodecContext, frame *AVFrame) error {
	if err := AVError(C.avcodec_receive_frame((*C.AVCodecContext)(ctx), (*C.AVFrame)(frame))); err != 0 {
		if err == AVERROR_EOF {
			return io.EOF
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
		if err == AVERROR_EOF {
			return io.EOF
		} else {
			return err
		}
	}
	return nil
}
