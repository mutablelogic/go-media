package ffmpeg

import (
	"fmt"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/dict.h>
#include <stdlib.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (d *AVDictionary) String() string {
	if d.ctx == nil {
		return "<nil>"
	} else {
		return fmt.Sprintf("<AVDictionary>{ count=%v entries=%v }", d.AVUtil_av_dict_count(), d.AVUtil_av_dict_entries())
	}
}

func (e *AVDictionaryEntry) String() string {
	return fmt.Sprintf("%v=%q", e.Key(), e.Value())
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Create a new dictionary.
func AVUtil_av_dict_new() *AVDictionary {
	return new(AVDictionary)
}

// Free a dictionary and all entries in the dictionary.
func (d *AVDictionary) AVUtil_av_dict_free() {
	C.av_dict_free(((**C.struct_AVDictionary)(unsafe.Pointer(&d.ctx))))
}

// Return the context object for the dictionary.
func (d *AVDictionary) AVUtil_av_dict_context() *C.struct_AVDictionary {
	return d.ctx
}

// Get the number of entries in the dictionary.
func (d *AVDictionary) AVUtil_av_dict_count() int {
	if d.ctx == nil {
		return 0
	} else {
		return int(C.av_dict_count(d.ctx))
	}
}

// Set the given entry, overwriting an existing entry.
func (d *AVDictionary) AVUtil_av_dict_set(key, value string, flags AVDictionaryFlag) error {
	cKey, cValue := C.CString(key), C.CString(value)
	defer C.free(unsafe.Pointer(cKey))
	defer C.free(unsafe.Pointer(cValue))
	if err := AVError(C.av_dict_set(&d.ctx, cKey, cValue, C.int(flags))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Get a dictionary entry with matching key.
func (d *AVDictionary) AVUtil_av_dict_get(key string, prev *AVDictionaryEntry, flags AVDictionaryFlag) *AVDictionaryEntry {
	if d.ctx == nil {
		return nil
	}
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	return (*AVDictionaryEntry)(C.av_dict_get(d.ctx, cKey, (*C.struct_AVDictionaryEntry)(prev), C.int(flags)))
}

// Get the keys for the dictionary.
func (d *AVDictionary) AVUtil_av_dict_keys() []string {
	if d.ctx == nil {
		return nil
	}
	keys := make([]string, 0, d.AVUtil_av_dict_count())
	entry := d.AVUtil_av_dict_get("", nil, AV_DICT_IGNORE_SUFFIX)
	for entry != nil {
		keys = append(keys, entry.Key())
		entry = d.AVUtil_av_dict_get("", entry, AV_DICT_IGNORE_SUFFIX)
	}
	return keys
}

// Get the entries for the dictionary.
func (d *AVDictionary) AVUtil_av_dict_entries() []*AVDictionaryEntry {
	if d.ctx == nil {
		return nil
	}
	result := make([]*AVDictionaryEntry, 0, d.AVUtil_av_dict_count())
	entry := d.AVUtil_av_dict_get("", nil, AV_DICT_IGNORE_SUFFIX)
	for entry != nil {
		result = append(result, entry)
		entry = d.AVUtil_av_dict_get("", entry, AV_DICT_IGNORE_SUFFIX)
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////
// DICTIONARY ENTRY

func (e *AVDictionaryEntry) Key() string {
	return C.GoString(e.key)
}

func (e *AVDictionaryEntry) Value() string {
	return C.GoString(e.value)
}
