package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the LIBAVCODEC_VERSION_INT constant.
func AVCodec_version() uint {
	return uint(C.avcodec_version())
}

// Return the libavcodec build-time configuration.
func AVCodec_configuration() string {
	return C.GoString(C.avcodec_configuration())
}

// Return the libavcodec license.
func AVCodec_license() string {
	return C.GoString(C.avcodec_license())
}
