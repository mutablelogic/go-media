package ffmpeg

import (
	"encoding/json"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVFormatFlag C.int
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AVFMT_FLAG_NONE            AVFormatFlag = 0
	AVFMT_FLAG_GENPTS          AVFormatFlag = C.AVFMT_FLAG_GENPTS          ///< Generate missing pts even if it requires parsing future frames.
	AVFMT_FLAG_IGNIDX          AVFormatFlag = C.AVFMT_FLAG_IGNIDX          ///< Ignore index.
	AVFMT_FLAG_NONBLOCK        AVFormatFlag = C.AVFMT_FLAG_NONBLOCK        ///< Do not block when reading packets from input.
	AVFMT_FLAG_IGNDTS          AVFormatFlag = C.AVFMT_FLAG_IGNDTS          ///< Ignore DTS on frames that contain both DTS & PTS
	AVFMT_FLAG_NOFILLIN        AVFormatFlag = C.AVFMT_FLAG_NOFILLIN        ///< Do not infer any values from other values, just return what is stored in the container
	AVFMT_FLAG_NOPARSE         AVFormatFlag = C.AVFMT_FLAG_NOPARSE         ///< Do not use AVParsers, you also must set AVFMT_FLAG_NOFILLIN as the fillin code works on frames and no parsing -> no frames. Also seeking to frames can not work if parsing to find frame boundaries has been disabled
	AVFMT_FLAG_NOBUFFER        AVFormatFlag = C.AVFMT_FLAG_NOBUFFER        ///< Do not buffer frames when possible
	AVFMT_FLAG_CUSTOM_IO       AVFormatFlag = C.AVFMT_FLAG_CUSTOM_IO       ///< The caller has supplied a custom AVIOContext, don't avio_close() it.
	AVFMT_FLAG_DISCARD_CORRUPT AVFormatFlag = C.AVFMT_FLAG_DISCARD_CORRUPT ///< Discard frames marked corrupted
	AVFMT_FLAG_FLUSH_PACKETS   AVFormatFlag = C.AVFMT_FLAG_FLUSH_PACKETS   ///< Flush the AVIOContext every packet.
	AVFMT_FLAG_BITEXACT        AVFormatFlag = C.AVFMT_FLAG_BITEXACT        // When muxing, try to avoid writing any random/volatile data to the output.
	AVFMT_FLAG_SORT_DTS        AVFormatFlag = C.AVFMT_FLAG_SORT_DTS        ///< try to interleave outputted packets by dts (using this flag can slow demuxing down)
	AVFMT_FLAG_FAST_SEEK       AVFormatFlag = C.AVFMT_FLAG_FAST_SEEK       ///< Enable fast, but inaccurate seeks for some formats
	AVFMT_FLAG_AUTO_BSF        AVFormatFlag = C.AVFMT_FLAG_AUTO_BSF        ///< Add bitstream filters as requested by the muxer
	AVFMT_FLAG_MIN                          = AVFMT_FLAG_GENPTS
	AVFMT_FLAG_MAX                          = AVFMT_FLAG_AUTO_BSF
)

func (f AVFormatFlag) FlagString() string {
	switch f {
	case AVFMT_FLAG_NONE:
		return "AVFMT_FLAG_NONE"
	case AVFMT_FLAG_GENPTS:
		return "AVFMT_FLAG_GENPTS"
	case AVFMT_FLAG_IGNIDX:
		return "AVFMT_FLAG_IGNIDX"
	case AVFMT_FLAG_NONBLOCK:
		return "AVFMT_FLAG_NONBLOCK"
	case AVFMT_FLAG_IGNDTS:
		return "AVFMT_FLAG_IGNDTS"
	case AVFMT_FLAG_NOFILLIN:
		return "AVFMT_FLAG_NOFILLIN"
	case AVFMT_FLAG_NOPARSE:
		return "AVFMT_FLAG_NOPARSE"
	case AVFMT_FLAG_NOBUFFER:
		return "AVFMT_FLAG_NOBUFFER"
	case AVFMT_FLAG_CUSTOM_IO:
		return "AVFMT_FLAG_CUSTOM_IO"
	case AVFMT_FLAG_DISCARD_CORRUPT:
		return "AVFMT_FLAG_DISCARD_CORRUPT"
	case AVFMT_FLAG_FLUSH_PACKETS:
		return "AVFMT_FLAG_FLUSH_PACKETS"
	case AVFMT_FLAG_BITEXACT:
		return "AVFMT_FLAG_BITEXACT"
	case AVFMT_FLAG_SORT_DTS:
		return "AVFMT_FLAG_SORT_DTS"
	case AVFMT_FLAG_FAST_SEEK:
		return "AVFMT_FLAG_FAST_SEEK"
	case AVFMT_FLAG_AUTO_BSF:
		return "AVFMT_FLAG_AUTO_BSF"
	default:
		return fmt.Sprintf("AVFormatFlag(0x%06X)", int(f))
	}
}

func (f AVFormatFlag) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func (f AVFormatFlag) String() string {
	if f == AVFMT_FLAG_NONE {
		return f.FlagString()
	}
	str := ""
	for i := AVFMT_FLAG_MIN; i <= AVFMT_FLAG_MAX; i <<= 1 {
		if f&i != 0 {
			str += "|" + i.FlagString()
		}
	}
	return str[1:]
}

func (f AVFormatFlag) Is(flag AVFormatFlag) bool {
	return f&flag == flag
}
