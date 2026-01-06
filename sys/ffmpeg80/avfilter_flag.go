package ffmpeg

import (
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavfilter
#include <libavfilter/avfilter.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type AVFilterFlag C.int

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	AVFILTER_FLAG_NONE                      AVFilterFlag = 0
	AVFILTER_FLAG_DYNAMIC_INPUTS            AVFilterFlag = C.AVFILTER_FLAG_DYNAMIC_INPUTS
	AVFILTER_FLAG_DYNAMIC_OUTPUTS           AVFilterFlag = C.AVFILTER_FLAG_DYNAMIC_OUTPUTS
	AVFILTER_FLAG_SLICE_THREADS             AVFilterFlag = C.AVFILTER_FLAG_SLICE_THREADS
	AVFILTER_FLAG_METADATA_ONLY             AVFilterFlag = C.AVFILTER_FLAG_METADATA_ONLY
	AVFILTER_FLAG_HWDEVICE                  AVFilterFlag = C.AVFILTER_FLAG_HWDEVICE
	AVFILTER_FLAG_SUPPORT_TIMELINE_GENERIC  AVFilterFlag = C.AVFILTER_FLAG_SUPPORT_TIMELINE_GENERIC
	AVFILTER_FLAG_SUPPORT_TIMELINE_INTERNAL AVFilterFlag = C.AVFILTER_FLAG_SUPPORT_TIMELINE_INTERNAL
	AVFILTER_FLAG_SUPPORT_TIMELINE          AVFilterFlag = AVFILTER_FLAG_SUPPORT_TIMELINE_GENERIC | AVFILTER_FLAG_SUPPORT_TIMELINE_INTERNAL
	AVFILTER_FLAG_MAX                                    = AVFILTER_FLAG_SUPPORT_TIMELINE_INTERNAL
)

////////////////////////////////////////////////////////////////////////////////
// AVFilterFlag

func (v AVFilterFlag) Is(f AVFilterFlag) bool {
	return v&f == f
}

func (v AVFilterFlag) String() string {
	if v == AVFILTER_FLAG_NONE {
		return v.FlagString()
	}
	str := ""
	for i := AVFilterFlag(C.int(1)); i <= AVFILTER_FLAG_MAX; i <<= 1 {
		if v&i == i {
			str += "|" + i.FlagString()
		}
	}
	return str[1:]
}

func (v AVFilterFlag) FlagString() string {
	switch v {
	case AVFILTER_FLAG_NONE:
		return "AVFILTER_FLAG_NONE"
	case AVFILTER_FLAG_DYNAMIC_INPUTS:
		return "AVFILTER_FLAG_DYNAMIC_INPUTS"
	case AVFILTER_FLAG_DYNAMIC_OUTPUTS:
		return "AVFILTER_FLAG_DYNAMIC_OUTPUTS"
	case AVFILTER_FLAG_SLICE_THREADS:
		return "AVFILTER_FLAG_SLICE_THREADS"
	case AVFILTER_FLAG_METADATA_ONLY:
		return "AVFILTER_FLAG_METADATA_ONLY"
	case AVFILTER_FLAG_HWDEVICE:
		return "AVFILTER_FLAG_HWDEVICE"
	case AVFILTER_FLAG_SUPPORT_TIMELINE:
		return "AVFILTER_FLAG_SUPPORT_TIMELINE"
	case AVFILTER_FLAG_SUPPORT_TIMELINE_GENERIC:
		return "AVFILTER_FLAG_SUPPORT_TIMELINE_GENERIC"
	case AVFILTER_FLAG_SUPPORT_TIMELINE_INTERNAL:
		return "AVFILTER_FLAG_SUPPORT_TIMELINE_INTERNAL"
	default:
		return fmt.Sprintf("AVFilterFlag(0x%08X)", uint32(v))
	}
}
