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
	return (*[1 << 30]byte)(p)[:int(sz)]
}

func cUint16Slice(p unsafe.Pointer, sz C.int) []uint16 {
	return (*[1 << 30]uint16)(p)[:int(sz)]
}

func cInt16Slice(p unsafe.Pointer, sz C.int) []int16 {
	return (*[1 << 30]int16)(p)[:int(sz)]
}
