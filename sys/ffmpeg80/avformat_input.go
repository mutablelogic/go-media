package ffmpeg

import (
	"encoding/json"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVInputFormat C.struct_AVInputFormat
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVInputFormat) MarshalJSON() ([]byte, error) {
	type jsonAVInputFormat struct {
		Name       string   `json:"name,omitempty"`
		LongName   string   `json:"long_name,omitempty"`
		MimeTypes  string   `json:"mime_type,omitempty"`
		Extensions string   `json:"extensions,omitempty"`
		Flags      AVFormat `json:"flags,omitempty"`
	}
	return json.Marshal(jsonAVInputFormat{
		Name:       C.GoString(ctx.name),
		LongName:   C.GoString(ctx.long_name),
		MimeTypes:  C.GoString(ctx.mime_type),
		Extensions: C.GoString(ctx.extensions),
		Flags:      AVFormat(ctx.flags),
	})
}

func (ctx *AVInputFormat) String() string {
	str, _ := json.MarshalIndent(ctx, "", "  ")
	return string(str)
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

// Find AVInputFormat based on the short name of the input format.
func AVFormat_find_input_format(name string) *AVInputFormat {
	cString := C.CString(name)
	defer C.free(unsafe.Pointer(cString))
	return (*AVInputFormat)(C.av_find_input_format(cString))
}

// Iterate over all AVInputFormats
func AVFormat_demuxer_iterate(opaque *uintptr) *AVInputFormat {
	return (*AVInputFormat)(C.av_demuxer_iterate((*unsafe.Pointer)(unsafe.Pointer(opaque))))
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (ctx *AVInputFormat) Name() string {
	return C.GoString(ctx.name)
}

func (ctx *AVInputFormat) LongName() string {
	return C.GoString(ctx.long_name)
}

func (ctx *AVInputFormat) Flags() AVFormat {
	return AVFormat(ctx.flags)
}

func (ctx *AVInputFormat) MimeTypes() string {
	return C.GoString(ctx.mime_type)
}

func (ctx *AVInputFormat) Extensions() string {
	return C.GoString(ctx.extensions)
}
