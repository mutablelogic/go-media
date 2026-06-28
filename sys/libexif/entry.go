package libexif

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libexif
#include <stdlib.h>
#include <libexif/exif-entry.h>
#include <libexif/exif-content.h>

static ExifIfd entry_get_ifd(ExifEntry *e) {
	return e ? exif_content_get_ifd(e->parent) : EXIF_IFD_COUNT;
}
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Entry C.ExifEntry
)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - LIFECYCLE

func Exif_entry_new() *Entry {
	return (*Entry)(C.exif_entry_new())
}

func Exif_entry_ref(entry *Entry) {
	C.exif_entry_ref((*C.ExifEntry)(entry))
}

func Exif_entry_unref(entry *Entry) {
	C.exif_entry_unref((*C.ExifEntry)(entry))
}

func Exif_entry_free(entry *Entry) {
	C.exif_entry_free((*C.ExifEntry)(entry))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - OPERATIONS

func Exif_entry_initialize(entry *Entry, tag Tag) {
	C.exif_entry_initialize((*C.ExifEntry)(entry), C.ExifTag(tag))
}

func Exif_entry_fix(entry *Entry) {
	C.exif_entry_fix((*C.ExifEntry)(entry))
}

func Exif_entry_get_value(entry *Entry) string {
	buf := make([]byte, 1024)
	C.exif_entry_get_value((*C.ExifEntry)(entry), (*C.char)(unsafe.Pointer(&buf[0])), C.uint(len(buf)))
	return C.GoString((*C.char)(unsafe.Pointer(&buf[0])))
}

func Exif_entry_dump(entry *Entry, indent uint) {
	C.exif_entry_dump((*C.ExifEntry)(entry), C.uint(indent))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - FIELD ACCESSORS

func Exif_entry_get_tag(entry *Entry) Tag {
	return Tag((*C.ExifEntry)(entry).tag)
}

func Exif_entry_get_format(entry *Entry) Format {
	return Format((*C.ExifEntry)(entry).format)
}

func Exif_entry_get_components(entry *Entry) uint {
	return uint((*C.ExifEntry)(entry).components)
}

func Exif_entry_get_size(entry *Entry) uint {
	return uint((*C.ExifEntry)(entry).size)
}

func Exif_entry_get_data(entry *Entry) []byte {
	e := (*C.ExifEntry)(entry)
	if e.data == nil || e.size == 0 {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(e.data), C.int(e.size))
}

func Exif_entry_get_ifd(entry *Entry) IFD {
	return IFD(C.entry_get_ifd((*C.ExifEntry)(entry)))
}
