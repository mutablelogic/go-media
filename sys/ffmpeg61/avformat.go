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
	AVInputFormat   C.struct_AVInputFormat
	AVOutputFormat  C.struct_AVOutputFormat
	AVFormat        C.int
	AVFormatFlag    C.int
	AVFormatContext C.struct_AVFormatContext
	AVIOContext     C.struct_AVIOContext
	AVStream        C.struct_AVStream
	AVIOFlag        C.int
	AVTimestamp     C.int64_t
)

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
// CONSTANTS

const (
	/**
	* ORing this as the "whence" parameter to a seek function causes it to
	* return the filesize without seeking anywhere. Supporting this is optional.
	* If it is not supported then the seek function will return <0.
	 */
	AVSEEK_SIZE = C.AVSEEK_SIZE

	/**
	 * Passing this flag as the "whence" parameter to a seek function causes it to
	 * seek by any means (like reopening and linear reading) or other normally unreasonable
	 * means that can be extremely slow.
	 * This may be ignored by the seek code.
	 */
	AVSEEK_FORCE = C.AVSEEK_FORCE
)

const (
	AVIO_FLAG_NONE       AVIOFlag = 0
	AVIO_FLAG_READ       AVIOFlag = C.AVIO_FLAG_READ
	AVIO_FLAG_WRITE      AVIOFlag = C.AVIO_FLAG_WRITE
	AVIO_FLAG_READ_WRITE AVIOFlag = C.AVIO_FLAG_READ_WRITE
)

////////////////////////////////////////////////////////////////////////////////
// JSON OUTPUT

func (ctx *AVIOContext) MarshalJSON() ([]byte, error) {
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

type jsonAVFormatContext struct {
	Pb         *AVIOContext    `json:"pb,omitempty"`
	Input      *AVInputFormat  `json:"input_format,omitempty"`
	Output     *AVOutputFormat `json:"output_format,omitempty"`
	Url        string          `json:"url,omitempty"`
	NumStreams uint            `json:"num_streams,omitempty"`
	Streams    []*AVStream     `json:"streams,omitempty"`
	StartTime  int64           `json:"start_time,omitempty"`
	Duration   int64           `json:"duration,omitempty"`
	BitRate    int64           `json:"bit_rate,omitempty"`
	Flags      AVFormatFlag    `json:"flags,omitempty"`
}

func (ctx *AVFormatContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVFormatContext{
		Pb:         (*AVIOContext)(ctx.pb),
		Input:      (*AVInputFormat)(ctx.iformat),
		Output:     (*AVOutputFormat)(ctx.oformat),
		Url:        C.GoString(ctx.url),
		NumStreams: uint(ctx.nb_streams),
		Streams:    ctx.Streams(),
		StartTime:  int64(ctx.start_time),
		Duration:   int64(ctx.duration),
		BitRate:    int64(ctx.bit_rate),
		Flags:      AVFormatFlag(ctx.flags),
	})
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVIOContext) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

func (ctx *AVFormatContext) String() string {
	data, _ := json.MarshalIndent(ctx, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// AVTimestamp

func (v AVTimestamp) MarshalJSON() ([]byte, error) {
	if v == AV_NOPTS_VALUE {
		return json.Marshal(nil)
	} else {
		return json.Marshal(int64(v))
	}
}

////////////////////////////////////////////////////////////////////////////////
// AVFormatContext functions

func (ctx *AVFormatContext) Input() *AVInputFormat {
	return (*AVInputFormat)(ctx.iformat)
}

func (ctx *AVFormatContext) Output() *AVOutputFormat {
	return (*AVOutputFormat)(ctx.oformat)
}

func (ctx *AVFormatContext) Metadata() *AVDictionary {
	return &AVDictionary{ctx.metadata}
}

func (ctx *AVFormatContext) SetPb(pb *AVIOContextEx) {
	if pb == nil {
		ctx.pb = nil
	} else {
		ctx.pb = (*C.struct_AVIOContext)(pb.AVIOContext)
	}
}

func (ctx *AVFormatContext) NumStreams() uint {
	return uint(ctx.nb_streams)
}

func (ctx *AVFormatContext) Streams() []*AVStream {
	return cAVStreamSlice(unsafe.Pointer(ctx.streams), C.int(ctx.nb_streams))
}

func (ctx *AVFormatContext) Stream(stream int) *AVStream {
	streams := ctx.Streams()
	if stream < 0 || stream >= len(streams) {
		return nil
	} else {
		return streams[stream]
	}
}

func (ctx *AVFormatContext) Flags() AVFormat {
	return AVFormat(ctx.flags)
}

////////////////////////////////////////////////////////////////////////////////
// AVFormatFlag

func (f AVFormatFlag) Is(flag AVFormatFlag) bool {
	return f&flag != 0
}
