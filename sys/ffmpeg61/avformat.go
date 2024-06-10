package ffmpeg

import (
	"encoding/json"
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
	AVInputFormat   C.struct_AVInputFormat
	AVOutputFormat  C.struct_AVOutputFormat
	AVFormatContext C.struct_AVFormatContext
	AVIOContext     C.struct_AVIOContext
)

type jsonAVFormatContext struct {
	Pb     *AVIOContext    `json:"pb,omitempty"`
	Input  *AVInputFormat  `json:"input_format,omitempty"`
	Output *AVOutputFormat `json:"output_format,omitempty"`
	Url    string          `json:"url,omitempty"`
}

type jsonAVIOContext struct {
	IsEOF        bool   `json:"is_eof,omitempty"`
	IsWriteable  bool   `json:"is_writeable,omitempty"`
	IsSeekable   bool   `json:"is_seekable,omitempty"`
	IsDirect     bool   `json:"is_direct,omitempty"`
	Pos          int64  `json:"pos,omitempty"`
	BufferSize   int    `json:"buffer_size,omitempty"`
	BytesRead    int64  `json:"bytes_read,omitempty"`
	BytesWritten int64  `json:"bytes_written,omitempty"`
	Error        string `json:"error,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// JSON OUTPUT

func (ctx AVIOContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVIOContext{
		IsEOF:        ctx.eof_reached != 0,
		IsWriteable:  ctx.write_flag != 0,
		IsSeekable:   ctx.seekable != 0,
		IsDirect:     ctx.direct != 0,
		Pos:          int64(ctx.pos),
		BufferSize:   int(ctx.buffer_size),
		BytesRead:    int64(ctx.bytes_read),
		BytesWritten: int64(ctx.bytes_written),
		Error:        AVError(ctx.error).Error(),
	})
}

func (ctx AVFormatContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVFormatContext{
		Pb:     (*AVIOContext)(ctx.pb),
		Input:  (*AVInputFormat)(ctx.iformat),
		Output: (*AVOutputFormat)(ctx.oformat),
		Url:    C.GoString(ctx.url),
	})
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx AVIOContext) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

func (ctx AVFormatContext) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}
