package media

import (
	"encoding/json"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Type of codec, device, format or stream
type Type int

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	NONE     Type = 0           // Type is not defined
	VIDEO    Type = (1 << iota) // Type is video
	AUDIO                       // Type is audio
	SUBTITLE                    // Type is subtitle
	DATA                        // Type is data
	UNKNOWN                     // Type is unknown
	INPUT                       // Type is input format
	OUTPUT                      // Type is output format
	DEVICE                      // Type is input or output device
	maxtype
	mintype = VIDEO
	ANY     = NONE // Type is any (used for filtering)
)

///////////////////////////////////////////////////////////////////////////////
// STINGIFY

// Return the type as a string
func (t Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// Return the type as a string
func (t Type) String() string {
	if t == NONE {
		return t.FlagString()
	}
	str := ""
	for f := mintype; f < maxtype; f <<= 1 {
		if t&f == f {
			str += "|" + f.FlagString()
		}
	}
	return str[1:]
}

// Return a flag as a string
func (t Type) FlagString() string {
	switch t {
	case NONE:
		return "NONE"
	case VIDEO:
		return "VIDEO"
	case AUDIO:
		return "AUDIO"
	case SUBTITLE:
		return "SUBTITLE"
	case DATA:
		return "DATA"
	case INPUT:
		return "INPUT"
	case OUTPUT:
		return "OUTPUT"
	case DEVICE:
		return "DEVICE"
	default:
		return "UNKNOWN"
	}
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

// Returns true if the type matches a set of flags
func (t Type) Is(u Type) bool {
	return t&u == u
}
