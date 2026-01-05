package ffmpeg

import (
	"encoding/json"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#include <stdint.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// UTILITY FUNCTIONS

// Marshal a value to JSON string with indentation.
// Returns error message if marshaling fails.
func marshalToString(v interface{}) string {
	if str, err := json.MarshalIndent(v, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

// Convert a Go bool to a C int (0 or 1)
func boolToInt(v bool) C.int {
	if v {
		return C.int(1)
	}
	return 0
}

// Convert a C pointer and size to a Go byte slice
func cByteSlice(p unsafe.Pointer, sz C.int) []byte {
	return cUint8Slice(p, sz)
}

// Convert a C pointer and size to a Go int slice
// Returns nil if pointer is nil or size is non-positive
func cIntSlice(p unsafe.Pointer, sz C.int) []int {
	if p == nil || sz <= 0 {
		return nil
	}
	return (*[1 << 30]int)(p)[:int(sz):int(sz)]
}

// Convert a C pointer and size to a Go uint8 slice
// Returns nil if pointer is nil or size is non-positive
func cUint8Slice(p unsafe.Pointer, sz C.int) []uint8 {
	if p == nil || sz <= 0 {
		return nil
	}
	return (*[1 << 30]uint8)(p)[:int(sz):int(sz)]
}

// Convert a C pointer and size to a Go int8 slice
// Returns nil if pointer is nil or size is non-positive
func cInt8Slice(p unsafe.Pointer, sz C.int) []int8 {
	if p == nil || sz <= 0 {
		return nil
	}
	return (*[1 << 30]int8)(p)[:int(sz):int(sz)]
}

// Convert a C pointer and size to a Go uint16 slice
// Returns nil if pointer is nil or size is non-positive
func cUint16Slice(p unsafe.Pointer, sz C.int) []uint16 {
	if p == nil || sz <= 0 {
		return nil
	}
	return (*[1 << 30]uint16)(p)[:int(sz):int(sz)]
}

// Convert a C pointer and size to a Go int16 slice
// Returns nil if pointer is nil or size is non-positive
func cInt16Slice(p unsafe.Pointer, sz C.int) []int16 {
	if p == nil || sz <= 0 {
		return nil
	}
	return (*[1 << 30]int16)(p)[:int(sz):int(sz)]
}

// Convert a C pointer and size to a Go uint32 slice
// Returns nil if pointer is nil or size is non-positive
func cUint32Slice(p unsafe.Pointer, sz C.int) []uint32 {
	if p == nil || sz <= 0 {
		return nil
	}
	return (*[1 << 30]uint32)(p)[:int(sz):int(sz)]
}

// Convert a C pointer and size to a Go int32 slice
// Returns nil if pointer is nil or size is non-positive
func cInt32Slice(p unsafe.Pointer, sz C.int) []int32 {
	if p == nil || sz <= 0 {
		return nil
	}
	return (*[1 << 30]int32)(p)[:int(sz):int(sz)]
}

// Convert a C pointer and size to a Go float32 slice
// Returns nil if pointer is nil or size is non-positive
func cFloat32Slice(p unsafe.Pointer, sz C.int) []float32 {
	if p == nil || sz <= 0 {
		return nil
	}
	return (*[1 << 30]float32)(p)[:int(sz):int(sz)]
}

// Convert a C pointer and size to a Go float64 slice
// Returns nil if pointer is nil or size is non-positive
func cFloat64Slice(p unsafe.Pointer, sz C.int) []float64 {
	if p == nil || sz <= 0 {
		return nil
	}
	return (*[1 << 30]float64)(p)[:int(sz):int(sz)]
}

func cAVStreamSlice(p unsafe.Pointer, sz C.int) []*AVStream {
	if p == nil || sz <= 0 {
		return nil
	}
	return (*[1 << 30]*AVStream)(p)[:int(sz)]
}

func cAVDeviceInfoSlice(p unsafe.Pointer, sz C.int) []*AVDeviceInfo {
	if p == nil || sz <= 0 {
		return nil
	}
	return (*[1 << 30]*AVDeviceInfo)(p)[:int(sz)]
}

func cAVMediaTypeSlice(p unsafe.Pointer, sz C.int) []AVMediaType {
	if p == nil || sz <= 0 {
		return nil
	}
	return (*[1 << 30]AVMediaType)(p)[:int(sz)]
}

// Convert Go image slices to C pointers for swscale operations
func avutil_image_ptr(data [][]byte, stride []int) ([4]*C.uint8_t, [4]C.int) {
	var ptrs [4]*C.uint8_t
	var strides [4]C.int
	for i := 0; i < 4; i++ {
		if i < len(data) && len(data[i]) > 0 {
			ptrs[i] = (*C.uint8_t)(unsafe.Pointer(&data[i][0]))
		} else {
			ptrs[i] = nil
		}
		if i < len(stride) {
			strides[i] = C.int(stride[i])
		}
	}
	return ptrs, strides
}
