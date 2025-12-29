package ffmpeg

import (
	"encoding/json"
	"fmt"
)

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
	AVPictureType C.enum_AVPictureType
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_PICTURE_TYPE_NONE AVPictureType = C.AV_PICTURE_TYPE_NONE ///< Undefined
	AV_PICTURE_TYPE_I    AVPictureType = C.AV_PICTURE_TYPE_I    ///< Intra
	AV_PICTURE_TYPE_P    AVPictureType = C.AV_PICTURE_TYPE_P    ///< Predicted
	AV_PICTURE_TYPE_B    AVPictureType = C.AV_PICTURE_TYPE_B    ///< Bi-dir predicted
	AV_PICTURE_TYPE_S    AVPictureType = C.AV_PICTURE_TYPE_S    ///< S(GMC)-VOP MPEG-4
	AV_PICTURE_TYPE_SI   AVPictureType = C.AV_PICTURE_TYPE_SI   ///< Switching Intra
	AV_PICTURE_TYPE_SP   AVPictureType = C.AV_PICTURE_TYPE_SP   ///< Switching Predicted
	AV_PICTURE_TYPE_BI   AVPictureType = C.AV_PICTURE_TYPE_BI   ///< BI type
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v AVPictureType) String() string {
	switch v {
	case AV_PICTURE_TYPE_NONE:
		return "NONE"
	case AV_PICTURE_TYPE_I:
		return "I"
	case AV_PICTURE_TYPE_P:
		return "P"
	case AV_PICTURE_TYPE_B:
		return "B"
	case AV_PICTURE_TYPE_S:
		return "S"
	case AV_PICTURE_TYPE_SI:
		return "SI"
	case AV_PICTURE_TYPE_SP:
		return "SP"
	case AV_PICTURE_TYPE_BI:
		return "BI"
	default:
		return fmt.Sprintf("AVPictureType(%d)", int(v))
	}
}

func (v AVPictureType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Return a single character representing the picture type
func AVUtil_get_picture_type_char(pict_type AVPictureType) rune {
	return rune(C.av_get_picture_type_char(C.enum_AVPictureType(pict_type)))
}
