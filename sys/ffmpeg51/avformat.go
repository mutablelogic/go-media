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
	AVFormat        C.int
	AVFormatFlag    C.int
	AVFormatContext C.struct_AVFormatContext
	AVContextFlags  C.int
	AVStream        C.struct_AVStream
	AVDisposition   C.int
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

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

const (
	AVFMTCTX_NONE       AVContextFlags = 0
	AVFMTCTX_NOHEADER   AVContextFlags = C.AVFMTCTX_NOHEADER   // signal that no header is present (streams are added dynamically)
	AVFMTCTX_UNSEEKABLE AVContextFlags = C.AVFMTCTX_UNSEEKABLE // signal that the stream is definitely not seekable
)

const (
	AV_DISPOSITION_DEFAULT          AVDisposition = C.AV_DISPOSITION_DEFAULT
	AV_DISPOSITION_DUB              AVDisposition = C.AV_DISPOSITION_DUB
	AV_DISPOSITION_ORIGINAL         AVDisposition = C.AV_DISPOSITION_ORIGINAL
	AV_DISPOSITION_COMMENT          AVDisposition = C.AV_DISPOSITION_COMMENT
	AV_DISPOSITION_LYRICS           AVDisposition = C.AV_DISPOSITION_LYRICS
	AV_DISPOSITION_KARAOKE          AVDisposition = C.AV_DISPOSITION_KARAOKE
	AV_DISPOSITION_FORCED           AVDisposition = C.AV_DISPOSITION_FORCED
	AV_DISPOSITION_HEARING_IMPAIRED AVDisposition = C.AV_DISPOSITION_HEARING_IMPAIRED // Stream for hearing impaired audiences
	AV_DISPOSITION_VISUAL_IMPAIRED  AVDisposition = C.AV_DISPOSITION_VISUAL_IMPAIRED  // Stream for visual impaired audiences
	AV_DISPOSITION_CLEAN_EFFECTS    AVDisposition = C.AV_DISPOSITION_CLEAN_EFFECTS
	AV_DISPOSITION_ATTACHED_PIC     AVDisposition = C.AV_DISPOSITION_ATTACHED_PIC
	AV_DISPOSITION_TIMED_THUMBNAILS AVDisposition = C.AV_DISPOSITION_TIMED_THUMBNAILS
	AV_DISPOSITION_NON_DIEGETIC     AVDisposition = C.AV_DISPOSITION_NON_DIEGETIC
	AV_DISPOSITION_CAPTIONS         AVDisposition = C.AV_DISPOSITION_CAPTIONS
	AV_DISPOSITION_DESCRIPTIONS     AVDisposition = C.AV_DISPOSITION_DESCRIPTIONS
	AV_DISPOSITION_METADATA         AVDisposition = C.AV_DISPOSITION_METADATA
	AV_DISPOSITION_DEPENDENT        AVDisposition = C.AV_DISPOSITION_DEPENDENT
	AV_DISPOSITION_STILL_IMAGE      AVDisposition = C.AV_DISPOSITION_STILL_IMAGE
	AV_DISPOSITION_NONE             AVDisposition = 0
	AV_DISPOSITION_MAX                            = AV_DISPOSITION_STILL_IMAGE
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v AVDisposition) String() string {
	if v == AV_DISPOSITION_NONE {
		return v.FlagString()
	}
	str := ""
	for i := AVDisposition(1); i <= AV_DISPOSITION_MAX; i <<= 1 {
		if v&i != 0 {
			str += "|" + i.FlagString()
		}
	}
	return str[1:]
}

func (v AVDisposition) FlagString() string {
	switch v {
	case AV_DISPOSITION_NONE:
		return "AV_DISPOSITION_NONE"
	case AV_DISPOSITION_DEFAULT:
		return "AV_DISPOSITION_DEFAULT"
	case AV_DISPOSITION_DUB:
		return "AV_DISPOSITION_DUB"
	case AV_DISPOSITION_ORIGINAL:
		return "AV_DISPOSITION_ORIGINAL"
	case AV_DISPOSITION_COMMENT:
		return "AV_DISPOSITION_COMMENT"
	case AV_DISPOSITION_LYRICS:
		return "AV_DISPOSITION_LYRICS"
	case AV_DISPOSITION_KARAOKE:
		return "AV_DISPOSITION_KARAOKE"
	case AV_DISPOSITION_FORCED:
		return "AV_DISPOSITION_FORCED"
	case AV_DISPOSITION_HEARING_IMPAIRED:
		return "AV_DISPOSITION_HEARING_IMPAIRED"
	case AV_DISPOSITION_VISUAL_IMPAIRED:
		return "AV_DISPOSITION_VISUAL_IMPAIRED"
	case AV_DISPOSITION_CLEAN_EFFECTS:
		return "AV_DISPOSITION_CLEAN_EFFECTS"
	case AV_DISPOSITION_ATTACHED_PIC:
		return "AV_DISPOSITION_ATTACHED_PIC"
	case AV_DISPOSITION_TIMED_THUMBNAILS:
		return "AV_DISPOSITION_TIMED_THUMBNAILS"
	case AV_DISPOSITION_NON_DIEGETIC:
		return "AV_DISPOSITION_NON_DIEGETIC"
	case AV_DISPOSITION_CAPTIONS:
		return "AV_DISPOSITION_CAPTIONS"
	case AV_DISPOSITION_DESCRIPTIONS:
		return "AV_DISPOSITION_DESCRIPTIONS"
	case AV_DISPOSITION_METADATA:
		return "AV_DISPOSITION_METADATA"
	case AV_DISPOSITION_DEPENDENT:
		return "AV_DISPOSITION_DEPENDENT"
	case AV_DISPOSITION_STILL_IMAGE:
		return "AV_DISPOSITION_STILL_IMAGE"
	default:
		return fmt.Sprintf("AVDisposition(%d)", v)
	}
}

func (ctx *AVStream) String() string {
	str := "<AVStream"
	str += fmt.Sprint(" index=", ctx.Index())
	if id := ctx.ID(); id != 0 {
		str += fmt.Sprint(" id=", id)
	}
	if time_base := ctx.TimeBase(); time_base.den != 0 {
		str += fmt.Sprint(" time_base=", time_base)
	}
	if start_time := ctx.StartTime(); start_time != 0 {
		str += fmt.Sprint(" start_time=", start_time)
	}
	if duration := ctx.Duration(); duration != 0 {
		str += fmt.Sprint(" duration=", duration)
	}
	if nb_frames := ctx.NumFrames(); nb_frames != 0 {
		str += fmt.Sprint(" nb_frames=", nb_frames)
	}
	if disposition := ctx.Disposition(); disposition != AV_DISPOSITION_NONE {
		str += fmt.Sprint(" disposition=", disposition)
	}
	if codecpar := ctx.CodecPar(); codecpar != nil {
		str += fmt.Sprint(" codecpar=", codecpar)
	}
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
	if flags := ctx.Flags(); flags != AVFMT_FLAG_NONE {
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
// PUBLIC METHODS - STREAM

func (ctx *AVStream) Index() int {
	return int(ctx.index)
}

func (ctx *AVStream) ID() int {
	return int(ctx.id)
}

func (ctx *AVStream) TimeBase() AVRational {
	return AVRational(ctx.time_base)
}

func (ctx *AVStream) StartTime() int64 {
	return int64(ctx.start_time)
}

func (ctx *AVStream) Duration() int64 {
	return int64(ctx.duration)
}

func (ctx *AVStream) NumFrames() int64 {
	return int64(ctx.nb_frames)
}

func (ctx *AVStream) Disposition() AVDisposition {
	return AVDisposition(ctx.disposition)
}

func (ctx *AVStream) SampleAspectRatio() AVRational {
	return AVRational(ctx.sample_aspect_ratio)
}

func (ctx *AVStream) Metadata() *AVDictionary {
	return (*AVDictionary)(ctx.metadata)
}

func (ctx *AVStream) AverageFrameRate() AVRational {
	return AVRational(ctx.avg_frame_rate)
}

func (ctx *AVStream) RealFrameRate() AVRational {
	return AVRational(ctx.r_frame_rate)
}

func (ctx *AVStream) AttachedPic() AVPacket {
	return AVPacket(ctx.attached_pic)
}

func (ctx *AVStream) CodecPar() *AVCodecParameters {
	return (*AVCodecParameters)(ctx.codecpar)
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

func (this *AVInputFormat) Format() AVFormat {
	return AVFormat(this.flags)
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

func (this *AVOutputFormat) Format() AVFormat {
	return AVFormat(this.flags)
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
	if flags := this.Format(); flags != 0 {
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
	if flags := this.Format(); flags != 0 {
		str += fmt.Sprint(" flags=", flags)
	}
	return str + ">"
}

func (f AVFormat) String() string {
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
		return "[?? Invalid AVFormatFlag value]"
	}
}

func (f AVFormat) FlagString() string {
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
