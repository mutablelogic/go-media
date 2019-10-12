package ffmpeg

import (
	"fmt"
	"strconv"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

//#cgo pkg-config: libavutil
//#include <libavutil/error.h>
//#include <libavutil/dict.h>
//#include <libavutil/mem.h>
//#include <stdlib.h>
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVError           int
	AVDictionaryEntry C.struct_AVDictionaryEntry
	AVDictionaryFlag  int
	AVRational        C.struct_AVRational
)

type AVDictionary struct {
	ctx *C.struct_AVDictionary
}

////////////////////////////////////////////////////////////////////////////////
// COMSTANTS

const (
	BUF_SIZE = 1024
)

const (
	AV_DICT_NONE            AVDictionaryFlag = 0
	AV_DICT_MATCH_CASE      AVDictionaryFlag = 1
	AV_DICT_IGNORE_SUFFIX   AVDictionaryFlag = 2
	AV_DICT_DONT_STRDUP_KEY AVDictionaryFlag = 4
	AV_DICT_DONT_STRDUP_VAL AVDictionaryFlag = 8
	AV_DICT_DONT_OVERWRITE  AVDictionaryFlag = 16
	AV_DICT_APPEND          AVDictionaryFlag = 32
	AV_DICT_MULTIKEY        AVDictionaryFlag = 64
)

////////////////////////////////////////////////////////////////////////////////
// ERROR HANDLINE

func (this AVError) Error() string {
	cbuffer := make([]byte, BUF_SIZE)
	if err := C.av_strerror(C.int(this), (*C.char)(unsafe.Pointer(&cbuffer[0])), BUF_SIZE); err == 0 {
		return string(cbuffer)
	} else {
		return fmt.Sprintf("Error code: %v", this)
	}
}

////////////////////////////////////////////////////////////////////////////////
// DICTIONARY

func NewAVDictionary() *AVDictionary {
	return new(AVDictionary)
}

func (this *AVDictionary) Close() {
	if this.ctx != nil {
		C.av_dict_free(&this.ctx)
	}
}

func (this *AVDictionary) Count() int {
	if this.ctx == nil {
		return 0
	} else {
		return int(C.av_dict_count(this.ctx))
	}
}

func (this *AVDictionary) Get(key string, prev *AVDictionaryEntry, flags AVDictionaryFlag) *AVDictionaryEntry {
	if this.ctx == nil {
		return nil
	} else {
		key_ := C.CString(key)
		defer C.free(unsafe.Pointer(key_))
		return (*AVDictionaryEntry)(C.av_dict_get(this.ctx, key_, (*C.struct_AVDictionaryEntry)(prev), C.int(flags)))
	}
}

func (this *AVDictionary) Set(key, value string, flags AVDictionaryFlag) error {
	key_ := C.CString(key)
	value_ := C.CString(value)
	defer C.free(unsafe.Pointer(key_))
	defer C.free(unsafe.Pointer(value_))
	if err := AVError(C.av_dict_set(&this.ctx, key_, value_, C.int(flags))); err != 0 {
		return err
	} else {
		return nil
	}
}

func (this *AVDictionary) Keys() []string {
	keys := make([]string, 0, this.Count())
	entry := this.Get("", nil, AV_DICT_IGNORE_SUFFIX)
	for entry != nil {
		keys = append(keys, entry.Key())
		entry = this.Get("", entry, AV_DICT_IGNORE_SUFFIX)
	}
	return keys
}

func (this *AVDictionary) Entries() []*AVDictionaryEntry {
	keys := make([]*AVDictionaryEntry, 0, this.Count())
	entry := this.Get("", nil, AV_DICT_IGNORE_SUFFIX)
	for entry != nil {
		keys = append(keys, entry)
		entry = this.Get("", entry, AV_DICT_IGNORE_SUFFIX)
	}
	return keys
}

func (this *AVDictionary) String() string {
	if this.Count() == 0 {
		return fmt.Sprintf("<AVDictionary>{ }")
	} else {
		return fmt.Sprintf("<AVDictionary>{ count=%v entries=%v }", this.Count(), this.Entries())
	}
}

func (this *AVDictionary) context() *C.struct_AVDictionary {
	return this.ctx
}

////////////////////////////////////////////////////////////////////////////////
// DICTIONARY ENTRY

func (this *AVDictionaryEntry) Key() string {
	return C.GoString(this.key)
}

func (this *AVDictionaryEntry) Value() string {
	return C.GoString(this.value)
}

func (this *AVDictionaryEntry) String() string {
	return fmt.Sprintf("%v=%v", this.Key(), strconv.Quote(this.Value()))
}

////////////////////////////////////////////////////////////////////////////////
// RATIONAL NUMBER

func (this AVRational) Num() int {
	return int(this.num)
}

func (this AVRational) Den() int {
	return int(this.den)
}

func (this AVRational) String() string {
	if this.Num() == 0 {
		return "0"
	} else {
		return fmt.Sprintf("<AVRational>{ num=%v den=%v }", this.Num(), this.Den())
	}
}
