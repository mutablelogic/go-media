package media

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Type int

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	NONE  Type = 0
	VIDEO Type = (1 << iota)
	AUDIO
	SUBTITLE
	DATA
	UNKNOWN
	ANY     = NONE
	mintype = VIDEO
	maxtype = UNKNOWN
)

///////////////////////////////////////////////////////////////////////////////
// STINGIFY

func (t Type) String() string {
	if t == NONE {
		return t.FlagString()
	}
	str := ""
	for f := mintype; f <= maxtype; f <<= 1 {
		if t&f == f {
			str += "|" + f.FlagString()
		}
	}
	return str[1:]
}

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
	default:
		return "UNKNOWN"
	}
}

///////////////////////////////////////////////////////////////////////////////
// METHODS

func (t Type) Is(u Type) bool {
	return t&u == u
}
