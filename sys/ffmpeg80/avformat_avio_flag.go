package ffmpeg

import (
	"encoding/json"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avio.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVIOFlag C.int
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AVIO_FLAG_NONE       AVIOFlag = 0
	AVIO_FLAG_READ       AVIOFlag = C.AVIO_FLAG_READ                     ///< read-only
	AVIO_FLAG_WRITE      AVIOFlag = C.AVIO_FLAG_WRITE                    ///< write-only
	AVIO_FLAG_READ_WRITE AVIOFlag = C.AVIO_FLAG_READ | C.AVIO_FLAG_WRITE ///< read-write pseudo flag
	AVIO_FLAG_NONBLOCK   AVIOFlag = C.AVIO_FLAG_NONBLOCK                 ///< Use non-blocking mode
	AVIO_FLAG_DIRECT     AVIOFlag = C.AVIO_FLAG_DIRECT                   ///< Use direct mode
	AVIO_FLAG_MIN        AVIOFlag = AVIO_FLAG_READ
	AVIO_FLAG_MAX        AVIOFlag = AVIO_FLAG_DIRECT
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f AVIOFlag) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func (f AVIOFlag) String() string {
	if f == AVIO_FLAG_NONE {
		return f.FlagString()
	}
	str := ""
	for i := AVIO_FLAG_MIN; i <= AVIO_FLAG_MAX; i <<= 1 {
		if f&i == i {
			str += "|" + i.FlagString()
		}
	}
	if str != "" {
		str = str[1:]
	}
	return str
}

func (f AVIOFlag) FlagString() string {
	switch f {
	case AVIO_FLAG_NONE:
		return "AVIO_FLAG_NONE"
	case AVIO_FLAG_READ:
		return "AVIO_FLAG_READ"
	case AVIO_FLAG_WRITE:
		return "AVIO_FLAG_WRITE"
	case AVIO_FLAG_READ_WRITE:
		return "AVIO_FLAG_READ_WRITE"
	case AVIO_FLAG_NONBLOCK:
		return "AVIO_FLAG_NONBLOCK"
	case AVIO_FLAG_DIRECT:
		return "AVIO_FLAG_DIRECT"
	default:
		return fmt.Sprintf("AVIOFlag(0x%04X)", int(f))
	}
}

func (f AVIOFlag) Is(flag AVIOFlag) bool {
	return f&flag != 0
}
