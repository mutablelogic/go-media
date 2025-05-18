package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/timestamp.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	buf [C.AV_TS_MAX_STRING_SIZE]C.char
)

func AVUtil_ts_make_string(ts int64) string {
	return C.GoString(C.av_ts_make_string(&buf[0], C.int64_t(ts)))
}

func AVUtil_ts_make_time_string(ts int64, tb *AVRational) string {
	return C.GoString(C.av_ts_make_time_string(&buf[0], C.int64_t(ts), (*C.struct_AVRational)(tb)))
}

func AVUtil_ts2str(ts int64) string {
	return AVUtil_ts_make_string(ts)
}

func AVUtil_ts2timestr(ts int64, tb *AVRational) string {
	return AVUtil_ts_make_time_string(ts, tb)
}
