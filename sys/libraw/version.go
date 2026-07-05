package libraw

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libraw
#include <libraw/libraw.h>
*/
import "C"

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - VERSION

func Libraw_version() string {
	return C.GoString(C.libraw_version())
}

func Libraw_versionNumber() int {
	return int(C.libraw_versionNumber())
}

func Libraw_capabilities() uint {
	return uint(C.libraw_capabilities())
}

func Libraw_cameraCount() int {
	return int(C.libraw_cameraCount())
}

func Libraw_cameraList() []string {
	list := C.libraw_cameraList()
	count := Libraw_cameraCount()
	if list == nil || count == 0 {
		return nil
	}
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = C.GoString(*(**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(list)) + uintptr(i)*unsafe.Sizeof(*list))))
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - ERROR

func Libraw_strerror(errcode int) string {
	return C.GoString(C.libraw_strerror(C.int(errcode)))
}
