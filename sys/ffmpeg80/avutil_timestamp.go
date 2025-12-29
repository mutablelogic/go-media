package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#include <libavutil/timestamp.h>
*/
import "C"
import "encoding/json"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVTimestamp C.int64_t
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_NOPTS_VALUE AVTimestamp = C.AV_NOPTS_VALUE
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Convert a timestamp to a string representation
func AVUtil_ts_make_string(ts int64) string {
	var buf [C.AV_TS_MAX_STRING_SIZE]C.char
	return C.GoString(C.av_ts_make_string(&buf[0], C.int64_t(ts)))
}

// Convert a timestamp and time base to a time string representation
func AVUtil_ts_make_time_string(ts int64, tb *AVRational) string {
	var buf [C.AV_TS_MAX_STRING_SIZE]C.char
	return C.GoString(C.av_ts_make_time_string(&buf[0], C.int64_t(ts), (*C.struct_AVRational)(tb)))
}

// Convenience function: convert timestamp to string
func AVUtil_ts2str(ts int64) string {
	return AVUtil_ts_make_string(ts)
}

// Convenience function: convert timestamp and time base to time string
func AVUtil_ts2timestr(ts int64, tb *AVRational) string {
	return AVUtil_ts_make_time_string(ts, tb)
}

////////////////////////////////////////////////////////////////////////////////
// AVTimestamp

func (v AVTimestamp) MarshalJSON() ([]byte, error) {
	if v == AV_NOPTS_VALUE {
		return json.Marshal(nil)
	} else {
		return json.Marshal(int64(v))
	}
}
