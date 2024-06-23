package ffmpeg

import (
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

import "C"

////////////////////////////////////////////////////////////////////////////////
// FUNCTIONS

func boolToInt(v bool) C.int {
	if v {
		return C.int(1)
	}
	return 0
}

func cByteSlice(p unsafe.Pointer, sz C.int) []byte {
	return cUint8Slice(p, sz)
}

func cIntSlice(p unsafe.Pointer, sz C.int) []int {
	if p == nil {
		return nil
	}
	return (*[1 << 30]int)(p)[:int(sz)]
}

func cUint8Slice(p unsafe.Pointer, sz C.int) []uint8 {
	if p == nil {
		return nil
	}
	return (*[1 << 30]uint8)(p)[:int(sz)]
}

func cInt8Slice(p unsafe.Pointer, sz C.int) []int8 {
	if p == nil {
		return nil
	}
	return (*[1 << 30]int8)(p)[:int(sz)]
}

func cUint16Slice(p unsafe.Pointer, sz C.int) []uint16 {
	if p == nil {
		return nil
	}
	return (*[1 << 30]uint16)(p)[:int(sz)]
}

func cInt16Slice(p unsafe.Pointer, sz C.int) []int16 {
	if p == nil {
		return nil
	}
	return (*[1 << 30]int16)(p)[:int(sz)]
}

func cUint32Slice(p unsafe.Pointer, sz C.int) []uint32 {
	if p == nil {
		return nil
	}
	return (*[1 << 30]uint32)(p)[:int(sz)]
}

func cInt32Slice(p unsafe.Pointer, sz C.int) []int32 {
	if p == nil {
		return nil
	}
	return (*[1 << 30]int32)(p)[:int(sz)]
}

func cFloat32Slice(p unsafe.Pointer, sz C.int) []float32 {
	if p == nil {
		return nil
	}
	return (*[1 << 30]float32)(p)[:int(sz)]
}

func cFloat64Slice(p unsafe.Pointer, sz C.int) []float64 {
	if p == nil {
		return nil
	}
	return (*[1 << 30]float64)(p)[:int(sz)]
}

func cAVStreamSlice(p unsafe.Pointer, sz C.int) []*AVStream {
	if p == nil {
		return nil
	}
	return (*[1 << 30]*AVStream)(p)[:int(sz)]
}

func cAVDeviceInfoSlice(p unsafe.Pointer, sz C.int) []*AVDeviceInfo {
	if p == nil {
		return nil
	}
	return (*[1 << 30]*AVDeviceInfo)(p)[:int(sz)]
}

func cAVMediaTypeSlice(p unsafe.Pointer, sz C.int) []AVMediaType {
	if p == nil {
		return nil
	}
	return (*[1 << 30]AVMediaType)(p)[:int(sz)]
}
