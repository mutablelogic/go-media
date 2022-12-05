package chromaprint

import (
	"strings"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Meta uint

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	META_RECORDING Meta = (1 << iota)
	META_RECORDINGID
	META_RELEASE
	META_RELEASEID
	META_RELEASEGROUP
	META_RELEASEGROUPID
	META_TRACK
	META_COMPRESS
	META_USERMETA
	META_SOURCE
	META_MIN       = META_RECORDING
	META_MAX       = META_SOURCE
	META_NONE Meta = 0
	META_ALL       = META_RECORDING | META_RECORDINGID | META_RELEASE | META_RELEASEID | META_RELEASEGROUP | META_RELEASEGROUPID | META_TRACK | META_COMPRESS | META_USERMETA | META_SOURCE
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m Meta) String() string {
	if m == META_NONE {
		return m.FlagString()
	}
	str := ""
	for v := META_MIN; v <= META_MAX; v <<= 1 {
		if m&v == v {
			str += v.FlagString() + " "
		}
	}
	return strings.TrimSuffix(str, " ")
}

func (m Meta) FlagString() string {
	switch m {
	case META_NONE:
		return ""
	case META_RECORDING:
		return "recordings"
	case META_RECORDINGID:
		return "recordingids"
	case META_RELEASE:
		return "releases"
	case META_RELEASEID:
		return "releaseids"
	case META_RELEASEGROUP:
		return "releasegroups"
	case META_RELEASEGROUPID:
		return "releasegroupids"
	case META_TRACK:
		return "tracks"
	case META_COMPRESS:
		return "compress"
	case META_USERMETA:
		return "usermeta"
	case META_SOURCE:
		return "sources"
	default:
		return "[?? Invalid Meta value]"
	}
}
