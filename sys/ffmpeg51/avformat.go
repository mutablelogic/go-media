package ffmpeg

import (
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
	AVInputFormat   C.struct_AVInputFormat
	AVOutputFormat  C.struct_AVOutputFormat
	AVFormatFlag    C.int
	AVFormatContext C.struct_AVFormatContext
	AVContextFlags  C.int
	AVStream        C.struct_AVStream
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AVFMT_NONE AVFormatFlag = 0
	// Demuxer will use avio_open, no opened file should be provided by the caller.
	AVFMT_NOFILE AVFormatFlag = C.AVFMT_NOFILE
	// Needs '%d' in filename.
	AVFMT_NEEDNUMBER AVFormatFlag = C.AVFMT_NEEDNUMBER
	// The muxer/demuxer is experimental and should be used with caution
	AVFMT_EXPERIMENTAL AVFormatFlag = C.AVFMT_EXPERIMENTAL
	// Show format stream IDs numbers.
	AVFMT_SHOWIDS AVFormatFlag = C.AVFMT_SHOW_IDS
	// Format wants global header.
	AVFMT_GLOBALHEADER AVFormatFlag = C.AVFMT_GLOBALHEADER
	// Format does not need / have any timestamps.
	AVFMT_NOTIMESTAMPS AVFormatFlag = C.AVFMT_NOTIMESTAMPS
	// Use generic index building code.
	AVFMT_GENERICINDEX AVFormatFlag = C.AVFMT_GENERIC_INDEX
	// Format allows timestamp discontinuities. Note, muxers always require valid (monotone) timestamps
	AVFMT_TSDISCONT AVFormatFlag = C.AVFMT_TS_DISCONT
	// Format allows variable fps.
	AVFMT_VARIABLEFPS AVFormatFlag = C.AVFMT_VARIABLE_FPS
	// Format does not need width/height
	AVFMT_NODIMENSIONS AVFormatFlag = C.AVFMT_NODIMENSIONS
	// Format does not require any streams
	AVFMT_NOSTREAMS AVFormatFlag = C.AVFMT_NOSTREAMS
	// Format does not allow to fall back on binary search via read_timestamp
	AVFMT_NOBINSEARCH AVFormatFlag = C.AVFMT_NOBINSEARCH
	// Format does not allow to fall back on generic search
	AVFMT_NOGENSEARCH AVFormatFlag = C.AVFMT_NOGENSEARCH
	// Format does not allow seeking by bytes
	AVFMT_NOBYTESEEK AVFormatFlag = C.AVFMT_NO_BYTE_SEEK
	// Format allows flushing. If not set, the muxer will not receive a NULL packet in the write_packet function.
	AVFMT_ALLOWFLUSH AVFormatFlag = C.AVFMT_ALLOW_FLUSH
	// Format does not require strictly increasing timestamps, but they must still be monotonic
	AVFMT_TS_NONSTRICT AVFormatFlag = C.AVFMT_TS_NONSTRICT
	// Format allows muxing negative timestamps
	AVFMT_TS_NEGATIVE AVFormatFlag = C.AVFMT_TS_NEGATIVE
	// Min
	AVFMT_MIN AVFormatFlag = AVFMT_NOFILE
	// Max
	AVFMT_MAX AVFormatFlag = AVFMT_TS_NEGATIVE
)

const (
	AVFMTCTX_NONE       AVContextFlags = 0
	AVFMTCTX_NOHEADER   AVContextFlags = C.AVFMTCTX_NOHEADER   // signal that no header is present (streams are added dynamically)
	AVFMTCTX_UNSEEKABLE AVContextFlags = C.AVFMTCTX_UNSEEKABLE // signal that the stream is definitely not seekable
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVStream) String() string {
	str := "<AVStream"
	return str + ">"
}

func (ctx *AVFormatContext) String() string {
	str := "<AVFormatContext"
	if input := ctx.Input(); input != nil {
		str += fmt.Sprint(" input=", input)
	}
	if output := ctx.Output(); output != nil {
		str += fmt.Sprint(" output=", output)
	}
	if flags := ctx.ContextFlags(); flags != AVFMTCTX_NONE {
		str += fmt.Sprint(" ctx_flags=", flags)
	}
	if num_streams := ctx.NumStreams(); num_streams != 0 {
		str += fmt.Sprint(" nb_streams=", num_streams)
	}
	if streams := ctx.Streams(); len(streams) != 0 {
		str += fmt.Sprint(" streams=", streams)
	}
	if url := ctx.Url(); url != "" {
		str += fmt.Sprintf(" url=%q", url)
	}
	if metadata := ctx.Metadata(); metadata != nil {
		str += fmt.Sprint(" metadata=", metadata)
	}
	if start_time := ctx.StartTime(); start_time != 0 {
		str += fmt.Sprint(" start_time=", start_time)
	}
	if duration := ctx.Duration(); duration != 0 {
		str += fmt.Sprint(" duration=", duration)
	}
	if bit_rate := ctx.BitRate(); bit_rate != 0 {
		str += fmt.Sprint(" bit_rate=", bit_rate)
	}
	if packet_size := ctx.PacketSize(); packet_size != 0 {
		str += fmt.Sprint(" packet_size=", packet_size)
	}
	if max_delay := ctx.MaxDelay(); max_delay >= 0 {
		str += fmt.Sprint(" max_delay=", max_delay)
	}
	if flags := ctx.Flags(); flags != AVFMT_NONE {
		str += fmt.Sprint(" flags=", flags)
	}
	if probesize := ctx.ProbeSize(); probesize != 0 {
		str += fmt.Sprint(" probesize=", probesize)
	}
	if max_analyze_duration := ctx.MaxAnalyzeDuration(); max_analyze_duration != 0 {
		str += fmt.Sprint(" max_analyze_duration=", max_analyze_duration)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - FORMAT CONTEXT

func (ctx *AVFormatContext) Class() *AVClass {
	return (*AVClass)(ctx.av_class)
}

func (ctx *AVFormatContext) Input() *AVInputFormat {
	return (*AVInputFormat)(ctx.iformat)
}

func (ctx *AVFormatContext) Output() *AVOutputFormat {
	return (*AVOutputFormat)(ctx.oformat)
}

func (ctx *AVFormatContext) ContextFlags() AVContextFlags {
	return (AVContextFlags)(ctx.ctx_flags)
}

func (ctx *AVFormatContext) NumStreams() uint {
	return (uint)(ctx.nb_streams)
}

func (ctx *AVFormatContext) Streams() []*AVStream {
	return (*[1 << 28]*AVStream)(unsafe.Pointer(ctx.streams))[:ctx.nb_streams:ctx.nb_streams]
}

func (ctx *AVFormatContext) Url() string {
	return C.GoString(ctx.url)
}

func (ctx *AVFormatContext) StartTime() int64 {
	return int64(ctx.start_time)
}

func (ctx *AVFormatContext) Duration() int64 {
	return int64(ctx.duration)
}

func (ctx *AVFormatContext) BitRate() int64 {
	return int64(ctx.bit_rate)
}

func (ctx *AVFormatContext) PacketSize() uint {
	return uint(ctx.packet_size)
}

func (ctx *AVFormatContext) MaxDelay() int {
	return int(ctx.max_delay)
}

func (ctx *AVFormatContext) Flags() AVFormatFlag {
	return AVFormatFlag(ctx.flags)
}

func (ctx *AVFormatContext) ProbeSize() int64 {
	return int64(ctx.probesize)
}

func (ctx *AVFormatContext) Metadata() *AVDictionary {
	return (*AVDictionary)(ctx.metadata)
}

func (ctx *AVFormatContext) MaxAnalyzeDuration() int64 {
	return int64(ctx.max_analyze_duration)
}

func (ctx *AVFormatContext) VideoCodecID() AVCodecID {
	return AVCodecID(ctx.video_codec_id)
}

func (ctx *AVFormatContext) AudioCodecID() AVCodecID {
	return AVCodecID(ctx.audio_codec_id)
}

func (ctx *AVFormatContext) SubtitleCodecID() AVCodecID {
	return AVCodecID(ctx.subtitle_codec_id)
}

func (ctx *AVFormatContext) DataCodecID() AVCodecID {
	return AVCodecID(ctx.data_codec_id)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - INPUT

func (this *AVInputFormat) Name() string {
	return C.GoString(this.name)
}

func (this *AVInputFormat) Description() string {
	return C.GoString(this.long_name)
}

func (this *AVInputFormat) Ext() string {
	return C.GoString(this.extensions)
}

func (this *AVInputFormat) MimeType() string {
	return C.GoString(this.mime_type)
}

func (this *AVInputFormat) Flags() AVFormatFlag {
	return AVFormatFlag(this.flags)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - OUTPUT

func (this *AVOutputFormat) Name() string {
	return C.GoString(this.name)
}

func (this *AVOutputFormat) Description() string {
	return C.GoString(this.long_name)
}

func (this *AVOutputFormat) Ext() string {
	return C.GoString(this.extensions)
}

func (this *AVOutputFormat) MimeType() string {
	return C.GoString(this.mime_type)
}

func (this *AVOutputFormat) Flags() AVFormatFlag {
	return AVFormatFlag(this.flags)
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *AVInputFormat) String() string {
	str := "<AVInputFormat"
	if name := this.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if description := this.Description(); description != "" {
		str += fmt.Sprintf(" description=%q", description)
	}
	if ext := this.Ext(); ext != "" {
		str += fmt.Sprintf(" ext=%q", ext)
	}
	if mimeType := this.MimeType(); mimeType != "" {
		str += fmt.Sprintf(" mime_type=%q", mimeType)
	}
	if flags := this.Flags(); flags != 0 {
		str += fmt.Sprint(" flags=", flags)
	}
	return str + ">"
}

func (this *AVOutputFormat) String() string {
	str := "<AVOutputFormat"
	if name := this.Name(); name != "" {
		str += fmt.Sprintf(" name=%q", name)
	}
	if description := this.Description(); description != "" {
		str += fmt.Sprintf(" description=%q", description)
	}
	if ext := this.Ext(); ext != "" {
		str += fmt.Sprintf(" ext=%q", ext)
	}
	if mimeType := this.MimeType(); mimeType != "" {
		str += fmt.Sprintf(" mime_type=%q", mimeType)
	}
	if flags := this.Flags(); flags != 0 {
		str += fmt.Sprint(" flags=", flags)
	}
	return str + ">"
}

func (f AVFormatFlag) String() string {
	if f == AVFMT_NONE {
		return f.FlagString()
	}
	str := ""
	for i := AVFMT_MIN; i <= AVFMT_MAX; i <<= 1 {
		if f&i != 0 {
			str += "|" + i.FlagString()
		}
	}
	return str[1:]
}

func (f AVFormatFlag) FlagString() string {
	switch f {
	case AVFMT_NONE:
		return "AVFMT_NONE"
	case AVFMT_NOFILE:
		return "AVFMT_NOFILE"
	case AVFMT_NEEDNUMBER:
		return "AVFMT_NEEDNUMBER"
	case AVFMT_EXPERIMENTAL:
		return "AVFMT_EXPERIMENTAL"
	case AVFMT_SHOWIDS:
		return "AVFMT_SHOWIDS"
	case AVFMT_GLOBALHEADER:
		return "AVFMT_GLOBALHEADER"
	case AVFMT_NOTIMESTAMPS:
		return "AVFMT_NOTIMESTAMPS"
	case AVFMT_GENERICINDEX:
		return "AVFMT_GENERICINDEX"
	case AVFMT_TSDISCONT:
		return "AVFMT_TSDISCONT"
	case AVFMT_VARIABLEFPS:
		return "AVFMT_VARIABLEFPS"
	case AVFMT_NODIMENSIONS:
		return "AVFMT_NODIMENSIONS"
	case AVFMT_NOSTREAMS:
		return "AVFMT_NOSTREAMS"
	case AVFMT_NOBINSEARCH:
		return "AVFMT_NOBINSEARCH"
	case AVFMT_NOGENSEARCH:
		return "AVFMT_NOGENSEARCH"
	case AVFMT_NOBYTESEEK:
		return "AVFMT_NOBYTESEEK"
	case AVFMT_ALLOWFLUSH:
		return "AVFMT_ALLOWFLUSH"
	case AVFMT_TS_NONSTRICT:
		return "AVFMT_TS_NONSTRICT"
	case AVFMT_TS_NEGATIVE:
		return "AVFMT_TS_NEGATIVE"
	default:
		return "[?? Invalid AVFormatFlag value]"
	}
}
