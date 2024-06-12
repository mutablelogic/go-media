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
)

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

type jsonAVInputFormat struct {
	Name       string   `json:"name,omitempty"`
	LongName   string   `json:"long_name,omitempty"`
	Flags      AVFormat `json:"flags,omitempty"`
	Extensions string   `json:"extensions,omitempty"`
	MimeTypes  string   `json:"mime_types,omitempty"`
}

type jsonAVOutputFormat struct {
	Name       string   `json:"name,omitempty"`
	LongName   string   `json:"long_name,omitempty"`
	Flags      AVFormat `json:"flags,omitempty"`
	Extensions string   `json:"extensions,omitempty"`
	MimeTypes  string   `json:"mime_types,omitempty"`
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

type jsonAVStream struct {
	Index     int               `json:"index"`
	Id        int               `json:"id,omitempty"`
	CodecPar  AVCodecParameters `json:"codec_par,omitempty"`
	StartTime int64             `json:"start_time,omitempty"`
	Duration  int64             `json:"duration,omitempty"`
	NumFrames int64             `json:"num_frames,omitempty"`
	TimeBase  AVRational        `json:"time_base,omitempty"`
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
	AVFMT_NONE AVFormat = 0
	// Demuxer will use avio_open, no opened file should be provided by the caller.
	AVFMT_NOFILE AVFormat = C.AVFMT_NOFILE
	// Needs '%d' in filename.
	AVFMT_NEEDNUMBER AVFormat = C.AVFMT_NEEDNUMBER
	// The muxer/demuxer is experimental and should be used with caution
	AVFMT_EXPERIMENTAL AVFormat = C.AVFMT_EXPERIMENTAL
	// Show format stream IDs numbers.
	AVFMT_SHOWIDS AVFormat = C.AVFMT_SHOW_IDS
	// Format wants global header.
	AVFMT_GLOBALHEADER AVFormat = C.AVFMT_GLOBALHEADER
	// Format does not need / have any timestamps.
	AVFMT_NOTIMESTAMPS AVFormat = C.AVFMT_NOTIMESTAMPS
	// Use generic index building code.
	AVFMT_GENERICINDEX AVFormat = C.AVFMT_GENERIC_INDEX
	// Format allows timestamp discontinuities. Note, muxers always require valid (monotone) timestamps
	AVFMT_TSDISCONT AVFormat = C.AVFMT_TS_DISCONT
	// Format allows variable fps.
	AVFMT_VARIABLEFPS AVFormat = C.AVFMT_VARIABLE_FPS
	// Format does not need width/height
	AVFMT_NODIMENSIONS AVFormat = C.AVFMT_NODIMENSIONS
	// Format does not require any streams
	AVFMT_NOSTREAMS AVFormat = C.AVFMT_NOSTREAMS
	// Format does not allow to fall back on binary search via read_timestamp
	AVFMT_NOBINSEARCH AVFormat = C.AVFMT_NOBINSEARCH
	// Format does not allow to fall back on generic search
	AVFMT_NOGENSEARCH AVFormat = C.AVFMT_NOGENSEARCH
	// Format does not allow seeking by bytes
	AVFMT_NOBYTESEEK AVFormat = C.AVFMT_NO_BYTE_SEEK
	// Format allows flushing. If not set, the muxer will not receive a NULL packet in the write_packet function.
	AVFMT_ALLOWFLUSH AVFormat = C.AVFMT_ALLOW_FLUSH
	// Format does not require strictly increasing timestamps, but they must still be monotonic
	AVFMT_TS_NONSTRICT AVFormat = C.AVFMT_TS_NONSTRICT
	// Format allows muxing negative timestamps
	AVFMT_TS_NEGATIVE AVFormat = C.AVFMT_TS_NEGATIVE
	AVFMT_MIN         AVFormat = AVFMT_NOFILE
	AVFMT_MAX         AVFormat = AVFMT_TS_NEGATIVE
)

const (
	AVIO_FLAG_NONE       AVIOFlag = 0
	AVIO_FLAG_READ       AVIOFlag = C.AVIO_FLAG_READ
	AVIO_FLAG_WRITE      AVIOFlag = C.AVIO_FLAG_WRITE
	AVIO_FLAG_READ_WRITE AVIOFlag = C.AVIO_FLAG_READ_WRITE
)

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

func (ctx AVInputFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVInputFormat{
		Name:       C.GoString(ctx.name),
		LongName:   C.GoString(ctx.long_name),
		MimeTypes:  C.GoString(ctx.mime_type),
		Extensions: C.GoString(ctx.extensions),
		Flags:      AVFormat(ctx.flags),
	})
}

func (ctx AVOutputFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVOutputFormat{
		Name:       C.GoString(ctx.name),
		LongName:   C.GoString(ctx.long_name),
		MimeTypes:  C.GoString(ctx.mime_type),
		Extensions: C.GoString(ctx.extensions),
		Flags:      AVFormat(ctx.flags),
	})
}

func (ctx AVStream) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVStream{
		Index:     int(ctx.index),
		Id:        int(ctx.id),
		StartTime: int64(ctx.start_time),
		Duration:  int64(ctx.duration),
		NumFrames: int64(ctx.nb_frames),
		TimeBase:  AVRational(ctx.time_base),
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

func (ctx AVInputFormat) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

func (ctx AVOutputFormat) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

func (ctx AVStream) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
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
	return (*AVDictionary)(ctx.metadata)
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
	if ctx.streams == nil {
		return nil
	}
	return (*[1 << 30]*AVStream)(unsafe.Pointer(ctx.streams))[:]
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
// AVStream functions

func (ctx *AVStream) Index() int {
	return int(ctx.index)
}

func (ctx *AVStream) Id() int {
	return int(ctx.id)
}

func (ctx *AVStream) CodecPar() *AVCodecParameters {
	return (*AVCodecParameters)(ctx.codecpar)
}

func (ctx *AVStream) TimeBase() AVRational {
	return AVRational(ctx.time_base)
}

////////////////////////////////////////////////////////////////////////////////
// AVFormat

func (f AVFormat) Is(flag AVFormat) bool {
	return f&flag != 0
}

////////////////////////////////////////////////////////////////////////////////
// AVFormatFlag

func (f AVFormatFlag) Is(flag AVFormatFlag) bool {
	return f&flag != 0
}
