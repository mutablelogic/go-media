package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libswresample
#include <libswresample/swresample.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the LIBSWRESAMPLE_VERSION_INT constant.
func SWResample_version() uint {
	return uint(C.swresample_version())
}

// Return the swr build-time configuration.
func SWResample_configuration() string {
	return C.GoString(C.swresample_configuration())
}

// Return the swr license.
func SWResample_license() string {
	return C.GoString(C.swresample_license())
}
