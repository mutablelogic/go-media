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
// BINDINGS - PROCESSED IMAGE

func Libraw_dcraw_make_mem_image(data *Data) (*ProcessedImage, int) {
	var errc C.int
	img := C.libraw_dcraw_make_mem_image((*C.libraw_data_t)(data), &errc)
	return (*ProcessedImage)(img), int(errc)
}

func Libraw_dcraw_make_mem_thumb(data *Data) (*ProcessedImage, int) {
	var errc C.int
	img := C.libraw_dcraw_make_mem_thumb((*C.libraw_data_t)(data), &errc)
	return (*ProcessedImage)(img), int(errc)
}

func Libraw_dcraw_clear_mem(img *ProcessedImage) {
	C.libraw_dcraw_clear_mem((*C.libraw_processed_image_t)(img))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - WRITE

func Libraw_dcraw_ppm_tiff_writer(data *Data, filename string) int {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))
	return int(C.libraw_dcraw_ppm_tiff_writer((*C.libraw_data_t)(data), cfilename))
}

func Libraw_dcraw_thumb_writer(data *Data, filename string) int {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))
	return int(C.libraw_dcraw_thumb_writer((*C.libraw_data_t)(data), cfilename))
}
