package ffmpeg

import (
	"fmt"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#include <libavutil/dict.h>
#include <libavutil/rational.h>
#include <libavutil/samplefmt.h>
#include <libavutil/channel_layout.h>
#include <libavutil/pixfmt.h>

AVChannelLayout _AV_CHANNEL_LAYOUT_MONO = AV_CHANNEL_LAYOUT_MONO;
AVChannelLayout _AV_CHANNEL_LAYOUT_STEREO = AV_CHANNEL_LAYOUT_STEREO;
AVChannelLayout _AV_CHANNEL_LAYOUT_2POINT1 = AV_CHANNEL_LAYOUT_2POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_2_1 = AV_CHANNEL_LAYOUT_2_1;
AVChannelLayout _AV_CHANNEL_LAYOUT_SURROUND = AV_CHANNEL_LAYOUT_SURROUND;
AVChannelLayout _AV_CHANNEL_LAYOUT_3POINT1 = AV_CHANNEL_LAYOUT_3POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_4POINT0 = AV_CHANNEL_LAYOUT_4POINT0;
AVChannelLayout _AV_CHANNEL_LAYOUT_4POINT1 = AV_CHANNEL_LAYOUT_4POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_2_2 = AV_CHANNEL_LAYOUT_2_2;
AVChannelLayout _AV_CHANNEL_LAYOUT_QUAD = AV_CHANNEL_LAYOUT_QUAD;
AVChannelLayout _AV_CHANNEL_LAYOUT_5POINT0 = AV_CHANNEL_LAYOUT_5POINT0;
AVChannelLayout _AV_CHANNEL_LAYOUT_5POINT1 = AV_CHANNEL_LAYOUT_5POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_5POINT0_BACK = AV_CHANNEL_LAYOUT_5POINT0_BACK;
AVChannelLayout _AV_CHANNEL_LAYOUT_5POINT1_BACK = AV_CHANNEL_LAYOUT_5POINT1_BACK;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT0 = AV_CHANNEL_LAYOUT_6POINT0;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT0_FRONT = AV_CHANNEL_LAYOUT_6POINT0_FRONT;
AVChannelLayout _AV_CHANNEL_LAYOUT_HEXAGONAL = AV_CHANNEL_LAYOUT_HEXAGONAL;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT1 = AV_CHANNEL_LAYOUT_6POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT1_BACK = AV_CHANNEL_LAYOUT_6POINT1_BACK;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT1_FRONT = AV_CHANNEL_LAYOUT_6POINT1_FRONT;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT0 = AV_CHANNEL_LAYOUT_7POINT0;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT0_FRONT = AV_CHANNEL_LAYOUT_7POINT0_FRONT;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT1 = AV_CHANNEL_LAYOUT_7POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT1_WIDE = AV_CHANNEL_LAYOUT_7POINT1_WIDE;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT1_WIDE_BACK = AV_CHANNEL_LAYOUT_7POINT1_WIDE_BACK;
AVChannelLayout _AV_CHANNEL_LAYOUT_OCTAGONAL = AV_CHANNEL_LAYOUT_OCTAGONAL;
AVChannelLayout _AV_CHANNEL_LAYOUT_HEXADECAGONAL = AV_CHANNEL_LAYOUT_HEXADECAGONAL;
AVChannelLayout _AV_CHANNEL_LAYOUT_STEREO_DOWNMIX = AV_CHANNEL_LAYOUT_STEREO_DOWNMIX;
AVChannelLayout _AV_CHANNEL_LAYOUT_22POINT2 = AV_CHANNEL_LAYOUT_22POINT2;
AVChannelLayout _AV_CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER = AV_CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER;
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVError           C.int
	AVClass           C.struct_AVClass
	AVLogLevel        C.int
	AVLogCallback     func(AVLogLevel, string, uintptr)
	AVDictionaryEntry C.struct_AVDictionaryEntry
	AVDictionaryFlag  int
	AVDictionary      C.struct_AVDictionary
	AVRational        C.struct_AVRational
	AVSampleFormat    C.enum_AVSampleFormat
	AVChannelOrder    C.enum_AVChannelOrder
	AVChannelCustom   C.struct_AVChannelCustom
	AVChannel         C.enum_AVChannel
	AVChannelLayout   C.struct_AVChannelLayout
	AVPixelFormat     C.enum_AVPixelFormat
	AVRounding        C.enum_AVRounding
	AVMediaType       C.enum_AVMediaType
	AVFrame           C.struct_AVFrame
	AVPictureType     C.enum_AVPictureType
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

var (
	AV_CHANNEL_LAYOUT_MONO                  = AVChannelLayout(C._AV_CHANNEL_LAYOUT_MONO)
	AV_CHANNEL_LAYOUT_STEREO                = AVChannelLayout(C._AV_CHANNEL_LAYOUT_STEREO)
	AV_CHANNEL_LAYOUT_2POINT1               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_2POINT1)
	AV_CHANNEL_LAYOUT_2_1                   = AVChannelLayout(C._AV_CHANNEL_LAYOUT_2_1)
	AV_CHANNEL_LAYOUT_SURROUND              = AVChannelLayout(C._AV_CHANNEL_LAYOUT_SURROUND)
	AV_CHANNEL_LAYOUT_3POINT1               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_3POINT1)
	AV_CHANNEL_LAYOUT_4POINT0               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_4POINT0)
	AV_CHANNEL_LAYOUT_4POINT1               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_4POINT1)
	AV_CHANNEL_LAYOUT_2_2                   = AVChannelLayout(C._AV_CHANNEL_LAYOUT_2_2)
	AV_CHANNEL_LAYOUT_QUAD                  = AVChannelLayout(C._AV_CHANNEL_LAYOUT_QUAD)
	AV_CHANNEL_LAYOUT_5POINT0               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_5POINT0)
	AV_CHANNEL_LAYOUT_5POINT1               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_5POINT1)
	AV_CHANNEL_LAYOUT_5POINT0_BACK          = AVChannelLayout(C._AV_CHANNEL_LAYOUT_5POINT0_BACK)
	AV_CHANNEL_LAYOUT_5POINT1_BACK          = AVChannelLayout(C._AV_CHANNEL_LAYOUT_5POINT1_BACK)
	AV_CHANNEL_LAYOUT_6POINT0               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT0)
	AV_CHANNEL_LAYOUT_6POINT0_FRONT         = AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT0_FRONT)
	AV_CHANNEL_LAYOUT_HEXAGONAL             = AVChannelLayout(C._AV_CHANNEL_LAYOUT_HEXAGONAL)
	AV_CHANNEL_LAYOUT_6POINT1               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT1)
	AV_CHANNEL_LAYOUT_6POINT1_BACK          = AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT1_BACK)
	AV_CHANNEL_LAYOUT_6POINT1_FRONT         = AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT1_FRONT)
	AV_CHANNEL_LAYOUT_7POINT0               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT0)
	AV_CHANNEL_LAYOUT_7POINT0_FRONT         = AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT0_FRONT)
	AV_CHANNEL_LAYOUT_7POINT1               = AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT1)
	AV_CHANNEL_LAYOUT_7POINT1_WIDE          = AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT1_WIDE)
	AV_CHANNEL_LAYOUT_7POINT1_WIDE_BACK     = AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT1_WIDE_BACK)
	AV_CHANNEL_LAYOUT_OCTAGONAL             = AVChannelLayout(C._AV_CHANNEL_LAYOUT_OCTAGONAL)
	AV_CHANNEL_LAYOUT_HEXADECAGONAL         = AVChannelLayout(C._AV_CHANNEL_LAYOUT_HEXADECAGONAL)
	AV_CHANNEL_LAYOUT_STEREO_DOWNMIX        = AVChannelLayout(C._AV_CHANNEL_LAYOUT_STEREO_DOWNMIX)
	AV_CHANNEL_LAYOUT_22POINT2              = AVChannelLayout(C._AV_CHANNEL_LAYOUT_22POINT2)
	AV_CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER = AVChannelLayout(C._AV_CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER)
)

const (
	AV_ROUND_ZERO        = C.AV_ROUND_ZERO        ///< Round toward zero.
	AV_ROUND_INF         = C.AV_ROUND_INF         ///< Round away from zero.
	AV_ROUND_DOWN        = C.AV_ROUND_DOWN        ///< Round toward -infinity.
	AV_ROUND_UP          = C.AV_ROUND_UP          ///< Round toward +infinity.
	AV_ROUND_NEAR_INF    = C.AV_ROUND_NEAR_INF    ///< Round to nearest and halfway cases away from zero.
	AV_ROUND_PASS_MINMAX = C.AV_ROUND_PASS_MINMAX ///< Flag telling rescaling functions to pass INT64_MIN/MAX through unchanged
)

const (
	AVMEDIA_TYPE_UNKNOWN    AVMediaType = C.AVMEDIA_TYPE_UNKNOWN ///< Usually treated as AVMEDIA_TYPE_DATA
	AVMEDIA_TYPE_VIDEO      AVMediaType = C.AVMEDIA_TYPE_VIDEO
	AVMEDIA_TYPE_AUDIO      AVMediaType = C.AVMEDIA_TYPE_AUDIO
	AVMEDIA_TYPE_DATA       AVMediaType = C.AVMEDIA_TYPE_DATA ///< Opaque data information usually continuous
	AVMEDIA_TYPE_SUBTITLE   AVMediaType = C.AVMEDIA_TYPE_SUBTITLE
	AVMEDIA_TYPE_ATTACHMENT AVMediaType = C.AVMEDIA_TYPE_ATTACHMENT ///< Opaque data information usually sparse
)

const (
	AV_PICTURE_TYPE_NONE AVPictureType = C.AV_PICTURE_TYPE_NONE ///< Undefined
	AV_PICTURE_TYPE_I    AVPictureType = C.AV_PICTURE_TYPE_I    ///< Intra
	AV_PICTURE_TYPE_P    AVPictureType = C.AV_PICTURE_TYPE_P    ///< Predicted
	AV_PICTURE_TYPE_B    AVPictureType = C.AV_PICTURE_TYPE_B    ///< Bi-dir predicted
	AV_PICTURE_TYPE_S    AVPictureType = C.AV_PICTURE_TYPE_S    ///< S(GMC)-VOP MPEG-4
	AV_PICTURE_TYPE_SI   AVPictureType = C.AV_PICTURE_TYPE_SI   ///< Switching Intra
	AV_PICTURE_TYPE_SP   AVPictureType = C.AV_PICTURE_TYPE_SP   ///< Switching Predicted
	AV_PICTURE_TYPE_BI   AVPictureType = C.AV_PICTURE_TYPE_BI   ///< BI type
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v AVPictureType) String() string {
	switch v {
	case AV_PICTURE_TYPE_NONE:
		return "AV_PICTURE_TYPE_NONE"
	case AV_PICTURE_TYPE_I:
		return "AV_PICTURE_TYPE_I"
	case AV_PICTURE_TYPE_P:
		return "AV_PICTURE_TYPE_P"
	case AV_PICTURE_TYPE_B:
		return "AV_PICTURE_TYPE_B"
	case AV_PICTURE_TYPE_S:
		return "AV_PICTURE_TYPE_S"
	case AV_PICTURE_TYPE_SI:
		return "AV_PICTURE_TYPE_SI"
	case AV_PICTURE_TYPE_SP:
		return "AV_PICTURE_TYPE_SP"
	case AV_PICTURE_TYPE_BI:
		return "AV_PICTURE_TYPE_BI"
	default:
		return "[?? Invalid AVPictureType value]"
	}
}

func (v AVMediaType) String() string {
	switch v {
	case AVMEDIA_TYPE_UNKNOWN:
		return "AVMEDIA_TYPE_UNKNOWN"
	case AVMEDIA_TYPE_VIDEO:
		return "AVMEDIA_TYPE_VIDEO"
	case AVMEDIA_TYPE_AUDIO:
		return "AVMEDIA_TYPE_AUDIO"
	case AVMEDIA_TYPE_DATA:
		return "AVMEDIA_TYPE_DATA"
	case AVMEDIA_TYPE_SUBTITLE:
		return "AVMEDIA_TYPE_SUBTITLE"
	case AVMEDIA_TYPE_ATTACHMENT:
		return "AVMEDIA_TYPE_ATTACHMENT"
	default:
		return "[?? Invalid AVMediaType value]"
	}
}

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

func (f *AVFrame) String() string {
	str := "<AVFrame"
	if sample_fmt := f.SampleFormat(); sample_fmt != AV_SAMPLE_FMT_NONE {
		str += fmt.Sprint(" sample_format=", sample_fmt)
		if sample_rate := f.SampleRate(); sample_rate > 0 {
			str += fmt.Sprint(" sample_rate=", sample_rate)
		}
		if c := f.Channels(); c > 0 {
			str += fmt.Sprint(" channels=", c)
		}
		if n := f.NumSamples(); n > 0 {
			str += fmt.Sprint(" nb_samples=", n)
		}
	}
	if pix_fmt := f.PixelFormat(); pix_fmt != AV_PIX_FMT_NONE {
		str += fmt.Sprint(" pixel_format=", pix_fmt)
		if w, h := f.Width(), f.Height(); w >= 0 && h >= 0 {
			str += fmt.Sprint(" size={", w, ",", h, "}")
		}
		if pict_type := f.PictType(); pict_type != AV_PICTURE_TYPE_NONE {
			str += fmt.Sprint(" pict_type=", pict_type)
		}
		if f.IsKeyFrame() {
			str += " key_frame"
		}
		if f.IsInterlaced() {
			str += " interlaced"
		}
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// ACCESSORS - FRAME

func (f *AVFrame) Data(ch int) *byte {
	return (*byte)(unsafe.Pointer(f.data[ch]))
}

func (f *AVFrame) LineSize(ch int) int {
	return int(f.linesize[ch])
}

func (f *AVFrame) NumSamples() int {
	return int(f.nb_samples)
}

func (f *AVFrame) SampleRate() int {
	return int(f.sample_rate)
}

func (f *AVFrame) PixelFormat() AVPixelFormat {
	if f.format == -1 {
		return AV_PIX_FMT_NONE
	} else if f.channels != 0 {
		return AV_PIX_FMT_NONE
	} else {
		return AVPixelFormat(f.format)
	}
}

func (f *AVFrame) SampleFormat() AVSampleFormat {
	if f.format == -1 {
		return AV_SAMPLE_FMT_NONE
	} else if f.channels == 0 {
		return AV_SAMPLE_FMT_NONE
	} else {
		return AVSampleFormat(f.format)
	}
}

func (f *AVFrame) PictType() AVPictureType {
	return AVPictureType(f.pict_type)
}

func (f *AVFrame) ChannelLayout() AVChannelLayout {
	return AVChannelLayout(f.ch_layout)
}

func (f *AVFrame) Channels() int {
	return int(f.channels)
}

func (f *AVFrame) IsPlanar() bool {
	if fmt := f.SampleFormat(); fmt == AV_SAMPLE_FMT_NONE {
		return false
	} else {
		return AVUtil_av_sample_fmt_is_planar(fmt)
	}
}

func (f *AVFrame) IsInterlaced() bool {
	return intToBool(int(f.interlaced_frame))
}

func (f *AVFrame) IsKeyFrame() bool {
	return intToBool(int(f.key_frame))
}

func (f *AVFrame) Width() int {
	return int(f.width)
}

func (f *AVFrame) Height() int {
	return int(f.height)
}

/*

func (this *AVFrame) PictType() AVPictureType {
	ctx := (*C.AVFrame)(unsafe.Pointer(this))
	return AVPictureType(ctx.pict_type)
}

func (this *AVFrame) PictWidth() int {
	ctx := (*C.AVFrame)(unsafe.Pointer(this))
	return int(ctx.width)
}

func (this *AVFrame) PictHeight() int {
	ctx := (*C.AVFrame)(unsafe.Pointer(this))
	return int(ctx.height)
}

func (this *AVFrame) Buffer(plane int) *AVBufferRef {
	ctx := (*C.AVFrame)(this)
	if buf := (C.av_frame_get_plane_buffer(ctx, C.int(plane))); buf == nil {
		return nil
	} else {
		return (*AVBufferRef)(buf)
	}
}

func (this *AVFrame) StrideForPlane(i int) int {
	ctx := (*C.AVFrame)(unsafe.Pointer(this))
	return int(ctx.linesize[i])
}

func (this *AVFrame) GetAudioBuffer(num_samples int) error {
	ctx := (*C.AVFrame)(unsafe.Pointer(this))

	ctx.nb_samples = C.int(num_samples)
	if err := AVError(C.av_frame_get_buffer(ctx, 0)); err != 0 {
		return err
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// AVBufferRef

func (this *AVBufferRef) Data() []byte {
	var bytes []byte

	ctx := (*C.AVBufferRef)(this)
	if ctx.data == nil {
		return nil
	}
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&bytes)))
	sliceHeader.Cap = int(ctx.size)
	sliceHeader.Len = int(ctx.size)
	sliceHeader.Data = uintptr(unsafe.Pointer(ctx.data))
	return bytes
}

func (this *AVBufferRef) Size() int {
	ctx := (*C.AVBufferRef)(this)
	return int(ctx.size)
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *AVBufferRef) String() string {
	str := "<AVBufferRef"
	str += " size=" + fmt.Sprint(this.Size())
	return str + ">"
}

*/
