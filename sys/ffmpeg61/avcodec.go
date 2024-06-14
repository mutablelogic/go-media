package ffmpeg

import (
	"encoding/json"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec libavutil
#include <libavcodec/avcodec.h>
#include <libavutil/opt.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVPacket             C.AVPacket
	AVCodec              C.AVCodec
	AVCodecContext       C.AVCodecContext
	AVCodecParameters    C.AVCodecParameters
	AVCodecParser        C.AVCodecParser
	AVCodecParserContext C.AVCodecParserContext
	AVProfile            C.AVProfile
	AVCodecID            C.enum_AVCodecID
)

type jsonAVPacket struct {
	Pts           int64 `json:"pts,omitempty"`
	Dts           int64 `json:"dts,omitempty"`
	Size          int   `json:"size,omitempty"`
	StreamIndex   int   `json:"stream_index"` // Stream index starts at 0
	Flags         int   `json:"flags,omitempty"`
	SideDataElems int   `json:"side_data_elems,omitempty"`
	Duration      int64 `json:"duration,omitempty"`
	Pos           int64 `json:"pos,omitempty"`
}

type jsonAVCodecParameters struct {
	CodecType AVMediaType `json:"codec_type,omitempty"`
	CodecID   AVCodecID   `json:"codec_id,omitempty"`
	CodecTag  uint32      `json:"codec_tag,omitempty"`
	Format    int         `json:"format,omitempty"`
	BitRate   int64       `json:"bit_rate,omitempty"`
}

type jsonAVCodec struct {
	Name         string      `json:"name,omitempty"`
	LongName     string      `json:"long_name,omitempty"`
	Type         AVMediaType `json:"type,omitempty"`
	ID           AVCodecID   `json:"id,omitempty"`
	Capabilities int         `json:"capabilities,omitempty"`
}

type jsonAVCodecContext struct {
	Class            *AVClass        `json:"class,omitempty"`
	CodecType        AVMediaType     `json:"codec_type,omitempty"`
	Codec            *AVCodec        `json:"codec,omitempty"`
	BitRate          int64           `json:"bit_rate,omitempty"`
	BitRateTolerance int             `json:"bit_rate_tolerance,omitempty"`
	TimeBase         AVRational      `json:"time_base,omitempty"`
	Width            int             `json:"width,omitempty"`
	Height           int             `json:"height,omitempty"`
	PixelFormat      AVPixelFormat   `json:"pix_fmt,omitempty"`
	SampleFormat     AVSampleFormat  `json:"sample_fmt,omitempty"`
	SampleRate       int             `json:"sample_rate,omitempty"`
	ChannelLayout    AVChannelLayout `json:"channel_layout,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_CODEC_ID_NONE       AVCodecID = C.AV_CODEC_ID_NONE
	AV_CODEC_ID_MP2        AVCodecID = C.AV_CODEC_ID_MP2
	AV_CODEC_ID_H264       AVCodecID = C.AV_CODEC_ID_H264
	AV_CODEC_ID_MPEG1VIDEO AVCodecID = C.AV_CODEC_ID_MPEG1VIDEO
	AV_CODEC_ID_MPEG2VIDEO AVCodecID = C.AV_CODEC_ID_MPEG2VIDEO
)

/**
 * Required number of additionally allocated bytes at the end of the input bitstream for decoding.
 * This is mainly needed because some optimized bitstream readers read
 * 32 or 64 bit at once and could read over the end.
 * Note: If the first 23 bits of the additional bytes are not 0, then damaged
 * MPEG bitstreams could cause overread and segfault.
 */
const (
	AV_INPUT_BUFFER_PADDING_SIZE = 64
)

////////////////////////////////////////////////////////////////////////////////
// JSON OUTPUT

func (ctx *AVPacket) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVPacket{
		Pts:           int64(ctx.pts),
		Dts:           int64(ctx.dts),
		Size:          int(ctx.size),
		StreamIndex:   int(ctx.stream_index),
		Flags:         int(ctx.flags),
		SideDataElems: int(ctx.side_data_elems),
		Duration:      int64(ctx.duration),
		Pos:           int64(ctx.pos),
	})
}

func (ctx *AVCodecParameters) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVCodecParameters{
		CodecType: AVMediaType(ctx.codec_type),
		CodecID:   AVCodecID(ctx.codec_id),
		CodecTag:  uint32(ctx.codec_tag),
		Format:    int(ctx.format),
		BitRate:   int64(ctx.bit_rate),
	})
}

func (ctx *AVCodec) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVCodec{
		Name:         C.GoString(ctx.name),
		LongName:     C.GoString(ctx.long_name),
		Type:         AVMediaType(ctx._type),
		ID:           AVCodecID(ctx.id),
		Capabilities: int(ctx.capabilities),
	})
}

func (ctx *AVCodecContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVCodecContext{
		Class:            (*AVClass)(ctx.av_class),
		CodecType:        AVMediaType(ctx.codec_type),
		Codec:            (*AVCodec)(ctx.codec),
		BitRate:          int64(ctx.bit_rate),
		BitRateTolerance: int(ctx.bit_rate_tolerance),
		TimeBase:         (AVRational)(ctx.time_base),
		Width:            int(ctx.width),
		Height:           int(ctx.height),
		PixelFormat:      AVPixelFormat(ctx.pix_fmt),
		SampleFormat:     AVSampleFormat(ctx.sample_fmt),
		SampleRate:       int(ctx.sample_rate),
		ChannelLayout:    AVChannelLayout(ctx.ch_layout),
	})
}

func (ctx AVProfile) MarshalJSON() ([]byte, error) {
	return json.Marshal(ctx.Name())
}

func (ctx AVMediaType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ctx.String())
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVPacket) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

func (ctx *AVCodec) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

func (ctx *AVCodecParameters) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

func (ctx *AVCodecContext) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

func (ctx AVProfile) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}

////////////////////////////////////////////////////////////////////////////////
// AVCodecParameters

func (ctx *AVCodecParameters) CodecType() AVMediaType {
	return AVMediaType(ctx.codec_type)
}

func (ctx *AVCodecParameters) CodecID() AVCodecID {
	return AVCodecID(ctx.codec_id)
}

func (ctx *AVCodecParameters) CodecTag() uint32 {
	return uint32(ctx.codec_tag)
}

func (ctx *AVCodecParameters) SetCodecTag(tag uint32) {
	ctx.codec_tag = C.uint32_t(tag)
}

////////////////////////////////////////////////////////////////////////////////
// AVCodec

func (c *AVCodec) Name() string {
	return C.GoString(c.name)
}

func (c *AVCodec) LongName() string {
	return C.GoString(c.long_name)
}

func (c *AVCodec) Type() AVMediaType {
	return AVMediaType(c._type)
}

func (c *AVCodec) ID() AVCodecID {
	return AVCodecID(c.id)
}

func (c *AVCodec) SupportedFramerates() []AVRational {
	var result []AVRational
	ptr := uintptr(unsafe.Pointer(c.supported_framerates))
	if ptr == 0 {
		return nil
	}
	for {
		v := AVRational(*(*C.struct_AVRational)(unsafe.Pointer(ptr)))
		if v.IsZero() {
			break
		}
		result = append(result, v)
		ptr += unsafe.Sizeof(AVRational{})
	}
	return result
}

func (c *AVCodec) SampleFormats() []AVSampleFormat {
	var result []AVSampleFormat
	ptr := uintptr(unsafe.Pointer(c.sample_fmts))
	if ptr == 0 {
		return nil
	}
	for {
		v := AVSampleFormat(*(*C.enum_AVSampleFormat)(unsafe.Pointer(ptr)))
		if v == AV_SAMPLE_FMT_NONE {
			break
		}
		result = append(result, v)
		ptr += unsafe.Sizeof(AV_SAMPLE_FMT_NONE)
	}
	return result
}

func (c *AVCodec) PixelFormats() []AVPixelFormat {
	var result []AVPixelFormat
	ptr := uintptr(unsafe.Pointer(c.pix_fmts))
	if ptr == 0 {
		return nil
	}
	for {
		v := AVPixelFormat(*(*C.enum_AVPixelFormat)(unsafe.Pointer(ptr)))
		if v == AV_PIX_FMT_NONE {
			break
		}
		result = append(result, v)
		ptr += unsafe.Sizeof(AV_PIX_FMT_NONE)
	}
	return result
}

func (c *AVCodec) SupportedSamplerates() []int {
	var result []int
	ptr := uintptr(unsafe.Pointer(c.supported_samplerates))
	if ptr == 0 {
		return nil
	}
	for {
		v := int(*(*C.int)(unsafe.Pointer(ptr)))
		if v == 0 {
			break
		}
		result = append(result, v)
		ptr += unsafe.Sizeof(C.int(0))
	}
	return result
}

func (c *AVCodec) Profiles() []AVProfile {
	var result []AVProfile
	ptr := uintptr(unsafe.Pointer(c.profiles))
	if ptr == 0 {
		return nil
	}
	for {
		v := (AVProfile)(*(*C.struct_AVProfile)(unsafe.Pointer(ptr)))
		if v.profile == C.FF_PROFILE_UNKNOWN {
			break
		}
		result = append(result, v)
		ptr += unsafe.Sizeof(AVProfile{})
	}
	return result
}

func (c *AVCodec) ChannelLayouts() []AVChannelLayout {
	var result []AVChannelLayout
	ptr := uintptr(unsafe.Pointer(c.ch_layouts))
	if ptr == 0 {
		return nil
	}
	for {
		v := (AVChannelLayout)(*(*C.struct_AVChannelLayout)(unsafe.Pointer(ptr)))
		if v.nb_channels == 0 {
			break
		}
		result = append(result, v)
		ptr += unsafe.Sizeof(AVChannelLayout{})
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////
// AVCodecContext

func (ctx *AVCodecContext) BitRate() int64 {
	return int64(ctx.bit_rate)
}

func (ctx *AVCodecContext) SetBitRate(bit_rate int64) {
	ctx.bit_rate = C.int64_t(bit_rate)
}

func (ctx *AVCodecContext) Width() int {
	return int(ctx.width)
}

func (ctx *AVCodecContext) SetWidth(width int) {
	ctx.width = C.int(width)
}

func (ctx *AVCodecContext) Height() int {
	return int(ctx.height)
}

func (ctx *AVCodecContext) SetHeight(height int) {
	ctx.height = C.int(height)
}

func (ctx *AVCodecContext) TimeBase() AVRational {
	return (AVRational)(ctx.time_base)
}

func (ctx *AVCodecContext) SetTimeBase(time_base AVRational) {
	ctx.time_base = C.struct_AVRational(time_base)
}

func (ctx *AVCodecContext) Framerate() AVRational {
	return (AVRational)(ctx.framerate)
}

func (ctx *AVCodecContext) SetFramerate(framerate AVRational) {
	ctx.framerate = C.struct_AVRational(framerate)
}

// Audio sample format.
func (ctx *AVCodecContext) SampleFormat() AVSampleFormat {
	return AVSampleFormat(ctx.sample_fmt)
}

// Audio sample format.
func (ctx *AVCodecContext) SetSampleFormat(sample_fmt AVSampleFormat) {
	ctx.sample_fmt = C.enum_AVSampleFormat(sample_fmt)
}

// Audio sample rate.
func (ctx *AVCodecContext) SampleRate() int {
	return int(ctx.sample_rate)
}

// Audio sample rate.
func (ctx *AVCodecContext) SetSampleRate(sample_rate int) {
	ctx.sample_rate = C.int(sample_rate)
}

// Number of samples per channel in an audio frame.
func (ctx *AVCodecContext) FrameSize() int {
	return int(ctx.frame_size)
}

// Audio channel layout.
func (ctx *AVCodecContext) ChannelLayout() AVChannelLayout {
	return AVChannelLayout(ctx.ch_layout)
}

// Audio channel layout.
func (ctx *AVCodecContext) SetChannelLayout(src AVChannelLayout) error {
	if ret := AVError(C.av_channel_layout_copy((*C.struct_AVChannelLayout)(&ctx.ch_layout), (*C.struct_AVChannelLayout)(&src))); ret != 0 {
		return ret
	}
	return nil
}

// Group-of-pictures (GOP) size.
func (ctx *AVCodecContext) GopSize() int {
	return int(ctx.gop_size)
}

// Group-of-pictures (GOP) size.
func (ctx *AVCodecContext) SetGopSize(gop_size int) {
	ctx.gop_size = C.int(gop_size)
}

// Maximum number of B-frames between non-B-frames.
func (ctx *AVCodecContext) MaxBFrames() int {
	return int(ctx.max_b_frames)
}

// Maximum number of B-frames between non-B-frames.
func (ctx *AVCodecContext) SetMaxBFrames(max_b_frames int) {
	ctx.max_b_frames = C.int(max_b_frames)
}

// Pixel format.
func (ctx *AVCodecContext) PixFmt() AVPixelFormat {
	return AVPixelFormat(ctx.pix_fmt)
}

// Pixel format.
func (ctx *AVCodecContext) SetPixFmt(pix_fmt AVPixelFormat) {
	ctx.pix_fmt = C.enum_AVPixelFormat(pix_fmt)
}

// Private data, set key/value pair
func (ctx *AVCodecContext) SetPrivDataKV(name, value string) error {
	cName, cValue := C.CString(name), C.CString(value)
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cValue))
	if ret := AVError(C.av_opt_set(ctx.priv_data, cName, cValue, 0)); ret != 0 {
		return ret
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PROFILE

func (c *AVProfile) ID() int {
	return int(c.profile)
}

func (c *AVProfile) Name() string {
	return C.GoString(c.name)
}
