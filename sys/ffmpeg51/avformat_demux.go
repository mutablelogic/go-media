package ffmpeg

import "io"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the next frame of a stream. Returns io.EOF when no more frames are available.
// TODO: Check for io.EOF
func AVFormat_av_read_frame(ctx *AVFormatContext, pkt *AVPacket) error {
	if err := AVError(C.av_read_frame((*C.struct_AVFormatContext)(ctx), (*C.struct_AVPacket)(pkt))); err < 0 {
		if err == AVERROR_EOF {
			return io.EOF
		} else {
			return err
		}
	} else {
		return nil
	}
}
