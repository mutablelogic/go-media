package libexif

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libexif
#include <stdlib.h>
#include <libexif/exif-loader.h>
#include <libexif/exif-data.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Loader C.ExifLoader
)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

func Exif_loader_new() *Loader {
	return (*Loader)(C.exif_loader_new())
}

func Exif_loader_ref(loader *Loader) {
	C.exif_loader_ref((*C.ExifLoader)(loader))
}

func Exif_loader_reset(loader *Loader) {
	C.exif_loader_reset((*C.ExifLoader)(loader))
}

func Exif_loader_unref(loader *Loader) {
	C.exif_loader_unref((*C.ExifLoader)(loader))
}

func Exif_loader_write(loader *Loader, data []byte) {
	C.exif_loader_write((*C.ExifLoader)(loader), (*C.uchar)(unsafe.Pointer(&data[0])), C.uint(len(data)))
}

func Exif_loader_write_file(loader *Loader, filename string) {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))
	C.exif_loader_write_file((*C.ExifLoader)(loader), cfilename)
}

func Exif_loader_get_data(loader *Loader) *Data {
	return (*Data)(C.exif_loader_get_data((*C.ExifLoader)(loader)))
}

func Exif_loader_get_buf(loader *Loader) []byte {
	var ptr *C.uchar
	var size C.uint
	C.exif_loader_get_buf((*C.ExifLoader)(loader), &ptr, &size)
	if ptr == nil || size == 0 {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(ptr), C.int(size))
}
