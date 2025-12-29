package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavdevice
#include <libavdevice/avdevice.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the LIBAVDEVICE_VERSION_INT constant.
func AVDevice_version() uint {
	return uint(C.avdevice_version())
}

// Return the libavdevice build-time configuration.
func AVDevice_configuration() string {
	return C.GoString(C.avdevice_configuration())
}

// Return the libavdevice license.
func AVDevice_license() string {
	return C.GoString(C.avdevice_license())
}
