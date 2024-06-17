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
// STRINGIFY

type jsonAVInputFormat struct {
	Name       string   `json:"name,omitempty"`
	LongName   string   `json:"long_name,omitempty"`
	Flags      AVFormat `json:"flags,omitempty"`
	Extensions string   `json:"extensions,omitempty"`
	MimeTypes  string   `json:"mime_types,omitempty"`
}

func (ctx *AVInputFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVInputFormat{
		Name:       C.GoString(ctx.name),
		LongName:   C.GoString(ctx.long_name),
		MimeTypes:  C.GoString(ctx.mime_type),
		Extensions: C.GoString(ctx.extensions),
		Flags:      AVFormat(ctx.flags),
	})
}

func (ctx *AVInputFormat) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

// Iterate over all AVInputFormats
func AVFormat_demuxer_iterate(opaque *uintptr) *AVInputFormat {
	return (*AVInputFormat)(C.av_demuxer_iterate((*unsafe.Pointer)(unsafe.Pointer(opaque))))
}
