package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the LIBAVFORMAT_VERSION_INT constant.
func AVFormat_version() uint {
	return uint(C.avformat_version())
}

// Return the libavformat build-time configuration.
func AVFormat_configuration() string {
	return C.GoString(C.avformat_configuration())
}

// Return the libavformat license.
func AVFormat_license() string {
	return C.GoString(C.avformat_license())
}
