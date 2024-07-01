package ffmpeg

import (
	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Type ff.AVMediaType

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	UNKNOWN  Type = Type(ff.AVMEDIA_TYPE_UNKNOWN)
	VIDEO    Type = Type(ff.AVMEDIA_TYPE_VIDEO)
	AUDIO    Type = Type(ff.AVMEDIA_TYPE_AUDIO)
	DATA     Type = Type(ff.AVMEDIA_TYPE_DATA)
	SUBTITLE Type = Type(ff.AVMEDIA_TYPE_SUBTITLE)
)

///////////////////////////////////////////////////////////////////////////////
// STINGIFY

func (t Type) String() string {
	switch t {
	case VIDEO:
		return "VIDEO"
	case AUDIO:
		return "AUDIO"
	case DATA:
		return "DATA"
	case SUBTITLE:
		return "SUBTITLE"
	default:
		return "UNKNOWN"
	}
}
