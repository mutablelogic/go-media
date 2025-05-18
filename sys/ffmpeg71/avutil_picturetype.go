package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
*/
import "C"

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
		return "AV_PICTURE_TYPE_NONE"
	case AV_PICTURE_TYPE_I:
		return "AV_PICTURE_TYPE_I"
	case AV_PICTURE_TYPE_P:
		return "AV_PICTURE_TYPE_P"
	case AV_PICTURE_TYPE_B:
		return "AV_PICTURE_TYPE_B"
	case AV_PICTURE_TYPE_S:
		return "AV_PICTURE_TYPE_S"
	case AV_PICTURE_TYPE_SI:
		return "AV_PICTURE_TYPE_SI"
	case AV_PICTURE_TYPE_SP:
		return "AV_PICTURE_TYPE_SP"
	case AV_PICTURE_TYPE_BI:
		return "AV_PICTURE_TYPE_BI"
	default:
		return "[?? Invalid AVPictureType value]"
	}
}
