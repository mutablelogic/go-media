package ffmpeg

import (
	"fmt"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libswscale
#include <libswscale/swscale.h>
#include <stdio.h>
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

// Check if context can be reused, otherwise reallocate a new one.
//
// If context is nil, just calls sws_getContext() to get a new context.
// Otherwise, checks if the parameters match those already saved in context.
// If they match, returns the current context. Otherwise, frees context and
// gets a new context with the new parameters.
//
// Note: src_filter and dst_filter are not checked, they are assumed to remain the same.
func SWScale_get_cached_context(ctx *SWSContext, src_width, src_height int, src_format AVPixelFormat, dst_width, dst_height int, dst_format AVPixelFormat, flags SWSFlag, src_filter, dst_filter *SWSFilter, param []float64) *SWSContext {
	var params *C.double
	if len(param) > 0 {
		params = (*C.double)(unsafe.Pointer(&param[0]))
	}
	new_ctx := C.sws_getCachedContext(
		(*C.struct_SwsContext)(ctx),
		C.int(src_width), C.int(src_height), C.enum_AVPixelFormat(src_format),
		C.int(dst_width), C.int(dst_height), C.enum_AVPixelFormat(dst_format),
		C.int(flags),
		(*C.struct_SwsFilter)(src_filter), (*C.struct_SwsFilter)(dst_filter),
		params,
	)
	return (*SWSContext)(new_ctx)
}

// Scale the image slice in src and put the resulting scaled slice in the image in dst.
// Returns the height of the output slice.
func SWScale_scale(ctx *SWSContext, src [][]byte, src_stride []int, src_slice_y, src_slice_height int, dest [][]byte, dest_stride []int) int {
	src_, src_stride_ := avutil_image_ptr(src, src_stride)
	dest_, dest_stride_ := avutil_image_ptr(dest, dest_stride)
	return int(C.sws_scale(
		(*C.struct_SwsContext)(ctx),
		&src_[0], &src_stride_[0],
		C.int(src_slice_y),
		C.int(src_slice_height),
		&dest_[0], &dest_stride_[0],
	))
}

// Scale source data from src and write the output to dst.
//
// The native parameter controls which FFmpeg API is used:
//   - native=true:  Uses sws_scale_frame() - works in dynamic mode with uninitialized context
//                   (created via SWScale_alloc_context). Returns -22 (EINVAL) when used with
//                   initialized context (created via SWScale_get_context).
//   - native=false: Uses sws_frame_start/send_slice/receive_slice/frame_end sequence - works
//                   with both initialized and uninitialized contexts.
//
// For best compatibility, use native=false or use an uninitialized context (SWScale_alloc_context)
// with native=true.
func SWScale_scale_frame(ctx *SWSContext, dest, src *AVFrame, native bool) error {
	if native {
		if ret := C.sws_scale_frame((*C.struct_SwsContext)(ctx), (*C.struct_AVFrame)(dest), (*C.struct_AVFrame)(src)); ret != 0 {
			return AVError(ret)
		}
	} else {
		if err := SWScale_frame_start(ctx, dest, src); err != nil {
			return fmt.Errorf("SWScale_frame_start: %w", err)
		}
		if err := SWScale_send_slice(ctx, 0, uint(src.Height())); err != nil {
			return fmt.Errorf("SWScale_send_slice: %w", err)
		}
		if err := SWScale_receive_slice(ctx, 0, uint(dest.Height())); err != nil {
			return fmt.Errorf("SWScale_receive_slice: %w", err)
		}
		SWScale_frame_end(ctx)
	}

	// Return success
	return nil
}

// Initialize the scaling process for a given pair of source/destination frames.
func SWScale_frame_start(ctx *SWSContext, dest, src *AVFrame) error {
	if ret := C.sws_frame_start((*C.struct_SwsContext)(ctx), (*C.struct_AVFrame)(dest), (*C.struct_AVFrame)(src)); ret != 0 {
		return AVError(ret)
	} else {
		return nil
	}
}

// Finish the scaling process for a pair of source/destination frames.
func SWScale_frame_end(ctx *SWSContext) {
	C.sws_frame_end((*C.struct_SwsContext)(ctx))
}

// Indicate that a horizontal slice of input data is available in the source frame
func SWScale_send_slice(ctx *SWSContext, slice_start, slice_height uint) error {
	if ret := C.sws_send_slice((*C.struct_SwsContext)(ctx), C.uint(slice_start), C.uint(slice_height)); ret < 0 {
		return AVError(ret)
	} else {
		return nil
	}
}

// Request a horizontal slice of the output data to be written into the frame
func SWScale_receive_slice(ctx *SWSContext, slice_start, slice_height uint) error {
	if ret := C.sws_receive_slice((*C.struct_SwsContext)(ctx), C.uint(slice_start), C.uint(slice_height)); ret < 0 {
		return AVError(ret)
	} else {
		return nil
	}
}
