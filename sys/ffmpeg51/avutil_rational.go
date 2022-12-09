package ffmpeg

import (
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

func (r AVRational) String() string {
	if r.Num() == 0 {
		return "0"
	} else {
		return fmt.Sprintf("<AVRational>{ num=%v den=%v }", r.Num(), r.Den())
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Numerator
func (r AVRational) Num() int {
	return int(r.num)
}

// Denominator
func (r AVRational) Den() int {
	return int(r.den)
}

// Float is used to convert an int64 value multipled by the rational to a float64
func (r AVRational) Float(multiplier int64) float64 {
	return float64(int64(r.num)*multiplier) / float64(r.den)
}
