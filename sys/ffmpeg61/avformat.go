package ffmpeg

import (
	"encoding/json"
	"fmt"
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
	AVDisposition   C.int
	AVFormat        C.int
	AVFormatContext C.struct_AVFormatContext
	AVFormatFlag    C.int
	AVInputFormat   C.struct_AVInputFormat
	AVIOContext     C.struct_AVIOContext
	AVIOFlag        C.int
	AVOutputFormat  C.struct_AVOutputFormat
	AVStream        C.struct_AVStream
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

const (
	AV_DISPOSITION_DEFAULT          AVDisposition = C.AV_DISPOSITION_DEFAULT
	AV_DISPOSITION_DUB              AVDisposition = C.AV_DISPOSITION_DUB
	AV_DISPOSITION_ORIGINAL         AVDisposition = C.AV_DISPOSITION_ORIGINAL
	AV_DISPOSITION_COMMENT          AVDisposition = C.AV_DISPOSITION_COMMENT
	AV_DISPOSITION_LYRICS           AVDisposition = C.AV_DISPOSITION_LYRICS
	AV_DISPOSITION_KARAOKE          AVDisposition = C.AV_DISPOSITION_KARAOKE
	AV_DISPOSITION_FORCED           AVDisposition = C.AV_DISPOSITION_FORCED
	AV_DISPOSITION_HEARING_IMPAIRED AVDisposition = C.AV_DISPOSITION_HEARING_IMPAIRED
	AV_DISPOSITION_VISUAL_IMPAIRED  AVDisposition = C.AV_DISPOSITION_VISUAL_IMPAIRED
	AV_DISPOSITION_CLEAN_EFFECTS    AVDisposition = C.AV_DISPOSITION_CLEAN_EFFECTS
	AV_DISPOSITION_ATTACHED_PIC     AVDisposition = C.AV_DISPOSITION_ATTACHED_PIC
	AV_DISPOSITION_TIMED_THUMBNAILS AVDisposition = C.AV_DISPOSITION_TIMED_THUMBNAILS
	AV_DISPOSITION_CAPTIONS         AVDisposition = C.AV_DISPOSITION_CAPTIONS
	AV_DISPOSITION_DESCRIPTIONS     AVDisposition = C.AV_DISPOSITION_DESCRIPTIONS
	AV_DISPOSITION_METADATA         AVDisposition = C.AV_DISPOSITION_METADATA
	AV_DISPOSITION_MIN                            = AV_DISPOSITION_DEFAULT
	AV_DISPOSITION_MAX                            = AV_DISPOSITION_METADATA
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
	NumStreams uint            `json:"nb_streams,omitempty"`
	Streams    []*AVStream     `json:"streams,omitempty"`
	StartTime  int64           `json:"start_time,omitempty"`
	Duration   int64           `json:"duration,omitempty"`
	BitRate    int64           `json:"bit_rate,omitempty"`
	PacketSize uint            `json:"packet_size,omitempty"`
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
		PacketSize: uint(ctx.packet_size),
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

func (v AVDisposition) String() string {
	if v == 0 {
		return ""
	}
	str := ""
	for f := AV_DISPOSITION_MIN; f <= AV_DISPOSITION_MAX; f <<= 1 {
		if v&f != 0 {
			str += "|" + f.FlagString()
		}
	}
	return str[1:]
}

func (v AVDisposition) FlagString() string {
	switch v {
	case AV_DISPOSITION_DEFAULT:
		return "DEFAULT"
	case AV_DISPOSITION_DUB:
		return "DUB"
	case AV_DISPOSITION_ORIGINAL:
		return "ORIGINAL"
	case AV_DISPOSITION_COMMENT:
		return "COMMENT"
	case AV_DISPOSITION_LYRICS:
		return "LYRICS"
	case AV_DISPOSITION_KARAOKE:
		return "KARAOKE"
	case AV_DISPOSITION_FORCED:
		return "FORCED"
	case AV_DISPOSITION_HEARING_IMPAIRED:
		return "HEARING_IMPAIRED"
	case AV_DISPOSITION_VISUAL_IMPAIRED:
		return "VISUAL_IMPAIRED"
	case AV_DISPOSITION_CLEAN_EFFECTS:
		return "CLEAN_EFFECTS"
	case AV_DISPOSITION_ATTACHED_PIC:
		return "ATTACHED_PIC"
	case AV_DISPOSITION_TIMED_THUMBNAILS:
		return "TIMED_THUMBNAILS"
	case AV_DISPOSITION_CAPTIONS:
		return "CAPTIONS"
	case AV_DISPOSITION_DESCRIPTIONS:
		return "DESCRIPTIONS"
	case AV_DISPOSITION_METADATA:
		return "METADATA"
	default:
		return fmt.Sprintf("AVDisposition(0x%08X)", int(v))
	}
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

func (ctx *AVFormatContext) Duration() int64 {
	return int64(ctx.duration)
}

////////////////////////////////////////////////////////////////////////////////
// AVFormatFlag

const (
	AVFMT_FLAG_NONE            AVFormatFlag = 0
	AVFMT_FLAG_GENPTS          AVFormatFlag = C.AVFMT_FLAG_GENPTS          ///< Generate missing pts even if it requires parsing future frames.
	AVFMT_FLAG_IGNIDX          AVFormatFlag = C.AVFMT_FLAG_IGNIDX          ///< Ignore index.
	AVFMT_FLAG_NONBLOCK        AVFormatFlag = C.AVFMT_FLAG_NONBLOCK        ///< Do not block when reading packets from input.
	AVFMT_FLAG_IGNDTS          AVFormatFlag = C.AVFMT_FLAG_IGNDTS          ///< Ignore DTS on frames that contain both DTS & PTS
	AVFMT_FLAG_NOFILLIN        AVFormatFlag = C.AVFMT_FLAG_NOFILLIN        ///< Do not infer any values from other values, just return what is stored in the container
	AVFMT_FLAG_NOPARSE         AVFormatFlag = C.AVFMT_FLAG_NOPARSE         ///< Do not use AVParsers, you also must set AVFMT_FLAG_NOFILLIN as the fillin code works on frames and no parsing -> no frames. Also seeking to frames can not work if parsing to find frame boundaries has been disabled
	AVFMT_FLAG_NOBUFFER        AVFormatFlag = C.AVFMT_FLAG_NOBUFFER        ///< Do not buffer frames when possible
	AVFMT_FLAG_CUSTOM_IO       AVFormatFlag = C.AVFMT_FLAG_CUSTOM_IO       ///< The caller has supplied a custom AVIOContext, don't avio_close() it.
	AVFMT_FLAG_DISCARD_CORRUPT AVFormatFlag = C.AVFMT_FLAG_DISCARD_CORRUPT ///< Discard frames marked corrupted
	AVFMT_FLAG_FLUSH_PACKETS   AVFormatFlag = C.AVFMT_FLAG_FLUSH_PACKETS   ///< Flush the AVIOContext every packet.
	AVFMT_FLAG_BITEXACT        AVFormatFlag = C.AVFMT_FLAG_BITEXACT        // When muxing, try to avoid writing any random/volatile data to the output.
	AVFMT_FLAG_SORT_DTS        AVFormatFlag = C.AVFMT_FLAG_SORT_DTS        ///< try to interleave outputted packets by dts (using this flag can slow demuxing down)
	AVFMT_FLAG_FAST_SEEK       AVFormatFlag = C.AVFMT_FLAG_FAST_SEEK       ///< Enable fast, but inaccurate seeks for some formats
	AVFMT_FLAG_SHORTEST        AVFormatFlag = C.AVFMT_FLAG_SHORTEST        ///< Stop muxing when the shortest stream stops.
	AVFMT_FLAG_AUTO_BSF        AVFormatFlag = C.AVFMT_FLAG_AUTO_BSF        ///< Add bitstream filters as requested by the muxer
	AVFMT_FLAG_MIN                          = AVFMT_FLAG_GENPTS
	AVFMT_FLAG_MAX                          = AVFMT_FLAG_AUTO_BSF
)

func (f AVFormatFlag) FlagString() string {
	switch f {
	case AVFMT_FLAG_NONE:
		return "AVFMT_FLAG_NONE"
	case AVFMT_FLAG_GENPTS:
		return "AVFMT_FLAG_GENPTS"
	case AVFMT_FLAG_IGNIDX:
		return "AVFMT_FLAG_IGNIDX"
	case AVFMT_FLAG_NONBLOCK:
		return "AVFMT_FLAG_NONBLOCK"
	case AVFMT_FLAG_IGNDTS:
		return "AVFMT_FLAG_IGNDTS"
	case AVFMT_FLAG_NOFILLIN:
		return "AVFMT_FLAG_NOFILLIN"
	case AVFMT_FLAG_NOPARSE:
		return "AVFMT_FLAG_NOPARSE"
	case AVFMT_FLAG_NOBUFFER:
		return "AVFMT_FLAG_NOBUFFER"
	case AVFMT_FLAG_CUSTOM_IO:
		return "AVFMT_FLAG_CUSTOM_IO"
	case AVFMT_FLAG_DISCARD_CORRUPT:
		return "AVFMT_FLAG_DISCARD_CORRUPT"
	case AVFMT_FLAG_FLUSH_PACKETS:
		return "AVFMT_FLAG_FLUSH_PACKETS"
	case AVFMT_FLAG_BITEXACT:
		return "AVFMT_FLAG_BITEXACT"
	case AVFMT_FLAG_SORT_DTS:
		return "AVFMT_FLAG_SORT_DTS"
	case AVFMT_FLAG_FAST_SEEK:
		return "AVFMT_FLAG_FAST_SEEK"
	case AVFMT_FLAG_SHORTEST:
		return "AVFMT_FLAG_SHORTEST"
	case AVFMT_FLAG_AUTO_BSF:
		return "AVFMT_FLAG_AUTO_BSF"
	default:
		return fmt.Sprintf("AVFormatFlag(0x%06X)", int(f))
	}
}

func (f AVFormatFlag) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.String())
}

func (f AVFormatFlag) String() string {
	if f == AVFMT_FLAG_NONE {
		return f.FlagString()
	}
	str := ""
	for i := AVFMT_FLAG_MIN; i <= AVFMT_FLAG_MAX; i <<= 1 {
		if f&i != 0 {
			str += "|" + i.FlagString()
		}
	}
	return str[1:]
}

func (f AVFormatFlag) Is(flag AVFormatFlag) bool {
	return f&flag == flag
}

////////////////////////////////////////////////////////////////////////////////
// AVDisposition

func (v AVDisposition) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (f AVDisposition) Is(flag AVDisposition) bool {
	return f&flag == flag
}
