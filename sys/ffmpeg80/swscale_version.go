package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libswscale
#include <libswscale/swscale.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the LIBSWSCALE_VERSION_INT constant.
func SWScale_version() uint {
	return uint(C.swscale_version())
}

// Return the swr build-time configuration.
func SWScale_configuration() string {
	return C.GoString(C.swscale_configuration())
}

// Return the swr license.
func SWScale_license() string {
	return C.GoString(C.swscale_license())
}
