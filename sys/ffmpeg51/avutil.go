package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/log.h>
#include <libavutil/dict.h>
#include <libavutil/rational.h>
#include <libavutil/samplefmt.h>
#include <libavutil/channel_layout.h>
*/
import "C"

type (
	AVError           C.int
	AVClass           C.struct_AVClass
	AVLogLevel        C.int
	AVLogCallback     func(AVLogLevel, string, uintptr)
	AVDictionaryEntry C.struct_AVDictionaryEntry
	AVDictionaryFlag  int
	AVDictionary      struct {
		ctx *C.struct_AVDictionary
	}
	AVRational      C.struct_AVRational
	AVSampleFormat  C.enum_AVSampleFormat
	AVChannelOrder  C.enum_AVChannelOrder
	AVChannelCustom C.struct_AVChannelCustom
	AVChannel       C.enum_AVChannel
	AVChannelLayout C.struct_AVChannelLayout
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_LOG_QUIET   AVLogLevel = C.AV_LOG_QUIET
	AV_LOG_PANIC   AVLogLevel = C.AV_LOG_PANIC
	AV_LOG_FATAL   AVLogLevel = C.AV_LOG_FATAL
	AV_LOG_ERROR   AVLogLevel = C.AV_LOG_ERROR
	AV_LOG_WARNING AVLogLevel = C.AV_LOG_WARNING
	AV_LOG_INFO    AVLogLevel = C.AV_LOG_INFO
	AV_LOG_VERBOSE AVLogLevel = C.AV_LOG_VERBOSE
	AV_LOG_DEBUG   AVLogLevel = C.AV_LOG_DEBUG
	AV_LOG_TRACE   AVLogLevel = C.AV_LOG_TRACE
)

const (
	AV_DICT_MATCH_CASE      AVDictionaryFlag = C.AV_DICT_MATCH_CASE
	AV_DICT_IGNORE_SUFFIX   AVDictionaryFlag = C.AV_DICT_IGNORE_SUFFIX
	AV_DICT_DONT_STRDUP_KEY AVDictionaryFlag = C.AV_DICT_DONT_STRDUP_KEY
	AV_DICT_DONT_STRDUP_VAL AVDictionaryFlag = C.AV_DICT_DONT_STRDUP_VAL
	AV_DICT_DONT_OVERWRITE  AVDictionaryFlag = C.AV_DICT_DONT_OVERWRITE
	AV_DICT_APPEND          AVDictionaryFlag = C.AV_DICT_APPEND
	AV_DICT_MULTIKEY        AVDictionaryFlag = C.AV_DICT_MULTIKEY
)

const (
	AV_SAMPLE_FMT_NONE AVSampleFormat = C.AV_SAMPLE_FMT_NONE
	AV_SAMPLE_FMT_U8   AVSampleFormat = C.AV_SAMPLE_FMT_U8
	AV_SAMPLE_FMT_S16  AVSampleFormat = C.AV_SAMPLE_FMT_S16
	AV_SAMPLE_FMT_S32  AVSampleFormat = C.AV_SAMPLE_FMT_S32
	AV_SAMPLE_FMT_FLT  AVSampleFormat = C.AV_SAMPLE_FMT_FLT
	AV_SAMPLE_FMT_DBL  AVSampleFormat = C.AV_SAMPLE_FMT_DBL
	AV_SAMPLE_FMT_U8P  AVSampleFormat = C.AV_SAMPLE_FMT_U8P
	AV_SAMPLE_FMT_S16P AVSampleFormat = C.AV_SAMPLE_FMT_S16P
	AV_SAMPLE_FMT_S32P AVSampleFormat = C.AV_SAMPLE_FMT_S32P
	AV_SAMPLE_FMT_FLTP AVSampleFormat = C.AV_SAMPLE_FMT_FLTP
	AV_SAMPLE_FMT_DBLP AVSampleFormat = C.AV_SAMPLE_FMT_DBLP
	AV_SAMPLE_FMT_S64  AVSampleFormat = C.AV_SAMPLE_FMT_S64
	AV_SAMPLE_FMT_S64P AVSampleFormat = C.AV_SAMPLE_FMT_S64P
	AV_SAMPLE_FMT_NB   AVSampleFormat = C.AV_SAMPLE_FMT_NB
)

const (
	AV_CHANNEL_ORDER_UNSPEC    AVChannelOrder = C.AV_CHANNEL_ORDER_UNSPEC
	AV_CHANNEL_ORDER_NATIVE    AVChannelOrder = C.AV_CHANNEL_ORDER_NATIVE
	AV_CHANNEL_ORDER_CUSTOM    AVChannelOrder = C.AV_CHANNEL_ORDER_CUSTOM
	AV_CHANNEL_ORDER_AMBISONIC AVChannelOrder = C.AV_CHANNEL_ORDER_AMBISONIC
)

const (
	AV_CHAN_NONE                  AVChannel = C.AV_CHAN_NONE
	AV_CHAN_FRONT_LEFT            AVChannel = C.AV_CHAN_FRONT_LEFT
	AV_CHAN_FRONT_RIGHT           AVChannel = C.AV_CHAN_FRONT_RIGHT
	AV_CHAN_FRONT_CENTER          AVChannel = C.AV_CHAN_FRONT_CENTER
	AV_CHAN_LOW_FREQUENCY         AVChannel = C.AV_CHAN_LOW_FREQUENCY
	AV_CHAN_BACK_LEFT             AVChannel = C.AV_CHAN_BACK_LEFT
	AV_CHAN_BACK_RIGHT            AVChannel = C.AV_CHAN_BACK_RIGHT
	AV_CHAN_FRONT_LEFT_OF_CENTER  AVChannel = C.AV_CHAN_FRONT_LEFT_OF_CENTER
	AV_CHAN_FRONT_RIGHT_OF_CENTER AVChannel = C.AV_CHAN_FRONT_RIGHT_OF_CENTER
	AV_CHAN_BACK_CENTER           AVChannel = C.AV_CHAN_BACK_CENTER
	AV_CHAN_SIDE_LEFT             AVChannel = C.AV_CHAN_SIDE_LEFT
	AV_CHAN_SIDE_RIGHT            AVChannel = C.AV_CHAN_SIDE_RIGHT
	AV_CHAN_TOP_CENTER            AVChannel = C.AV_CHAN_TOP_CENTER
	AV_CHAN_TOP_FRONT_LEFT        AVChannel = C.AV_CHAN_TOP_FRONT_LEFT
	AV_CHAN_TOP_FRONT_CENTER      AVChannel = C.AV_CHAN_TOP_FRONT_CENTER
	AV_CHAN_TOP_FRONT_RIGHT       AVChannel = C.AV_CHAN_TOP_FRONT_RIGHT
	AV_CHAN_TOP_BACK_LEFT         AVChannel = C.AV_CHAN_TOP_BACK_LEFT
	AV_CHAN_TOP_BACK_CENTER       AVChannel = C.AV_CHAN_TOP_BACK_CENTER
	AV_CHAN_TOP_BACK_RIGHT        AVChannel = C.AV_CHAN_TOP_BACK_RIGHT
	AV_CHAN_STEREO_LEFT           AVChannel = C.AV_CHAN_STEREO_LEFT
	AV_CHAN_STEREO_RIGHT          AVChannel = C.AV_CHAN_STEREO_RIGHT
	AV_CHAN_WIDE_LEFT             AVChannel = C.AV_CHAN_WIDE_LEFT
	AV_CHAN_WIDE_RIGHT            AVChannel = C.AV_CHAN_WIDE_RIGHT
	AV_CHAN_SURROUND_DIRECT_LEFT  AVChannel = C.AV_CHAN_SURROUND_DIRECT_LEFT
	AV_CHAN_SURROUND_DIRECT_RIGHT AVChannel = C.AV_CHAN_SURROUND_DIRECT_RIGHT
	AV_CHAN_LOW_FREQUENCY_2       AVChannel = C.AV_CHAN_LOW_FREQUENCY_2
	AV_CHAN_TOP_SIDE_LEFT         AVChannel = C.AV_CHAN_TOP_SIDE_LEFT
	AV_CHAN_TOP_SIDE_RIGHT        AVChannel = C.AV_CHAN_TOP_SIDE_RIGHT
	AV_CHAN_BOTTOM_FRONT_CENTER   AVChannel = C.AV_CHAN_BOTTOM_FRONT_CENTER
	AV_CHAN_BOTTOM_FRONT_LEFT     AVChannel = C.AV_CHAN_BOTTOM_FRONT_LEFT
	AV_CHAN_BOTTOM_FRONT_RIGHT    AVChannel = C.AV_CHAN_BOTTOM_FRONT_RIGHT
	AV_CHAN_UNUSED                AVChannel = C.AV_CHAN_UNUSED
	AV_CHAN_UNKNOWN               AVChannel = C.AV_CHAN_UNKNOWN
	AV_CHAN_AMBISONIC_BASE        AVChannel = C.AV_CHAN_AMBISONIC_BASE
	AV_CHAN_AMBISONIC_END         AVChannel = C.AV_CHAN_AMBISONIC_END
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v AVLogLevel) String() string {
	switch v {
	case AV_LOG_QUIET:
		return "AV_LOG_QUIET"
	case AV_LOG_PANIC:
		return "AV_LOG_PANIC"
	case AV_LOG_FATAL:
		return "AV_LOG_FATAL"
	case AV_LOG_ERROR:
		return "AV_LOG_ERROR"
	case AV_LOG_WARNING:
		return "AV_LOG_WARNING"
	case AV_LOG_INFO:
		return "AV_LOG_INFO"
	case AV_LOG_VERBOSE:
		return "AV_LOG_VERBOSE"
	case AV_LOG_DEBUG:
		return "AV_LOG_DEBUG"
	case AV_LOG_TRACE:
		return "AV_LOG_TRACE"
	default:
		return "[?? Invalid AVLogLevel value]"
	}
}

func (v AVSampleFormat) String() string {
	switch v {
	case AV_SAMPLE_FMT_NONE:
		return "AV_SAMPLE_FMT_NONE"
	case AV_SAMPLE_FMT_U8:
		return "AV_SAMPLE_FMT_U8"
	case AV_SAMPLE_FMT_S16:
		return "AV_SAMPLE_FMT_S16"
	case AV_SAMPLE_FMT_S32:
		return "AV_SAMPLE_FMT_S32"
	case AV_SAMPLE_FMT_FLT:
		return "AV_SAMPLE_FMT_FLT"
	case AV_SAMPLE_FMT_DBL:
		return "AV_SAMPLE_FMT_DBL"
	case AV_SAMPLE_FMT_U8P:
		return "AV_SAMPLE_FMT_U8P"
	case AV_SAMPLE_FMT_S16P:
		return "AV_SAMPLE_FMT_S16P"
	case AV_SAMPLE_FMT_S32P:
		return "AV_SAMPLE_FMT_S32P"
	case AV_SAMPLE_FMT_FLTP:
		return "AV_SAMPLE_FMT_FLTP"
	case AV_SAMPLE_FMT_DBLP:
		return "AV_SAMPLE_FMT_DBLP"
	case AV_SAMPLE_FMT_S64:
		return "AV_SAMPLE_FMT_S64"
	case AV_SAMPLE_FMT_S64P:
		return "AV_SAMPLE_FMT_S64P"
	case AV_SAMPLE_FMT_NB:
		return "AV_SAMPLE_FMT_NB"
	default:
		return "[?? Invalid AVSampleFormat value]"
	}
}

func (v AVChannelOrder) String() string {
	switch v {
	case AV_CHANNEL_ORDER_UNSPEC:
		return "AV_CHANNEL_ORDER_UNSPEC"
	case AV_CHANNEL_ORDER_NATIVE:
		return "AV_CHANNEL_ORDER_NATIVE"
	case AV_CHANNEL_ORDER_CUSTOM:
		return "AV_CHANNEL_ORDER_CUSTOM"
	case AV_CHANNEL_ORDER_AMBISONIC:
		return "AV_CHANNEL_ORDER_AMBISONIC"
	default:
		return "[?? Invalid AVChannelOrder value]"
	}
}

func (v AVChannel) String() string {
	switch v {
	case AV_CHAN_NONE:
		return "AV_CHAN_NONE"
	case AV_CHAN_FRONT_LEFT:
		return "AV_CHAN_FRONT_LEFT"
	case AV_CHAN_FRONT_RIGHT:
		return "AV_CHAN_FRONT_RIGHT"
	case AV_CHAN_FRONT_CENTER:
		return "AV_CHAN_FRONT_CENTER"
	case AV_CHAN_LOW_FREQUENCY:
		return "AV_CHAN_LOW_FREQUENCY"
	case AV_CHAN_BACK_LEFT:
		return "AV_CHAN_BACK_LEFT"
	case AV_CHAN_BACK_RIGHT:
		return "AV_CHAN_BACK_RIGHT"
	case AV_CHAN_FRONT_LEFT_OF_CENTER:
		return "AV_CHAN_FRONT_LEFT_OF_CENTER"
	case AV_CHAN_FRONT_RIGHT_OF_CENTER:
		return "AV_CHAN_FRONT_RIGHT_OF_CENTER"
	case AV_CHAN_BACK_CENTER:
		return "AV_CHAN_BACK_CENTER"
	case AV_CHAN_SIDE_LEFT:
		return "AV_CHAN_SIDE_LEFT"
	case AV_CHAN_SIDE_RIGHT:
		return "AV_CHAN_SIDE_RIGHT"
	case AV_CHAN_TOP_CENTER:
		return "AV_CHAN_TOP_CENTER"
	case AV_CHAN_TOP_FRONT_LEFT:
		return "AV_CHAN_TOP_FRONT_LEFT"
	case AV_CHAN_TOP_FRONT_CENTER:
		return "AV_CHAN_TOP_FRONT_CENTER"
	case AV_CHAN_TOP_FRONT_RIGHT:
		return "AV_CHAN_TOP_FRONT_RIGHT"
	case AV_CHAN_TOP_BACK_LEFT:
		return "AV_CHAN_TOP_BACK_LEFT"
	case AV_CHAN_TOP_BACK_CENTER:
		return "AV_CHAN_TOP_BACK_CENTER"
	case AV_CHAN_TOP_BACK_RIGHT:
		return "AV_CHAN_TOP_BACK_RIGHT"
	case AV_CHAN_STEREO_LEFT:
		return "AV_CHAN_STEREO_LEFT"
	case AV_CHAN_STEREO_RIGHT:
		return "AV_CHAN_STEREO_RIGHT"
	case AV_CHAN_WIDE_LEFT:
		return "AV_CHAN_WIDE_LEFT"
	case AV_CHAN_WIDE_RIGHT:
		return "AV_CHAN_WIDE_RIGHT"
	case AV_CHAN_SURROUND_DIRECT_LEFT:
		return "AV_CHAN_SURROUND_DIRECT_LEFT"
	case AV_CHAN_SURROUND_DIRECT_RIGHT:
		return "AV_CHAN_SURROUND_DIRECT_RIGHT"
	case AV_CHAN_LOW_FREQUENCY_2:
		return "AV_CHAN_LOW_FREQUENCY_2"
	case AV_CHAN_TOP_SIDE_LEFT:
		return "AV_CHAN_TOP_SIDE_LEFT"
	case AV_CHAN_TOP_SIDE_RIGHT:
		return "AV_CHAN_TOP_SIDE_RIGHT"
	case AV_CHAN_BOTTOM_FRONT_CENTER:
		return "AV_CHAN_BOTTOM_FRONT_CENTER"
	case AV_CHAN_BOTTOM_FRONT_LEFT:
		return "AV_CHAN_BOTTOM_FRONT_LEFT"
	case AV_CHAN_BOTTOM_FRONT_RIGHT:
		return "AV_CHAN_BOTTOM_FRONT_RIGHT"
	case AV_CHAN_UNUSED:
		return "AV_CHAN_UNUSED"
	case AV_CHAN_UNKNOWN:
		return "AV_CHAN_UNKNOWN"
	case AV_CHAN_AMBISONIC_BASE:
		return "AV_CHAN_AMBISONIC_BASE"
	case AV_CHAN_AMBISONIC_END:
		return "AV_CHAN_AMBISONIC_END"
	default:
		return "[?? Invalid AVChannel value]"
	}
}
