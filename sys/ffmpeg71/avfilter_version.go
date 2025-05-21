package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavfilter
#include <libavfilter/avfilter.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the LIBAVFILTER_VERSION_INT constant.
func AVFilter_version() uint {
	return uint(C.avfilter_version())
}

// Return the libavfilter build-time configuration.
func AVFilter_configuration() string {
	return C.GoString(C.avfilter_configuration())
}

// Return the libavfilter license.
func AVFilter_license() string {
	return C.GoString(C.avfilter_license())
}
