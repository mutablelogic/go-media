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
	AVOutputFormat C.struct_AVOutputFormat
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVOutputFormat) MarshalJSON() ([]byte, error) {
	type jsonAVOutputFormat struct {
		Name          string    `json:"name,omitempty"`
		LongName      string    `json:"long_name,omitempty"`
		MimeTypes     string    `json:"mime_types,omitempty"`
		Flags         AVFormat  `json:"flags,omitempty"`
		Extensions    string    `json:"extensions,omitempty"`
		VideoCodec    AVCodecID `json:"video_codec,omitempty"`
		AudioCodec    AVCodecID `json:"audio_codec,omitempty"`
		SubtitleCodec AVCodecID `json:"subtitle_codec,omitempty"`
	}
	return json.Marshal(jsonAVOutputFormat{
		Name:          C.GoString(ctx.name),
		LongName:      C.GoString(ctx.long_name),
		MimeTypes:     C.GoString(ctx.mime_type),
		Flags:         AVFormat(ctx.flags),
		Extensions:    C.GoString(ctx.extensions),
		VideoCodec:    AVCodecID(ctx.video_codec),
		AudioCodec:    AVCodecID(ctx.audio_codec),
		SubtitleCodec: AVCodecID(ctx.subtitle_codec),
	})
}

func (ctx *AVOutputFormat) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (ctx *AVOutputFormat) Name() string {
	return C.GoString(ctx.name)
}

func (ctx *AVOutputFormat) LongName() string {
	return C.GoString(ctx.long_name)
}

func (ctx *AVOutputFormat) Flags() AVFormat {
	return AVFormat(ctx.flags)
}

func (ctx *AVOutputFormat) MimeTypes() string {
	return C.GoString(ctx.mime_type)
}

func (ctx *AVOutputFormat) Extensions() string {
	return C.GoString(ctx.extensions)
}

func (ctx *AVOutputFormat) VideoCodec() AVCodecID {
	return AVCodecID(ctx.video_codec)
}

func (ctx *AVOutputFormat) AudioCodec() AVCodecID {
	return AVCodecID(ctx.audio_codec)
}

func (ctx *AVOutputFormat) SubtitleCodec() AVCodecID {
	return AVCodecID(ctx.subtitle_codec)
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

// Iterate over all AVOutputFormats
func AVFormat_muxer_iterate(opaque *uintptr) *AVOutputFormat {
	return (*AVOutputFormat)(C.av_muxer_iterate((*unsafe.Pointer)(unsafe.Pointer(opaque))))
}

// Return the output format in the list of registered output formats which best matches the provided parameters, or return NULL if there is no match.
func AVFormat_guess_format(format, filename, mimetype string) *AVOutputFormat {
	var cFilename, cFormat, cMimeType *C.char
	if format != "" {
		cFormat = C.CString(format)
	}
	if filename != "" {
		cFilename = C.CString(filename)
	}
	if mimetype != "" {
		cMimeType = C.CString(mimetype)
	}
	defer C.free(unsafe.Pointer(cFormat))
	defer C.free(unsafe.Pointer(cFilename))
	defer C.free(unsafe.Pointer(cMimeType))
	return (*AVOutputFormat)(C.av_guess_format(cFormat, cFilename, cMimeType))
}
