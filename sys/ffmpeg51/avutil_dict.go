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
	str := "<AVDictionary"
	if d != nil {
		str += fmt.Sprintf(" count=%v", AVUtil_av_dict_count(d))
	}
	for _, entry := range AVUtil_av_dict_entries(d) {
		str += fmt.Sprint(" ", entry)
	}
	return str + ">"
}

func (e *AVDictionaryEntry) String() string {
	return fmt.Sprintf("%v=%q", e.Key(), e.Value())
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Free a dictionary and all entries in the dictionary.
func AVUtil_av_dict_free(dict **AVDictionary) {
	if *dict != nil {
		C.av_dict_free(((**C.struct_AVDictionary)(unsafe.Pointer(dict))))
	}
}

// Free a dictionary and all entries in the dictionary.
func AVUtil_av_dict_free_ptr(dict *AVDictionary) {
	if dict != nil {
		C.av_dict_free(((**C.struct_AVDictionary)(unsafe.Pointer(&dict))))
	}
}

// Get the number of entries in the dictionary.
func AVUtil_av_dict_count(dict *AVDictionary) int {
	if dict == nil {
		return 0
	}
	return int(C.av_dict_count((*C.struct_AVDictionary)(dict)))
}

// Set the given entry, overwriting an existing entry.
func AVUtil_av_dict_set(dict *AVDictionary, key, value string, flags AVDictionaryFlag) error {
	if dict == nil {
		return fmt.Errorf("AVUtil_av_dict_set: dict is nil")
	}
	cKey, cValue := C.CString(key), C.CString(value)
	defer C.free(unsafe.Pointer(cKey))
	defer C.free(unsafe.Pointer(cValue))
	if err := AVError(C.av_dict_set((**C.struct_AVDictionary)(unsafe.Pointer(&dict)), cKey, cValue, C.int(flags))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Set the given entry, overwriting an existing entry. If the dict pointer is nil,
// then allocate a new dictionary. Return the dictionary (newly created or not)
func AVUtil_av_dict_set_ptr(dict *AVDictionary, key, value string, flags AVDictionaryFlag) (*AVDictionary, error) {
	cKey, cValue := C.CString(key), C.CString(value)
	defer C.free(unsafe.Pointer(cKey))
	defer C.free(unsafe.Pointer(cValue))
	if err := AVError(C.av_dict_set((**C.struct_AVDictionary)(unsafe.Pointer(&dict)), cKey, cValue, C.int(flags))); err != 0 {
		return dict, err
	} else {
		return dict, nil
	}
}

// Delete the given entry. If dictionary becomes empty, the return value is nil
func AVUtil_av_dict_delete_ptr(dict *AVDictionary, key string) (*AVDictionary, error) {
	if dict == nil {
		return dict, nil
	}
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	if err := AVError(C.av_dict_set((**C.struct_AVDictionary)(unsafe.Pointer(&dict)), cKey, nil, 0)); err != 0 {
		return dict, err
	} else {
		return dict, nil
	}
}

// Get a dictionary entry with matching key.
func AVUtil_av_dict_get(dict *AVDictionary, key string, prev *AVDictionaryEntry, flags AVDictionaryFlag) *AVDictionaryEntry {
	if dict == nil {
		return nil
	}
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	return (*AVDictionaryEntry)(C.av_dict_get((*C.struct_AVDictionary)(dict), cKey, (*C.struct_AVDictionaryEntry)(prev), C.int(flags)))
}

// Get the keys for the dictionary.
func AVUtil_av_dict_keys(dict *AVDictionary) []string {
	if dict == nil {
		return nil
	}
	keys := make([]string, 0, AVUtil_av_dict_count(dict))
	entry := AVUtil_av_dict_get(dict, "", nil, AV_DICT_IGNORE_SUFFIX)
	for entry != nil {
		keys = append(keys, entry.Key())
		entry = AVUtil_av_dict_get(dict, "", entry, AV_DICT_IGNORE_SUFFIX)
	}
	return keys
}

// Get the entries for the dictionary.
func AVUtil_av_dict_entries(dict *AVDictionary) []*AVDictionaryEntry {
	if dict == nil {
		return nil
	}
	result := make([]*AVDictionaryEntry, 0, AVUtil_av_dict_count(dict))
	entry := AVUtil_av_dict_get(dict, "", nil, AV_DICT_IGNORE_SUFFIX)
	for entry != nil {
		result = append(result, entry)
		entry = AVUtil_av_dict_get(dict, "", entry, AV_DICT_IGNORE_SUFFIX)
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
