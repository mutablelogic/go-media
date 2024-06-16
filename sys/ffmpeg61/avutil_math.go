package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////

// Compare two timestamps each in its own time base. Returns -1 if a is before b, 1 if a is after b, or 0 if they are equal.
func AVUtil_compare_ts(a int64, a_tb AVRational, b int64, b_tb AVRational) int {
	return int(C.av_compare_ts(C.int64_t(a), C.AVRational(a_tb), C.int64_t(b), C.AVRational(b_tb)))
}
