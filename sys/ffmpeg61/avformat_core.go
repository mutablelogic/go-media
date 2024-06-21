package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

// Allocate an AVFormatContext.
func AVFormat_alloc_context() *AVFormatContext {
	return (*AVFormatContext)(C.avformat_alloc_context())
}

func AVFormat_free_context(ctx *AVFormatContext) {
	C.avformat_free_context((*C.struct_AVFormatContext)(ctx))
}

// Initialise network
func AVFormat_network_init() error {
	if ret := C.avformat_network_init(); ret != 0 {
		return AVError(ret)
	} else {
		return nil
	}
}
