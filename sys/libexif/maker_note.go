package libexif

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libexif
#include <stdlib.h>
#include <libexif/exif-mnote-data.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	MakerNoteData C.ExifMnoteData
)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - LIFECYCLE

func Exif_mnote_data_ref(d *MakerNoteData) {
	C.exif_mnote_data_ref((*C.ExifMnoteData)(d))
}

func Exif_mnote_data_unref(d *MakerNoteData) {
	C.exif_mnote_data_unref((*C.ExifMnoteData)(d))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - LOAD / SAVE

func Exif_mnote_data_load(d *MakerNoteData, buf []byte) {
	C.exif_mnote_data_load((*C.ExifMnoteData)(d), (*C.uchar)(unsafe.Pointer(&buf[0])), C.uint(len(buf)))
}

func Exif_mnote_data_save(d *MakerNoteData) []byte {
	var ptr *C.uchar
	var size C.uint
	C.exif_mnote_data_save((*C.ExifMnoteData)(d), &ptr, &size)
	if ptr == nil || size == 0 {
		return nil
	}
	defer C.free(unsafe.Pointer(ptr))
	return C.GoBytes(unsafe.Pointer(ptr), C.int(size))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - QUERY

func Exif_mnote_data_count(d *MakerNoteData) uint {
	return uint(C.exif_mnote_data_count((*C.ExifMnoteData)(d)))
}

func Exif_mnote_data_get_id(d *MakerNoteData, n uint) uint {
	return uint(C.exif_mnote_data_get_id((*C.ExifMnoteData)(d), C.uint(n)))
}

func Exif_mnote_data_get_name(d *MakerNoteData, n uint) string {
	return C.GoString(C.exif_mnote_data_get_name((*C.ExifMnoteData)(d), C.uint(n)))
}

func Exif_mnote_data_get_title(d *MakerNoteData, n uint) string {
	return C.GoString(C.exif_mnote_data_get_title((*C.ExifMnoteData)(d), C.uint(n)))
}

func Exif_mnote_data_get_description(d *MakerNoteData, n uint) string {
	return C.GoString(C.exif_mnote_data_get_description((*C.ExifMnoteData)(d), C.uint(n)))
}

func Exif_mnote_data_get_value(d *MakerNoteData, n uint) string {
	buf := make([]byte, 1024)
	C.exif_mnote_data_get_value((*C.ExifMnoteData)(d), C.uint(n), (*C.char)(unsafe.Pointer(&buf[0])), C.uint(len(buf)))
	return C.GoString((*C.char)(unsafe.Pointer(&buf[0])))
}

func Exif_mnote_data_log(d *MakerNoteData, log *Log) {
	C.exif_mnote_data_log((*C.ExifMnoteData)(d), (*C.ExifLog)(log))
}
