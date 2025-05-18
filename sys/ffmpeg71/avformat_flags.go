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

const (
	AVFMT_NONE         AVFormat = 0
	AVFMT_NOFILE       AVFormat = C.AVFMT_NOFILE        // Demuxer will use avio_open, no opened file should be provided by the caller.
	AVFMT_NEEDNUMBER   AVFormat = C.AVFMT_NEEDNUMBER    // Needs '%d' in filename.
	AVFMT_EXPERIMENTAL AVFormat = C.AVFMT_EXPERIMENTAL  // The muxer/demuxer is experimental and should be used with caution
	AVFMT_SHOWIDS      AVFormat = C.AVFMT_SHOW_IDS      // Show format stream IDs numbers.
	AVFMT_GLOBALHEADER AVFormat = C.AVFMT_GLOBALHEADER  // Format wants global header.
	AVFMT_NOTIMESTAMPS AVFormat = C.AVFMT_NOTIMESTAMPS  // Format does not need / have any timestamps.
	AVFMT_GENERICINDEX AVFormat = C.AVFMT_GENERIC_INDEX // Use generic index building code.
	AVFMT_TSDISCONT    AVFormat = C.AVFMT_TS_DISCONT    // Format allows timestamp discontinuities. Note, muxers always require valid (monotone) timestamps
	AVFMT_VARIABLEFPS  AVFormat = C.AVFMT_VARIABLE_FPS  // Format allows variable fps.
	AVFMT_NODIMENSIONS AVFormat = C.AVFMT_NODIMENSIONS  // Format does not need width/height
	AVFMT_NOSTREAMS    AVFormat = C.AVFMT_NOSTREAMS     // Format does not require any streams
	AVFMT_NOBINSEARCH  AVFormat = C.AVFMT_NOBINSEARCH   // Format does not allow to fall back on binary search via read_timestamp
	AVFMT_NOGENSEARCH  AVFormat = C.AVFMT_NOGENSEARCH   // Format does not allow to fall back on generic search
	AVFMT_NOBYTESEEK   AVFormat = C.AVFMT_NO_BYTE_SEEK  // Format does not allow seeking by bytes
	AVFMT_ALLOWFLUSH   AVFormat = C.AVFMT_ALLOW_FLUSH   // Format allows flushing. If not set, the muxer will not receive a NULL packet in the write_packet function.
	AVFMT_TS_NONSTRICT AVFormat = C.AVFMT_TS_NONSTRICT  // Format does not require strictly increasing timestamps, but they must still be monotonic
	AVFMT_TS_NEGATIVE  AVFormat = C.AVFMT_TS_NEGATIVE   // Format allows muxing negative timestamps
	AVFMT_SEEK_TO_PTS  AVFormat = C.AVFMT_SEEK_TO_PTS   // Seeking is based on PTS
	AVFMT_MIN          AVFormat = AVFMT_NOFILE
	AVFMT_MAX          AVFormat = AVFMT_SEEK_TO_PTS
)

////////////////////////////////////////////////////////////////////////////////
// AVFormat

func (f AVFormat) Is(flag AVFormat) bool {
	return f&flag != 0
}

func (v AVFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v AVFormat) String() string {
	if v == AVFMT_NONE {
		return v.FlagString()
	}
	str := ""
	for i := AVFMT_MIN; i <= AVFMT_MAX; i <<= 1 {
		if v&i == i {
			str += "|" + i.FlagString()
		}
	}
	if str != "" {
		str = str[1:]
	}
	return str
}

func (f AVFormat) FlagString() string {
	switch f {
	case AVFMT_NONE:
		return "AVFMT_NONE"
	case AVFMT_NOFILE:
		return "AVFMT_NOFILE"
	case AVFMT_NEEDNUMBER:
		return "AVFMT_NEEDNUMBER"
	case AVFMT_EXPERIMENTAL:
		return "AVFMT_EXPERIMENTAL"
	case AVFMT_SHOWIDS:
		return "AVFMT_SHOWIDS"
	case AVFMT_GLOBALHEADER:
		return "AVFMT_GLOBALHEADER"
	case AVFMT_NOTIMESTAMPS:
		return "AVFMT_NOTIMESTAMPS"
	case AVFMT_GENERICINDEX:
		return "AVFMT_GENERICINDEX"
	case AVFMT_TSDISCONT:
		return "AVFMT_TSDISCONT"
	case AVFMT_VARIABLEFPS:
		return "AVFMT_VARIABLEFPS"
	case AVFMT_NODIMENSIONS:
		return "AVFMT_NODIMENSIONS"
	case AVFMT_NOSTREAMS:
		return "AVFMT_NOSTREAMS"
	case AVFMT_NOBINSEARCH:
		return "AVFMT_NOBINSEARCH"
	case AVFMT_NOGENSEARCH:
		return "AVFMT_NOGENSEARCH"
	case AVFMT_NOBYTESEEK:
		return "AVFMT_NOBYTESEEK"
	case AVFMT_ALLOWFLUSH:
		return "AVFMT_ALLOWFLUSH"
	case AVFMT_TS_NONSTRICT:
		return "AVFMT_TS_NONSTRICT"
	case AVFMT_TS_NEGATIVE:
		return "AVFMT_TS_NEGATIVE"
	default:
		return fmt.Sprintf("AVFormat(0x%08X)", uint32(f))
	}
}
