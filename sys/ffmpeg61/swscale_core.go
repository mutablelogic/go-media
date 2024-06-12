package ffmpeg

import (
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libswscale
#include <libswscale/swscale.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Allocate an empty SWSContext.
func SWScale_alloc_context() *SWSContext {
	return (*SWSContext)(C.sws_alloc_context())
}

// Initialize the swscaler context sws_context.
func SWScale_init_context(ctx *SWSContext, src, dst *SWSFilter) {
	C.sws_init_context((*C.struct_SwsContext)(ctx), (*C.struct_SwsFilter)(src), (*C.struct_SwsFilter)(dst))
}

// Free the swscaler context swsContext.
func SWScale_free_context(ctx *SWSContext) {
	C.sws_freeContext((*C.struct_SwsContext)(ctx))
}

// Allocate and return an SwsContext.
func SWScale_get_context(src_width, src_height int, src_format AVPixelFormat, dst_width, dst_height int, dst_format AVPixelFormat, flags SWSFlag, src_filter, dst_filter *SWSFilter, param []float64) *SWSContext {
	var params *C.double
	if len(param) > 0 {
		params = (*C.double)(unsafe.Pointer(&param[0]))
	}
	ctx := C.sws_getContext(C.int(src_width), C.int(src_height), C.enum_AVPixelFormat(src_format), C.int(dst_width), C.int(dst_height), C.enum_AVPixelFormat(dst_format), C.int(flags), (*C.struct_SwsFilter)(src_filter), (*C.struct_SwsFilter)(dst_filter), params)
	return (*SWSContext)(ctx)
}

// Scale the image slice in src and put the resulting scaled slice in the image in dst.
// Returns the height of the output slice.
func SWScale_scale(ctx *SWSContext, src [][]byte, src_stride []int, src_slice_y, src_slice_height int, dest [][]byte, dest_stride []int) int {
	src_ := avutil_image_ptr(src)
	dest_ := avutil_image_ptr(dest)
	return int(C.sws_scale(
		(*C.struct_SwsContext)(ctx),
		(**C.uint8_t)(&src_[0]),
		(*C.int)(unsafe.Pointer(&src_stride[0])),
		C.int(src_slice_y),
		C.int(src_slice_height),
		(**C.uint8_t)(&dest_[0]),
		(*C.int)(unsafe.Pointer(&dest_stride[0]))))
}
