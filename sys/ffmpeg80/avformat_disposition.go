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
	AVDisposition C.int
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_DISPOSITION_DEFAULT          AVDisposition = C.AV_DISPOSITION_DEFAULT
	AV_DISPOSITION_DUB              AVDisposition = C.AV_DISPOSITION_DUB
	AV_DISPOSITION_ORIGINAL         AVDisposition = C.AV_DISPOSITION_ORIGINAL
	AV_DISPOSITION_COMMENT          AVDisposition = C.AV_DISPOSITION_COMMENT
	AV_DISPOSITION_LYRICS           AVDisposition = C.AV_DISPOSITION_LYRICS
	AV_DISPOSITION_KARAOKE          AVDisposition = C.AV_DISPOSITION_KARAOKE
	AV_DISPOSITION_FORCED           AVDisposition = C.AV_DISPOSITION_FORCED
	AV_DISPOSITION_HEARING_IMPAIRED AVDisposition = C.AV_DISPOSITION_HEARING_IMPAIRED
	AV_DISPOSITION_VISUAL_IMPAIRED  AVDisposition = C.AV_DISPOSITION_VISUAL_IMPAIRED
	AV_DISPOSITION_CLEAN_EFFECTS    AVDisposition = C.AV_DISPOSITION_CLEAN_EFFECTS
	AV_DISPOSITION_ATTACHED_PIC     AVDisposition = C.AV_DISPOSITION_ATTACHED_PIC
	AV_DISPOSITION_TIMED_THUMBNAILS AVDisposition = C.AV_DISPOSITION_TIMED_THUMBNAILS
	AV_DISPOSITION_NON_DIEGETIC     AVDisposition = C.AV_DISPOSITION_NON_DIEGETIC
	AV_DISPOSITION_CAPTIONS         AVDisposition = C.AV_DISPOSITION_CAPTIONS
	AV_DISPOSITION_DESCRIPTIONS     AVDisposition = C.AV_DISPOSITION_DESCRIPTIONS
	AV_DISPOSITION_METADATA         AVDisposition = C.AV_DISPOSITION_METADATA
	AV_DISPOSITION_DEPENDENT        AVDisposition = C.AV_DISPOSITION_DEPENDENT
	AV_DISPOSITION_STILL_IMAGE      AVDisposition = C.AV_DISPOSITION_STILL_IMAGE
	AV_DISPOSITION_MULTILAYER       AVDisposition = C.AV_DISPOSITION_MULTILAYER
	AV_DISPOSITION_MIN                            = AV_DISPOSITION_DEFAULT
	AV_DISPOSITION_MAX                            = AV_DISPOSITION_MULTILAYER
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v AVDisposition) String() string {
	if v == 0 {
		return ""
	}
	str := ""
	for f := AV_DISPOSITION_MIN; f <= AV_DISPOSITION_MAX; f <<= 1 {
		if v&f != 0 {
			str += "|" + f.FlagString()
		}
	}
	return str[1:]
}

func (v AVDisposition) FlagString() string {
	switch v {
	case AV_DISPOSITION_DEFAULT:
		return "DEFAULT"
	case AV_DISPOSITION_DUB:
		return "DUB"
	case AV_DISPOSITION_ORIGINAL:
		return "ORIGINAL"
	case AV_DISPOSITION_COMMENT:
		return "COMMENT"
	case AV_DISPOSITION_LYRICS:
		return "LYRICS"
	case AV_DISPOSITION_KARAOKE:
		return "KARAOKE"
	case AV_DISPOSITION_FORCED:
		return "FORCED"
	case AV_DISPOSITION_HEARING_IMPAIRED:
		return "HEARING_IMPAIRED"
	case AV_DISPOSITION_VISUAL_IMPAIRED:
		return "VISUAL_IMPAIRED"
	case AV_DISPOSITION_CLEAN_EFFECTS:
		return "CLEAN_EFFECTS"
	case AV_DISPOSITION_ATTACHED_PIC:
		return "ATTACHED_PIC"
	case AV_DISPOSITION_TIMED_THUMBNAILS:
		return "TIMED_THUMBNAILS"
	case AV_DISPOSITION_NON_DIEGETIC:
		return "NON_DIEGETIC"
	case AV_DISPOSITION_CAPTIONS:
		return "CAPTIONS"
	case AV_DISPOSITION_DESCRIPTIONS:
		return "DESCRIPTIONS"
	case AV_DISPOSITION_METADATA:
		return "METADATA"
	case AV_DISPOSITION_DEPENDENT:
		return "DEPENDENT"
	case AV_DISPOSITION_STILL_IMAGE:
		return "STILL_IMAGE"
	case AV_DISPOSITION_MULTILAYER:
		return "MULTILAYER"
	default:
		return fmt.Sprintf("AVDisposition(0x%08X)", int(v))
	}
}

////////////////////////////////////////////////////////////////////////////////
// AVDisposition

func (v AVDisposition) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (f AVDisposition) Is(flag AVDisposition) bool {
	return f&flag == flag
}
