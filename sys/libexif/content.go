package libexif

import (
	"runtime/cgo"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libexif
#include <stdint.h>
#include <stdlib.h>
#include <libexif/exif-content.h>

extern void goExifContentForeachEntry(ExifEntry *entry, uintptr_t user_data);

static void call_foreach_entry_cb(ExifEntry *entry, void *user_data) {
	goExifContentForeachEntry(entry, (uintptr_t)user_data);
}

static void call_foreach_entry(ExifContent *content, uintptr_t user_data) {
	exif_content_foreach_entry(content, call_foreach_entry_cb, (void *)user_data);
}
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Content C.ExifContent
)

type ContentForeachEntryFunc func(*Entry)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - LIFECYCLE

func Exif_content_new() *Content {
	return (*Content)(C.exif_content_new())
}

func Exif_content_ref(content *Content) {
	C.exif_content_ref((*C.ExifContent)(content))
}

func Exif_content_unref(content *Content) {
	C.exif_content_unref((*C.ExifContent)(content))
}

func Exif_content_free(content *Content) {
	C.exif_content_free((*C.ExifContent)(content))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - OPERATIONS

func Exif_content_add_entry(content *Content, entry *Entry) {
	C.exif_content_add_entry((*C.ExifContent)(content), (*C.ExifEntry)(entry))
}

func Exif_content_remove_entry(content *Content, entry *Entry) {
	C.exif_content_remove_entry((*C.ExifContent)(content), (*C.ExifEntry)(entry))
}

func Exif_content_get_entry(content *Content, tag Tag) *Entry {
	return (*Entry)(C.exif_content_get_entry((*C.ExifContent)(content), C.ExifTag(tag)))
}

func Exif_content_fix(content *Content) {
	C.exif_content_fix((*C.ExifContent)(content))
}

func Exif_content_foreach_entry(content *Content, fn ContentForeachEntryFunc) {
	h := cgo.NewHandle(fn)
	defer h.Delete()
	C.call_foreach_entry((*C.ExifContent)(content), C.uintptr_t(h))
}

func Exif_content_get_ifd(content *Content) IFD {
	return IFD(C.exif_content_get_ifd((*C.ExifContent)(content)))
}

func Exif_content_dump(content *Content, indent uint) {
	C.exif_content_dump((*C.ExifContent)(content), C.uint(indent))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - DATA ACCESSOR

func Exif_data_get_content(data *Data, ifd IFD) *Content {
	d := (*C.ExifData)(data)
	if int(ifd) < 0 || int(ifd) >= int(C.EXIF_IFD_COUNT) {
		return nil
	}
	return (*Content)(d.ifd[ifd])
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - FIELD ACCESSORS

func Exif_content_get_entry_count(content *Content) uint {
	return uint((*C.ExifContent)(content).count)
}

func Exif_content_get_entry_at(content *Content, n uint) *Entry {
	c := (*C.ExifContent)(content)
	if n >= uint(c.count) {
		return nil
	}
	entries := (*[1 << 20]*C.ExifEntry)(unsafe.Pointer(c.entries))
	return (*Entry)(entries[n])
}

////////////////////////////////////////////////////////////////////////////////
// CALLBACKS

//export goExifContentForeachEntry
func goExifContentForeachEntry(entry *C.ExifEntry, userData C.uintptr_t) {
	h := cgo.Handle(userData)
	h.Value().(ContentForeachEntryFunc)((*Entry)(entry))
}
