package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libswscale
#include <libswscale/swscale.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	SWSContext C.struct_SwsContext
	SWSFilter  C.struct_SwsFilter
	SWSFlag    C.int
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	SWS_NONE          SWSFlag = 0
	SWS_FAST_BILINEAR SWSFlag = C.SWS_FAST_BILINEAR
	SWS_BILINEAR      SWSFlag = C.SWS_BILINEAR
	SWS_BICUBIC       SWSFlag = C.SWS_BICUBIC
	SWS_X             SWSFlag = C.SWS_X
	SWS_POINT         SWSFlag = C.SWS_POINT
	SWS_AREA          SWSFlag = C.SWS_AREA
	SWS_BICUBLIN      SWSFlag = C.SWS_BICUBLIN
	SWS_GAUSS         SWSFlag = C.SWS_GAUSS
	SWS_SINC          SWSFlag = C.SWS_SINC
	SWS_LANCZOS       SWSFlag = C.SWS_LANCZOS
	SWS_SPLINE        SWSFlag = C.SWS_SPLINE
	SWS_MIN                   = SWS_FAST_BILINEAR
	SWS_MAX                   = SWS_SPLINE
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v SWSFlag) FlagString() string {
	switch v {
	case SWS_NONE:
		return "SWS_NONE"
	case SWS_FAST_BILINEAR:
		return "SWS_FAST_BILINEAR"
	case SWS_BILINEAR:
		return "SWS_BILINEAR"
	case SWS_BICUBIC:
		return "SWS_BICUBIC"
	case SWS_X:
		return "SWS_X"
	case SWS_POINT:
		return "SWS_POINT"
	case SWS_AREA:
		return "SWS_AREA"
	case SWS_BICUBLIN:
		return "SWS_BICUBLIN"
	case SWS_GAUSS:
		return "SWS_GAUSS"
	case SWS_SINC:
		return "SWS_SINC"
	case SWS_LANCZOS:
		return "SWS_LANCZOS"
	case SWS_SPLINE:
		return "SWS_SPLINE"
	default:
		return "[?? Invalid SWSFlag value]"
	}
}

func (v SWSFlag) String() string {
	if v == SWS_NONE {
		return v.FlagString()
	}
	str := ""
	for i := SWS_MIN; i <= SWS_MAX; i <<= 1 {
		if v&i != 0 {
			str += "|" + i.FlagString()
		}
	}
	return str[1:]
}

func (ctx *SWSContext) String() string {
	if ctx == nil {
		return "<nil>"
	}
	return "<SWSContext>"
}

func (filter *SWSFilter) String() string {
	if filter == nil {
		return "<nil>"
	}
	return "<SWSFilter>"
}
