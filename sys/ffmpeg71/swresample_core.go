package ffmpeg

import (
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libswresample
#include <libswresample/swresample.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Allocate SwrContext.
func SWResample_alloc() *SWRContext {
	return (*SWRContext)(C.swr_alloc())
}

// Free the given SwrContext.
func SWResample_free(ctx *SWRContext) {
	C.swr_free((**C.struct_SwrContext)(unsafe.Pointer(&ctx)))
}

// Initialize context after user parameters have been set.
func SWResample_init(ctx *SWRContext) error {
	if err := AVError(C.swr_init((*C.struct_SwrContext)(ctx))); err == 0 {
		return nil
	} else {
		return err
	}
}

// Closes the context so that swr_is_initialized() returns 0
func SWResample_close(ctx *SWRContext) {
	C.swr_close((*C.struct_SwrContext)(ctx))
}

// Check whether an swr context has been initialized or not.
func SWResample_is_initialized(ctx *SWRContext) bool {
	return C.swr_is_initialized((*C.struct_SwrContext)(ctx)) != 0
}
