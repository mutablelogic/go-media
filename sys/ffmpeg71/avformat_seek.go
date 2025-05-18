package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Seek to timestamp ts.
func AVFormat_seek_frame(ctx *AVFormatContext, stream int, ts int64, flags AVSeekFlag) error {
	if err := AVError(C.av_seek_frame((*C.struct_AVFormatContext)(ctx), C.int(stream), C.int64_t(ts), C.int(flags))); err != 0 {
		return err
	}
	return nil
}

// Seek to the keyframe at timestamp.
func AVFormat_seek_file(ctx *AVFormatContext, stream int, min_ts, ts, max_ts int64, flags AVSeekFlag) error {
	if err := AVError(C.avformat_seek_file((*C.struct_AVFormatContext)(ctx), C.int(stream), C.int64_t(min_ts), C.int64_t(ts), C.int64_t(max_ts), C.int(flags))); err != 0 {
		return err
	}
	return nil
}
