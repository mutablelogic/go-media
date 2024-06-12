package ffmpeg

import "encoding/json"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVClass       C.AVClass
	AVError       C.int
	AVDictionary  C.struct_AVDictionary
	AVMediaType   C.enum_AVMediaType
	AVRational    C.AVRational
	AVPixelFormat C.enum_AVPixelFormat
)

type jsonAVClass struct {
	ClassName string `json:"class_name"`
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AVMEDIA_TYPE_UNKNOWN    AVMediaType = C.AVMEDIA_TYPE_UNKNOWN ///< Usually treated as AVMEDIA_TYPE_DATA
	AVMEDIA_TYPE_VIDEO      AVMediaType = C.AVMEDIA_TYPE_VIDEO
	AVMEDIA_TYPE_AUDIO      AVMediaType = C.AVMEDIA_TYPE_AUDIO
	AVMEDIA_TYPE_DATA       AVMediaType = C.AVMEDIA_TYPE_DATA ///< Opaque data information usually continuous
	AVMEDIA_TYPE_SUBTITLE   AVMediaType = C.AVMEDIA_TYPE_SUBTITLE
	AVMEDIA_TYPE_ATTACHMENT AVMediaType = C.AVMEDIA_TYPE_ATTACHMENT ///< Opaque data information usually sparse
)

////////////////////////////////////////////////////////////////////////////////
// JSON OUTPUT

func (ctx AVClass) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVClass{
		ClassName: C.GoString(ctx.class_name),
	})
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx AVClass) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

func (v AVMediaType) String() string {
	switch v {
	case AVMEDIA_TYPE_UNKNOWN:
		return "AVMEDIA_TYPE_UNKNOWN"
	case AVMEDIA_TYPE_VIDEO:
		return "AVMEDIA_TYPE_VIDEO"
	case AVMEDIA_TYPE_AUDIO:
		return "AVMEDIA_TYPE_AUDIO"
	case AVMEDIA_TYPE_DATA:
		return "AVMEDIA_TYPE_DATA"
	case AVMEDIA_TYPE_SUBTITLE:
		return "AVMEDIA_TYPE_SUBTITLE"
	case AVMEDIA_TYPE_ATTACHMENT:
		return "AVMEDIA_TYPE_ATTACHMENT"
	}
	return "[AVMediaType]"
}
