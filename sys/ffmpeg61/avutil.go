package ffmpeg

import (
	"encoding/json"
	"fmt"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#include <libavutil/dict.h>
#include <libavutil/samplefmt.h>
#include <libavutil/pixdesc.h>
#include <libavutil/channel_layout.h>

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
	AVBufferRef        C.struct_AVBufferRef
	AVChannel          C.enum_AVChannel
	AVChannelLayout    C.AVChannelLayout
	AVChannelOrder     C.enum_AVChannelOrder
	AVClass            C.AVClass
	AVDictionary       struct{ ctx *C.struct_AVDictionary } // Wrapper
	AVDictionaryEntry  C.struct_AVDictionaryEntry
	AVDictionaryFlag   C.int
	AVError            C.int
	AVFrame            C.struct_AVFrame
	AVLog              C.int
	AVMediaType        C.enum_AVMediaType
	AVRational         C.AVRational
	AVPictureType      C.enum_AVPictureType
	AVPixelFormat      C.enum_AVPixelFormat
	AVPixFmtDescriptor C.AVPixFmtDescriptor
	AVRounding         C.enum_AVRounding
	AVSampleFormat     C.enum_AVSampleFormat
)

type jsonAVClass struct {
	ClassName string `json:"class_name"`
}

type jsonAVDictionary struct {
	Count int                  `json:"count"`
	Elems []*AVDictionaryEntry `json:"elems,omitempty"`
}

type jsonAVDictionaryEntry struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// Only get an entry with exact-case key match.
	AV_DICT_MATCH_CASE AVDictionaryFlag = C.AV_DICT_MATCH_CASE

	// Return first entry in a dictionary whose first part corresponds to the search key, ignoring the suffix of the found key string.
	AV_DICT_IGNORE_SUFFIX AVDictionaryFlag = C.AV_DICT_IGNORE_SUFFIX

	// Take ownership of  key that has been allocated with av_malloc()
	AV_DICT_DONT_STRDUP_KEY AVDictionaryFlag = C.AV_DICT_DONT_STRDUP_KEY

	// Take ownership of  value that has been allocated with av_malloc()
	AV_DICT_DONT_STRDUP_VAL AVDictionaryFlag = C.AV_DICT_DONT_STRDUP_VAL

	// Don't overwrite existing entries.
	AV_DICT_DONT_OVERWRITE AVDictionaryFlag = C.AV_DICT_DONT_OVERWRITE

	// Append to existing key.
	AV_DICT_APPEND AVDictionaryFlag = C.AV_DICT_APPEND

	// Allow to store several equal keys in the dictionary.
	AV_DICT_MULTIKEY AVDictionaryFlag = C.AV_DICT_MULTIKEY
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
	AV_CHANNEL_ORDER_UNSPEC    AVChannelOrder = C.AV_CHANNEL_ORDER_UNSPEC
	AV_CHANNEL_ORDER_NATIVE    AVChannelOrder = C.AV_CHANNEL_ORDER_NATIVE
	AV_CHANNEL_ORDER_CUSTOM    AVChannelOrder = C.AV_CHANNEL_ORDER_CUSTOM
	AV_CHANNEL_ORDER_AMBISONIC AVChannelOrder = C.AV_CHANNEL_ORDER_AMBISONIC
)

const (
	AV_NOPTS_VALUE = C.AV_NOPTS_VALUE ///< Undefined timestamp value
)

const (
	AV_ROUND_ZERO        AVRounding = C.AV_ROUND_ZERO        // Round toward zero.
	AV_ROUND_INF         AVRounding = C.AV_ROUND_INF         // Round away from zero.
	AV_ROUND_DOWN        AVRounding = C.AV_ROUND_DOWN        // Round toward -infinity.
	AV_ROUND_UP          AVRounding = C.AV_ROUND_UP          // Round toward +infinity.
	AV_ROUND_NEAR_INF    AVRounding = C.AV_ROUND_NEAR_INF    // Round to nearest and halfway cases away from zero.
	AV_ROUND_PASS_MINMAX AVRounding = C.AV_ROUND_PASS_MINMAX // Flag to pass INT64_MIN/MAX through instead of rescaling, this avoids special cases for AV_NOPTS_VALUE
)

const (
	AV_TIME_BASE = C.AV_TIME_BASE // Internal time base
)

const (
	AV_CHAN_NONE AVChannel = C.AV_CHAN_NONE // Invalid channel
)

////////////////////////////////////////////////////////////////////////////////
// JSON OUTPUT

func (ctx *AVClass) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVClass{
		ClassName: C.GoString(ctx.class_name),
	})
}

func (ctx *AVDictionary) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVDictionary{
		Count: AVUtil_dict_count(ctx),
		Elems: AVUtil_dict_entries(ctx),
	})
}

func (ctx *AVDictionaryEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVDictionaryEntry{
		Key:   ctx.Key(),
		Value: ctx.Value(),
	})
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVClass) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

func (ctx *AVDictionary) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
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
	}
	return fmt.Sprintf("AVChannelOrder(%d)", int(v))
}
