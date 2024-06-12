package ffmpeg

import (
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#include <libavutil/imgutils.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func AVUtil_image_linesizes(pixfmt AVPixelFormat, width int) [4]C.int {
	var strides [4]C.int
	C.av_image_fill_linesizes(&strides[0], C.enum_AVPixelFormat(pixfmt), C.int(width))
	return strides
}

// Fill plane sizes for an image with pixel format pix_fmt and height height.
func AVUtil_image_plane_sizes(pixfmt AVPixelFormat, height int, strides [4]C.int) ([4]C.size_t, error) {
	var planes [4]C.size_t
	var strides_ [4]C.ptrdiff_t
	for i := 0; i < 4; i++ {
		strides_[i] = C.ptrdiff_t(strides[i])
	}
	if ret := C.av_image_fill_plane_sizes(&planes[0], C.enum_AVPixelFormat(pixfmt), C.int(height), &strides_[0]); ret < 0 {
		return [4]C.size_t{}, AVError(ret)
	} else {
		return planes, nil
	}
}

// Fill plane sizes for an image with pixel format pix_fmt and height height.
func AVUtil_image_plane_sizes_ex(width, height int, pixfmt AVPixelFormat) ([4]C.size_t, error) {
	return AVUtil_image_plane_sizes(pixfmt, height, AVUtil_image_linesizes(pixfmt, width))
}

// Allocate an image buffer with size, pixel format and alignment suitable for the image
// The allocated image buffer has to be freed by using AVUtil_image_free
// The return values are the allocated data pointers, the strides and the size of the allocated data
func AVUtil_image_alloc(width, height int, pixfmt AVPixelFormat, align int) ([][]byte, []int, int, error) {
	var data [4]*C.uint8_t
	var stride [4]C.int
	if ret := C.av_image_alloc(&data[0], &stride[0], C.int(width), C.int(height), C.enum_AVPixelFormat(pixfmt), C.int(align)); ret < 0 {
		return nil, nil, 0, AVError(ret)
	} else if planeSizes, err := AVUtil_image_plane_sizes_ex(width, height, pixfmt); err != nil {
		return nil, nil, 0, err
	} else {
		dataSlice := make([][]byte, 4)
		strideSlice := make([]int, 4)
		for i := 0; i < 4; i++ {
			if data[i] != nil {
				dataSlice[i] = cByteSlice(unsafe.Pointer(data[i]), C.int(planeSizes[i]))
			}
			strideSlice[i] = int(stride[i])
		}
		return dataSlice, strideSlice, int(ret), nil
	}
}

// Free an image buffer allocated by AVUtil_image_alloc
func AVUtil_image_free(data [][]byte) {
	ptrs := avutil_image_ptr(data)
	C.av_free(unsafe.Pointer(ptrs[0]))
}

// Return the image as a byte buffer
func AVUtil_image_bytes(data [][]byte, size int) []byte {
	ptrs := avutil_image_ptr(data)
	return cByteSlice(unsafe.Pointer(ptrs[0]), C.int(size))
}

// Convert [][]byte to a [4]*C.uint8_t
func avutil_image_ptr(data [][]byte) [4]*C.uint8_t {
	var ptrs [4]*C.uint8_t
	if len(data) != 4 {
		return ptrs
	}
	for i := 0; i < 4; i++ {
		if len(data[i]) == 0 {
			ptrs[i] = nil
		} else {
			ptrs[i] = (*C.uint8_t)(unsafe.Pointer(&data[i][0]))
		}
	}
	return ptrs
}
