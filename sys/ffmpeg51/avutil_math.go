package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/mathematics.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Rescale a 64-bit integer with rounding to nearest. The operation is mathematically
// equivalent to `a * b / c`, but writing that udirectly can overflow.
func AVUtil_av_rescale(a, b, c int64) int64 {
	return int64(C.av_rescale(C.int64_t(a), C.int64_t(b), C.int64_t(c)))
}

// Rescale a 64-bit integer with specified rounding. The operation is mathematically
// equivalent to `a * b / c``, but writing that directly can overflow, and does not
// support different rounding methods. If the result is not representable then INT64_MIN is returned.
func AVUtil_av_rescale_rnd(a, b, c int64, rnd AVRounding) int64 {
	return int64(C.av_rescale_rnd(C.int64_t(a), C.int64_t(b), C.int64_t(c), C.enum_AVRounding(rnd)))
}
