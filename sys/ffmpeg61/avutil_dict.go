package ffmpeg

import (
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#include <libavutil/dict.h>
#include <stdlib.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Allocate a dictionary
func AVUtil_dict_alloc() *AVDictionary {
	return new(AVDictionary)
}

// Free a dictionary and all entries in the dictionary.
func AVUtil_dict_free(dict *AVDictionary) {
	C.av_dict_free(&dict.ctx)
}

// Get the number of entries in the dictionary.
func AVUtil_dict_count(dict *AVDictionary) int {
	return int(C.av_dict_count(dict.ctx))
}

// Set the given entry, overwriting an existing entry.
func AVUtil_dict_set(dict *AVDictionary, key, value string, flags AVDictionaryFlag) error {
	cKey, cValue := C.CString(key), C.CString(value)
	defer C.free(unsafe.Pointer(cKey))
	defer C.free(unsafe.Pointer(cValue))
	if err := AVError(C.av_dict_set(&dict.ctx, cKey, cValue, C.int(flags))); err != 0 {
		return err
	}
	return nil
}

// Delete the given entry. If dictionary becomes empty, the return value is nil
func AVUtil_dict_delete(dict *AVDictionary, key string) (*AVDictionary, error) {
	if dict == nil {
		return dict, nil
	}
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	if err := AVError(C.av_dict_set(&dict.ctx, cKey, nil, 0)); err != 0 {
		return dict, err
	} else {
		return dict, nil
	}
}

// Get a dictionary entry with matching key.
func AVUtil_dict_get(dict *AVDictionary, key string, prev *AVDictionaryEntry, flags AVDictionaryFlag) *AVDictionaryEntry {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	return (*AVDictionaryEntry)(C.av_dict_get(dict.ctx, cKey, (*C.struct_AVDictionaryEntry)(prev), C.int(flags)))
}

// Get the keys for the dictionary.
func AVUtil_dict_keys(dict *AVDictionary) []string {
	keys := make([]string, 0, AVUtil_dict_count(dict))
	entry := AVUtil_dict_get(dict, "", nil, AV_DICT_IGNORE_SUFFIX)
	for entry != nil {
		keys = append(keys, entry.Key())
		entry = AVUtil_dict_get(dict, "", entry, AV_DICT_IGNORE_SUFFIX)
	}
	return keys
}

// Get the entries for the dictionary.
func AVUtil_dict_entries(dict *AVDictionary) []*AVDictionaryEntry {
	if dict == nil {
		return nil
	}
	result := make([]*AVDictionaryEntry, 0, AVUtil_dict_count(dict))
	entry := AVUtil_dict_get(dict, "", nil, AV_DICT_IGNORE_SUFFIX)
	for entry != nil {
		result = append(result, entry)
		entry = AVUtil_dict_get(dict, "", entry, AV_DICT_IGNORE_SUFFIX)
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////
// DICTIONARY ENTRY

// Return dictionary entry key
func (e *AVDictionaryEntry) Key() string {
	return C.GoString(e.key)
}

// Return dictionary entry value
func (e *AVDictionaryEntry) Value() string {
	return C.GoString(e.value)
}
