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
	AVSeekFlag C.int
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	/**
	* ORing this as the "whence" parameter to a seek function causes it to
	* return the filesize without seeking anywhere. Supporting this is optional.
	* If it is not supported then the seek function will return <0.
	 */
	AVSEEK_SIZE = C.AVSEEK_SIZE

	/**
	 * Passing this flag as the "whence" parameter to a seek function causes it to
	 * seek by any means (like reopening and linear reading) or other normally unreasonable
	 * means that can be extremely slow.
	 * This may be ignored by the seek code.
	 */
	AVSEEK_FORCE = C.AVSEEK_FORCE
)

const (
	AVSEEK_FLAG_NONE     AVSeekFlag = 0                      ///< no special flags
	AVSEEK_FLAG_BACKWARD AVSeekFlag = C.AVSEEK_FLAG_BACKWARD ///< seek backward
	AVSEEK_FLAG_BYTE     AVSeekFlag = C.AVSEEK_FLAG_BYTE     ///< seek by byte
	AVSEEK_FLAG_ANY      AVSeekFlag = C.AVSEEK_FLAG_ANY      ///< seek to any frame
	AVSEEK_FLAG_FRAME    AVSeekFlag = C.AVSEEK_FLAG_FRAME    ///< seek to frame
	AVSEEK_FLAG_MIN                 = AVSEEK_FLAG_BACKWARD
	AVSEEK_FLAG_MAX                 = AVSEEK_FLAG_FRAME
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f AVSeekFlag) FlagString() string {
	switch f {
	case AVSEEK_FLAG_NONE:
		return ""
	case AVSEEK_FLAG_BACKWARD:
		return "AVSEEK_FLAG_BACKWARD"
	case AVSEEK_FLAG_BYTE:
		return "AVSEEK_FLAG_BYTE"
	case AVSEEK_FLAG_ANY:
		return "AVSEEK_FLAG_ANY"
	case AVSEEK_FLAG_FRAME:
		return "AVSEEK_FLAG_FRAME"
	default:
		return fmt.Sprintf("AVSeekFlag(0x%04X)", int(f))
	}
}

func (f AVSeekFlag) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func (f AVSeekFlag) String() string {
	if f == AVSEEK_FLAG_NONE {
		return "AVSEEK_FLAG_NONE"
	}
	str := ""
	for i := AVSEEK_FLAG_MIN; i <= AVSEEK_FLAG_MAX; i <<= 1 {
		if f&i != 0 {
			str += "|" + i.FlagString()
		}
	}
	return str[1:]
}

func (f AVSeekFlag) Is(flag AVSeekFlag) bool {
	return f&flag == flag
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Seek to timestamp ts.
func AVFormat_seek_frame(ctx *AVFormatContext, stream int, ts int64, flags AVSeekFlag) error {
	if err := AVError(C.av_seek_frame((*C.struct_AVFormatContext)(ctx), C.int(stream), C.int64_t(ts), C.int(flags))); err != 0 {
		return err
	}
	return nil
}

// Seek to the keyframe at timestamp.
func AVFormat_seek_file(ctx *AVFormatContext, stream int, min_ts, ts, max_ts int64, flags AVSeekFlag) error {
	if err := AVError(C.avformat_seek_file((*C.struct_AVFormatContext)(ctx), C.int(stream), C.int64_t(min_ts), C.int64_t(ts), C.int64_t(max_ts), C.int(flags))); err != 0 {
		return err
	}
	return nil
}
