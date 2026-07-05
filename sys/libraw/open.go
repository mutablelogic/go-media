package libraw

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libraw
#include <stdlib.h>
#include <libraw/libraw.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - OPEN

func Libraw_open_file(data *Data, filename string) int {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))
	return int(C.libraw_open_file((*C.libraw_data_t)(data), cfilename))
}

func Libraw_open_buffer(data *Data, buf []byte) int {
	return int(C.libraw_open_buffer((*C.libraw_data_t)(data), unsafe.Pointer(&buf[0]), C.size_t(len(buf))))
}
