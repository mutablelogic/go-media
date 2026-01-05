package ffmpeg

import (
	"encoding/json"
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
// TYPES

type (
	AVDictionary      struct{ ctx *C.struct_AVDictionary } // Wrapper
	AVDictionaryEntry C.struct_AVDictionaryEntry
	AVDictionaryFlag  C.int
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// No flags
	AV_DICT_NONE AVDictionaryFlag = 0

	// Only get an entry with exact-case key match.
	AV_DICT_MATCH_CASE AVDictionaryFlag = C.AV_DICT_MATCH_CASE

	// Return first entry in a dictionary whose first part corresponds to the search key, ignoring the suffix of the found key string.
	AV_DICT_IGNORE_SUFFIX AVDictionaryFlag = C.AV_DICT_IGNORE_SUFFIX

	// Take ownership of  key that has been allocated with av_malloc()
	AV_DICT_DONT_STRDUP_KEY AVDictionaryFlag = C.AV_DICT_DONT_STRDUP_KEY

	// Take ownership of  value that has been allocated with av_malloc()
	AV_DICT_DONT_STRDUP_VAL AVDictionaryFlag = C.AV_DICT_DONT_STRDUP_VAL

	// Don't overwrite existing entries.
	AV_DICT_DONT_OVERWRITE AVDictionaryFlag = C.AV_DICT_DONT_OVERWRITE

	// Append to existing key.
	AV_DICT_APPEND AVDictionaryFlag = C.AV_DICT_APPEND

	// Allow to store several equal keys in the dictionary.
	AV_DICT_MULTIKEY AVDictionaryFlag = C.AV_DICT_MULTIKEY
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Allocate a dictionary
func AVUtil_dict_alloc() *AVDictionary {
	return new(AVDictionary)
}

// Free a dictionary and all entries in the dictionary.
func AVUtil_dict_free(dict *AVDictionary) {
	if dict == nil {
		return
	}
	C.av_dict_free(&dict.ctx)
}

// Copy entries from one dictionary into another.
func AVUtil_dict_copy(dict *AVDictionary, flags AVDictionaryFlag) (*AVDictionary, error) {
	if dict == nil {
		return nil, nil
	}
	dest := new(AVDictionary)
	if err := AVError(C.av_dict_copy(&dest.ctx, dict.ctx, C.int(flags))); err != 0 {
		// Free any partially allocated dictionary on error
		C.av_dict_free(&dest.ctx)
		return nil, err
	}

	// Return success
	return dest, nil
}

// Get the number of entries in the dictionary.
func AVUtil_dict_count(dict *AVDictionary) int {
	if dict == nil {
		return 0
	}
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

// Parse the key/value pairs list and add the parsed entries to a dictionary.
func AVUtil_dict_parse_string(dict *AVDictionary, opts, key_value_sep, pairs_sep string, flags AVDictionaryFlag) error {
	if dict == nil {
		return nil
	}
	cOpts, cTupleSep, cKeyValueSep := C.CString(opts), C.CString(pairs_sep), C.CString(key_value_sep)
	defer C.free(unsafe.Pointer(cOpts))
	defer C.free(unsafe.Pointer(cTupleSep))
	defer C.free(unsafe.Pointer(cKeyValueSep))
	if err := AVError(C.av_dict_parse_string(&dict.ctx, cOpts, cKeyValueSep, cTupleSep, C.int(flags))); err != 0 {
		return err
	}

	// Success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// DICTIONARY ENTRY

// Return dictionary entry key
func (e *AVDictionaryEntry) Key() string {
	return C.GoString((*C.struct_AVDictionaryEntry)(e).key)
}

// Return dictionary entry value
func (e *AVDictionaryEntry) Value() string {
	return C.GoString((*C.struct_AVDictionaryEntry)(e).value)
}

////////////////////////////////////////////////////////////////////////////////
// JSON OUTPUT

func (ctx *AVDictionary) MarshalJSON() ([]byte, error) {
	type jsonAVDictionary struct {
		Count int                  `json:"count"`
		Elems []*AVDictionaryEntry `json:"elems,omitempty"`
	}
	return json.Marshal(jsonAVDictionary{
		Count: AVUtil_dict_count(ctx),
		Elems: AVUtil_dict_entries(ctx),
	})
}

func (ctx *AVDictionaryEntry) MarshalJSON() ([]byte, error) {
	type jsonAVDictionaryEntry struct {
		Key   string `json:"key"`
		Value string `json:"value,omitempty"`
	}
	return json.Marshal(jsonAVDictionaryEntry{
		Key:   ctx.Key(),
		Value: ctx.Value(),
	})
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVDictionary) String() string {
	return marshalToString(ctx)
}
