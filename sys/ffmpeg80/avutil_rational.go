package ffmpeg

import (
	"encoding/json"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/rational.h>
#include <libavutil/avutil.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVRational C.AVRational
	AVRounding C.enum_AVRounding
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_ROUND_ZERO        AVRounding = C.AV_ROUND_ZERO        // Round toward zero.
	AV_ROUND_INF         AVRounding = C.AV_ROUND_INF         // Round away from zero.
	AV_ROUND_DOWN        AVRounding = C.AV_ROUND_DOWN        // Round toward -infinity.
	AV_ROUND_UP          AVRounding = C.AV_ROUND_UP          // Round toward +infinity.
	AV_ROUND_NEAR_INF    AVRounding = C.AV_ROUND_NEAR_INF    // Round to nearest and halfway cases away from zero.
	AV_ROUND_PASS_MINMAX AVRounding = C.AV_ROUND_PASS_MINMAX // Flag to pass INT64_MIN/MAX through instead of rescaling, this avoids special cases for AV_NOPTS_VALUE
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r AVRational) MarshalJSON() ([]byte, error) {
	if r.num == 0 {
		return json.Marshal(0)
	}
	return json.Marshal(fmt.Sprintf("%d/%d", r.num, r.den))
}

func (r AVRational) String() string {
	return marshalToString(r)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Create a new rational
func AVUtil_rational(num, den int) AVRational {
	return AVRational{num: C.int(num), den: C.int(den)}
}

// Numerator
func (r AVRational) Num() int {
	return int(r.num)
}

// Denominator
func (r AVRational) Den() int {
	return int(r.den)
}

// IsZero returns true if the rational is zero
func (r AVRational) IsZero() bool {
	return r.num == 0
}

// Float is used to convert an int64 value multipled by the rational to a float64
func (r AVRational) Float(multiplier int64) float64 {
	if r.den == 0 {
		return 0
	}
	return float64(int64(r.num)*multiplier) / float64(r.den)
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

// Convert a float64 to a rational.
func AVUtil_rational_d2q(d float64, max int) AVRational {
	if max == 0 {
		max = C.INT_MAX
	}
	return AVRational(C.av_d2q(C.double(d), C.int(max)))
}

// Convert an AVRational to a float64.
func AVUtil_rational_q2d(a AVRational) float64 {
	return float64(C.av_q2d(C.AVRational(a)))
}

// Compare two rationals.
func AVUtil_rational_equal(a, b AVRational) bool {
	return C.av_cmp_q(C.AVRational(a), C.AVRational(b)) == 0
}

// Invert a rational.
func AVUtil_rational_invert(q AVRational) AVRational {
	return AVRational(C.av_inv_q(C.AVRational(q)))
}

// Rescale a rational
func AVUtil_rational_rescale_q(a int64, bq AVRational, cq AVRational) int64 {
	return int64(C.av_rescale_q(C.int64_t(a), C.AVRational(bq), C.AVRational(cq)))
}

// Rescale a value from one range to another.
func AVUtil_rescale_rnd(a, b, c int64, rnd AVRounding) int64 {
	return int64(C.av_rescale_rnd(C.int64_t(a), C.int64_t(b), C.int64_t(c), C.enum_AVRounding(rnd)))
}

// Compare two timestamps each in its own time base. Returns -1 if a is before b, 1 if a is after b, or 0 if they are equal.
func AVUtil_compare_ts(a int64, a_tb AVRational, b int64, b_tb AVRational) int {
	return int(C.av_compare_ts(C.int64_t(a), C.AVRational(a_tb), C.int64_t(b), C.AVRational(b_tb)))
}
