package ffmpeg

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libswscale
#include <libswscale/swscale.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	SWSContext C.struct_SwsContext
	SWSFilter  C.struct_SwsFilter
	SWSFlag    C.int
)

const (
	SWS_NONE          SWSFlag = 0
	SWS_FAST_BILINEAR SWSFlag = C.SWS_FAST_BILINEAR
	SWS_BILINEAR      SWSFlag = C.SWS_BILINEAR
	SWS_BICUBIC       SWSFlag = C.SWS_BICUBIC
	SWS_X             SWSFlag = C.SWS_X
	SWS_POINT         SWSFlag = C.SWS_POINT
	SWS_AREA          SWSFlag = C.SWS_AREA
	SWS_BICUBLIN      SWSFlag = C.SWS_BICUBLIN
	SWS_GAUSS         SWSFlag = C.SWS_GAUSS
	SWS_SINC          SWSFlag = C.SWS_SINC
	SWS_LANCZOS       SWSFlag = C.SWS_LANCZOS
	SWS_SPLINE        SWSFlag = C.SWS_SPLINE
	SWS_MIN                   = SWS_FAST_BILINEAR
	SWS_MAX                   = SWS_SPLINE
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v SWSFlag) FlagString() string {
	switch v {
	case SWS_NONE:
		return "SWS_NONE"
	case SWS_FAST_BILINEAR:
		return "SWS_FAST_BILINEAR"
	case SWS_BILINEAR:
		return "SWS_BILINEAR"
	case SWS_BICUBIC:
		return "SWS_BICUBIC"
	case SWS_X:
		return "SWS_X"
	case SWS_POINT:
		return "SWS_POINT"
	case SWS_AREA:
		return "SWS_AREA"
	case SWS_BICUBLIN:
		return "SWS_BICUBLIN"
	case SWS_GAUSS:
		return "SWS_GAUSS"
	case SWS_SINC:
		return "SWS_SINC"
	case SWS_LANCZOS:
		return "SWS_LANCZOS"
	case SWS_SPLINE:
		return "SWS_SPLINE"
	default:
		return "[?? Invalid SWSFlag value]"
	}
}

func (v SWSFlag) String() string {
	if v == SWS_NONE {
		return v.FlagString()
	}
	str := ""
	for i := SWS_MIN; i <= SWS_MAX; i <<= 1 {
		if v&i != 0 {
			str += "|" + i.FlagString()
		}
	}
	return str[1:]
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - VERSION

// Return the LIBSWSCALE_VERSION_INT constant.
func SWS_version() uint {
	return uint(C.swscale_version())
}

// Return the swr build-time configuration.
func SWS_configuration() string {
	return C.GoString(C.swscale_configuration())
}

// Return the swr license.
func SWS_license() string {
	return C.GoString(C.swscale_license())
}

// Get the AVClass for swsContext.
func SWS_get_class() *AVClass {
	return (*AVClass)(C.sws_get_class())
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a pointer to yuv<->rgb coefficients for the given colorspace suitable for sws_setColorspaceDetails().
func SWS_get_coefficients(colorspace int) *int {
	return (*int)(unsafe.Pointer(C.sws_getCoefficients(C.int(colorspace))))
}

// Return true if pix_fmt is a supported input format
func SWS_is_supported_input(pix_fmt AVPixelFormat) bool {
	return intToBool(int(C.sws_isSupportedInput(C.enum_AVPixelFormat(pix_fmt))))
}

// Return true if pix_fmt is a supported output format
func SWS_is_supported_output(pix_fmt AVPixelFormat) bool {
	return intToBool(int(C.sws_isSupportedOutput(C.enum_AVPixelFormat(pix_fmt))))
}

// Return true if an endianness conversion for pix_fmt is supported
func sws_is_supported_endianness_conversion(pix_fmt AVPixelFormat) bool {
	return intToBool(int(C.sws_isSupportedEndiannessConversion(C.enum_AVPixelFormat(pix_fmt))))
}

// Allocate an empty SWSContext.
func SWS_alloc_context() *SWSContext {
	return (*SWSContext)(C.sws_alloc_context())
}

// Initialize the swscaler context sws_context.
func SWS_init_context(ctx *SWSContext, src, dst *SWSFilter) bool {
	return intToBool(int(C.sws_init_context((*C.struct_SwsContext)(ctx), (*C.struct_SwsFilter)(src), (*C.struct_SwsFilter)(dst))))
}

// Free the swscaler context swsContext.
func SWS_free_context(ctx *SWSContext) {
	C.sws_freeContext((*C.struct_SwsContext)(ctx))
}

// Allocate and return an SwsContext.
func SWS_get_context(src_width, src_height int, src_format AVPixelFormat, dst_width, dst_height int, dst_format AVPixelFormat, flags SWSFlag, src_filter, dst_filter *SWSFilter, param *float64) *SWSContext {
	return (*SWSContext)(C.sws_getContext(C.int(src_width), C.int(src_height), C.enum_AVPixelFormat(src_format), C.int(dst_width), C.int(dst_height), C.enum_AVPixelFormat(dst_format), C.int(flags), (*C.struct_SwsFilter)(src_filter), (*C.struct_SwsFilter)(dst_filter), (*C.double)(unsafe.Pointer(param))))
}

// Scale the image slice in srcSlice and put the resulting scaled slice in the image in dst.
func SWS_scale(ctx *SWSContext, src_slice **uint8, src_stride *int, src_slice_y, src_slice_height int, dst **uint8, dst_stride *int) int {
	return int(C.sws_scale((*C.struct_SwsContext)(ctx), (**C.uint8_t)(unsafe.Pointer(src_slice)), (*C.int)(unsafe.Pointer(src_stride)), C.int(src_slice_y), C.int(src_slice_height), (**C.uint8_t)(unsafe.Pointer(dst)), (*C.int)(unsafe.Pointer(dst_stride))))
}

// Scale source data from src and write the output to dst.
func SWS_scale_frame(ctx *SWSContext, src, dst *AVFrame) error {
	if err := AVError(C.sws_scale_frame((*C.struct_SwsContext)(ctx), (*C.struct_AVFrame)(dst), (*C.struct_AVFrame)(src))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Initialize the scaling process for a given pair of source/destination frames.
func SWS_frame_start(ctx *SWSContext, src, dst *AVFrame) error {
	if err := AVError(C.sws_frame_start((*C.struct_SwsContext)(ctx), (*C.struct_AVFrame)(dst), (*C.struct_AVFrame)(src))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Finish the scaling process for a pair of source/destination frames previously submitted with sws_frame_start
func SWS_frame_end(ctx *SWSContext) {
	C.sws_frame_end((*C.struct_SwsContext)(ctx))
}

// Indicate that a horizontal slice of input data is available in the source frame previously provided to sws_frame_start
func SWS_send_slice(ctx *SWSContext, slice_start, slice_height uint) error {
	if err := AVError(C.sws_send_slice((*C.struct_SwsContext)(ctx), C.uint(slice_start), C.uint(slice_height))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Request a horizontal slice of the output data to be written into the frame previously provided to sws_frame_start
func SWS_receive_slice(ctx *SWSContext, slice_start, slice_height uint) error {
	if err := AVError(C.sws_receive_slice((*C.struct_SwsContext)(ctx), C.uint(slice_start), C.uint(slice_height))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Get the alignment required for slices.
func SWS_receive_slice_alignment(ctx *SWSContext) uint {
	return uint(C.sws_receive_slice_alignment((*C.struct_SwsContext)(ctx)))
}
