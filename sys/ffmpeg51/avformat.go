package ffmpeg

import (
	"fmt"
	"unsafe"
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
	AVInputFormat  C.struct_AVInputFormat
	AVOutputFormat C.struct_AVOutputFormat
	AVFormatFlag   C.int
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AVFMT_NONE AVFormatFlag = 0
	// Demuxer will use avio_open, no opened file should be provided by the caller.
	AVFMT_NOFILE AVFormatFlag = C.AVFMT_NOFILE
	// Needs '%d' in filename.
	AVFMT_NEEDNUMBER AVFormatFlag = C.AVFMT_NEEDNUMBER
	// The muxer/demuxer is experimental and should be used with caution
	AVFMT_EXPERIMENTAL AVFormatFlag = C.AVFMT_EXPERIMENTAL
	// Show format stream IDs numbers.
	AVFMT_SHOWIDS AVFormatFlag = C.AVFMT_SHOW_IDS
	// Format wants global header.
	AVFMT_GLOBALHEADER AVFormatFlag = C.AVFMT_GLOBALHEADER
	// Format does not need / have any timestamps.
	AVFMT_NOTIMESTAMPS AVFormatFlag = C.AVFMT_NOTIMESTAMPS
	// Use generic index building code.
	AVFMT_GENERICINDEX AVFormatFlag = C.AVFMT_GENERIC_INDEX
	// Format allows timestamp discontinuities. Note, muxers always require valid (monotone) timestamps
	AVFMT_TSDISCONT AVFormatFlag = C.AVFMT_TS_DISCONT
	// Format allows variable fps.
	AVFMT_VARIABLEFPS AVFormatFlag = C.AVFMT_VARIABLE_FPS
	// Format does not need width/height
	AVFMT_NODIMENSIONS AVFormatFlag = C.AVFMT_NODIMENSIONS
	// Format does not require any streams
	AVFMT_NOSTREAMS AVFormatFlag = C.AVFMT_NOSTREAMS
	// Format does not allow to fall back on binary search via read_timestamp
	AVFMT_NOBINSEARCH AVFormatFlag = C.AVFMT_NOBINSEARCH
	// Format does not allow to fall back on generic search
	AVFMT_NOGENSEARCH AVFormatFlag = C.AVFMT_NOGENSEARCH
	// Format does not allow seeking by bytes
	AVFMT_NOBYTESEEK AVFormatFlag = C.AVFMT_NO_BYTE_SEEK
	// Format allows flushing. If not set, the muxer will not receive a NULL packet in the write_packet function.
	AVFMT_ALLOWFLUSH AVFormatFlag = C.AVFMT_ALLOW_FLUSH
	// Format does not require strictly increasing timestamps, but they must still be monotonic
	AVFMT_TS_NONSTRICT AVFormatFlag = C.AVFMT_TS_NONSTRICT
	// Format allows muxing negative timestamps
	AVFMT_TS_NEGATIVE AVFormatFlag = C.AVFMT_TS_NEGATIVE
	// Min
	AVFMT_MIN AVFormatFlag = AVFMT_NOFILE
	// Max
	AVFMT_MAX AVFormatFlag = AVFMT_TS_NEGATIVE
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - INPUT

func (this *AVInputFormat) Name() string {
	return C.GoString(this.name)
}

func (this *AVInputFormat) Description() string {
	return C.GoString(this.long_name)
}

func (this *AVInputFormat) Ext() string {
	return C.GoString(this.extensions)
}

func (this *AVInputFormat) MimeType() string {
	return C.GoString(this.mime_type)
}

func (this *AVInputFormat) Flags() AVFormatFlag {
	return AVFormatFlag(this.flags)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - OUTPUT

func (this *AVOutputFormat) Name() string {
	return C.GoString(this.name)
}

func (this *AVOutputFormat) Description() string {
	return C.GoString(this.long_name)
}

func (this *AVOutputFormat) Ext() string {
	return C.GoString(this.extensions)
}

func (this *AVOutputFormat) MimeType() string {
	return C.GoString(this.mime_type)
}

func (this *AVOutputFormat) Flags() AVFormatFlag {
	return AVFormatFlag(this.flags)
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *AVInputFormat) String() string {
	str := "<AVInputFormat"
	if name := this.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if description := this.Description(); description != "" {
		str += fmt.Sprintf(" description=%q", description)
	}
	if ext := this.Ext(); ext != "" {
		str += fmt.Sprintf(" ext=%q", ext)
	}
	if mimeType := this.MimeType(); mimeType != "" {
		str += fmt.Sprintf(" mime_type=%q", mimeType)
	}
	if flags := this.Flags(); flags != 0 {
		str += fmt.Sprint(" flags=", flags)
	}
	return str + ">"
}

func (this *AVOutputFormat) String() string {
	str := "<AVOutputFormat"
	if name := this.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if description := this.Description(); description != "" {
		str += fmt.Sprintf(" description=%q", description)
	}
	if ext := this.Ext(); ext != "" {
		str += fmt.Sprintf(" ext=%q", ext)
	}
	if mimeType := this.MimeType(); mimeType != "" {
		str += fmt.Sprintf(" mime_type=%q", mimeType)
	}
	if flags := this.Flags(); flags != 0 {
		str += fmt.Sprint(" flags=", flags)
	}
	return str + ">"
}

func (f AVFormatFlag) String() string {
	if f == AVFMT_NONE {
		return f.FlagString()
	}
	str := ""
	for i := AVFMT_MIN; i <= AVFMT_MAX; i <<= 1 {
		if f&i != 0 {
			str += "|" + i.FlagString()
		}
	}
	return str[1:]
}

func (f AVFormatFlag) FlagString() string {
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
		return "[?? Invalid AVFormatFlag value]"
	}
}
