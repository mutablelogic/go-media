package ffmpeg

import (
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Allocate an AVFormatContext for an output format.
func AVFormat_alloc_output_context2(ctx **AVFormatContext, oformat *AVOutputFormat, format string, filename string) error {
	var cFilename, cFormat *C.char
	if format != "" {
		cFormat = C.CString(format)
	}
	if filename != "" {
		cFilename = C.CString(filename)
	}
	defer C.free(unsafe.Pointer(cFilename))
	defer C.free(unsafe.Pointer(cFormat))
	if err := AVError(C.avformat_alloc_output_context2((**C.struct_AVFormatContext)(unsafe.Pointer(ctx)), (*C.struct_AVOutputFormat)(oformat), cFormat, cFilename)); err < 0 {
		return err
	} else {
		return nil
	}
}
