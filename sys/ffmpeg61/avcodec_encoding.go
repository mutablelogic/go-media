package ffmpeg

import (
	"io"
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
// Supply a raw video or audio frame to the encoder.
func AVCodec_send_frame(ctx *AVCodecContext, frame *AVFrame) error {
	if err := AVError(C.avcodec_send_frame((*C.AVCodecContext)(ctx), (*C.AVFrame)(frame))); err != 0 {
		if err == AVERROR_EOF {
			return io.EOF
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
		} else {
			return err
		}
	}
	return nil
}
