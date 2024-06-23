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
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r AVRational) MarshalJSON() ([]byte, error) {
	if r.num == 0 {
		return json.Marshal(0)
	}
	return json.Marshal(fmt.Sprintf("%d/%d", r.num, r.den))
}

func (r AVRational) String() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
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
	return float64(int64(r.num)*multiplier) / float64(r.den)
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

// Convert a double precision floating point number to a rational.
func AVUtil_rational_d2q(d float64, max int) AVRational {
	if max == 0 {
		max = C.INT_MAX
	}
	return AVRational(C.av_d2q(C.double(d), C.int(max)))
}

// Convert an AVRational to a double.
func AVUtil_q2d(a AVRational) float64 {
	return float64(C.av_q2d(C.AVRational(a)))
}
