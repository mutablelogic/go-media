package media

import "encoding/json"

////////////////////////////////////////////////////////////////////////////
// TYPES

// Media type flags
type MediaType uint32

////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	NONE     MediaType = 0
	UNKNOWN  MediaType = (1 << iota) // Usually treated as DATA
	VIDEO                            // Video stream
	AUDIO                            // Audio stream
	DATA                             // Opaque data information usually continuous
	SUBTITLE                         // Subtitle stream
	INPUT                            // Demuxer
	OUTPUT                           // Muxer
	FILE                             // File or byte stream
	DEVICE                           // Device rather than stream
	CODEC                            // Codec

	// Set minimum and maximum values
	MIN = UNKNOWN
	MAX = CODEC

	// Convenience values
	ANY = NONE
)

////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v MediaType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v MediaType) String() string {
	if v == NONE {
		return v.FlagString()
	}
	str := ""
	for f := MIN; f <= MAX; f <<= 1 {
		if v&f == f {
			str += "|" + f.FlagString()
		}
	}
	return str[1:]
}

func (v MediaType) FlagString() string {
	switch v {
	case NONE:
		return "NONE"
	case VIDEO:
		return "VIDEO"
	case AUDIO:
		return "AUDIO"
	case DATA:
		return "DATA"
	case SUBTITLE:
		return "SUBTITLE"
	case INPUT:
		return "INPUT"
	case OUTPUT:
		return "OUTPUT"
	case FILE:
		return "FILE"
	case DEVICE:
		return "DEVICE"
	case CODEC:
		return "CODEC"
	}
	return ""
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (v MediaType) Is(f MediaType) bool {
	return v&f == f
}
