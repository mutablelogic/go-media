package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - VERSION

// Return the LIBAVFORMAT_VERSION_INT constant.
func AVUtil_version() uint {
	return uint(C.avutil_version())
}

// Return the libavformat build-time configuration.
func AVUtil_configuration() string {
	return C.GoString(C.avutil_configuration())
}

// Return the libavformat license.
func AVUtil_license() string {
	return C.GoString(C.avutil_license())
}
