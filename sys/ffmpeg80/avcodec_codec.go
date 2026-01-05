package ffmpeg

import (
	"encoding/json"
	"fmt"
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
	AVCodec           C.struct_AVCodec
	AVCodecCap        C.uint32_t
	AVCodecContext    C.struct_AVCodecContext
	AVCodecFlag       C.uint32_t
	AVCodecFlag2      C.uint32_t
	AVCodecID         C.enum_AVCodecID
	AVCodecParameters C.struct_AVCodecParameters
	AVProfile         C.struct_AVProfile
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_CODEC_ID_NONE       AVCodecID = C.AV_CODEC_ID_NONE
	AV_CODEC_ID_MP2        AVCodecID = C.AV_CODEC_ID_MP2
	AV_CODEC_ID_H264       AVCodecID = C.AV_CODEC_ID_H264
	AV_CODEC_ID_MPEG1VIDEO AVCodecID = C.AV_CODEC_ID_MPEG1VIDEO
	AV_CODEC_ID_MPEG2VIDEO AVCodecID = C.AV_CODEC_ID_MPEG2VIDEO
	AV_CODEC_ID_MJPEG      AVCodecID = C.AV_CODEC_ID_MJPEG
	AV_CODEC_ID_PNG        AVCodecID = C.AV_CODEC_ID_PNG
	AV_CODEC_ID_GIF        AVCodecID = C.AV_CODEC_ID_GIF
	AV_CODEC_ID_BMP        AVCodecID = C.AV_CODEC_ID_BMP
	AV_CODEC_ID_WEBP       AVCodecID = C.AV_CODEC_ID_WEBP
)

// Subtitle codec IDs
const (
	AV_CODEC_ID_FIRST_SUBTITLE     AVCodecID = C.AV_CODEC_ID_FIRST_SUBTITLE
	AV_CODEC_ID_DVD_SUBTITLE       AVCodecID = C.AV_CODEC_ID_DVD_SUBTITLE
	AV_CODEC_ID_DVB_SUBTITLE       AVCodecID = C.AV_CODEC_ID_DVB_SUBTITLE
	AV_CODEC_ID_TEXT               AVCodecID = C.AV_CODEC_ID_TEXT
	AV_CODEC_ID_XSUB               AVCodecID = C.AV_CODEC_ID_XSUB
	AV_CODEC_ID_SSA                AVCodecID = C.AV_CODEC_ID_SSA
	AV_CODEC_ID_MOV_TEXT           AVCodecID = C.AV_CODEC_ID_MOV_TEXT
	AV_CODEC_ID_HDMV_PGS_SUBTITLE  AVCodecID = C.AV_CODEC_ID_HDMV_PGS_SUBTITLE
	AV_CODEC_ID_DVB_TELETEXT       AVCodecID = C.AV_CODEC_ID_DVB_TELETEXT
	AV_CODEC_ID_SRT                AVCodecID = C.AV_CODEC_ID_SRT
	AV_CODEC_ID_MICRODVD           AVCodecID = C.AV_CODEC_ID_MICRODVD
	AV_CODEC_ID_EIA_608            AVCodecID = C.AV_CODEC_ID_EIA_608
	AV_CODEC_ID_JACOSUB            AVCodecID = C.AV_CODEC_ID_JACOSUB
	AV_CODEC_ID_SAMI               AVCodecID = C.AV_CODEC_ID_SAMI
	AV_CODEC_ID_REALTEXT           AVCodecID = C.AV_CODEC_ID_REALTEXT
	AV_CODEC_ID_STL                AVCodecID = C.AV_CODEC_ID_STL
	AV_CODEC_ID_SUBRIP             AVCodecID = C.AV_CODEC_ID_SUBRIP
	AV_CODEC_ID_SUBVIEWER1         AVCodecID = C.AV_CODEC_ID_SUBVIEWER1
	AV_CODEC_ID_SUBVIEWER          AVCodecID = C.AV_CODEC_ID_SUBVIEWER
	AV_CODEC_ID_SUBRIP_WEBVTT      AVCodecID = C.AV_CODEC_ID_SUBRIP // Alias
	AV_CODEC_ID_MPL2               AVCodecID = C.AV_CODEC_ID_MPL2
	AV_CODEC_ID_VPLAYER            AVCodecID = C.AV_CODEC_ID_VPLAYER
	AV_CODEC_ID_PJS                AVCodecID = C.AV_CODEC_ID_PJS
	AV_CODEC_ID_ASS                AVCodecID = C.AV_CODEC_ID_ASS
	AV_CODEC_ID_HDMV_TEXT_SUBTITLE AVCodecID = C.AV_CODEC_ID_HDMV_TEXT_SUBTITLE
	AV_CODEC_ID_TTML               AVCodecID = C.AV_CODEC_ID_TTML
	AV_CODEC_ID_ARIB_CAPTION       AVCodecID = C.AV_CODEC_ID_ARIB_CAPTION
	AV_CODEC_ID_WEBVTT             AVCodecID = C.AV_CODEC_ID_WEBVTT
)

const (
	AV_INPUT_BUFFER_PADDING_SIZE int = C.AV_INPUT_BUFFER_PADDING_SIZE
)

const (
	AV_CODEC_FLAG_UNALIGNED      AVCodecFlag  = C.AV_CODEC_FLAG_UNALIGNED
	AV_CODEC_FLAG_QSCALE         AVCodecFlag  = C.AV_CODEC_FLAG_QSCALE
	AV_CODEC_FLAG_4MV            AVCodecFlag  = C.AV_CODEC_FLAG_4MV
	AV_CODEC_FLAG_OUTPUT_CORRUPT AVCodecFlag  = C.AV_CODEC_FLAG_OUTPUT_CORRUPT
	AV_CODEC_FLAG_QPEL           AVCodecFlag  = C.AV_CODEC_FLAG_QPEL
	AV_CODEC_FLAG_RECON_FRAME    AVCodecFlag  = C.AV_CODEC_FLAG_RECON_FRAME
	AV_CODEC_FLAG_COPY_OPAQUE    AVCodecFlag  = C.AV_CODEC_FLAG_COPY_OPAQUE
	AV_CODEC_FLAG_FRAME_DURATION AVCodecFlag  = C.AV_CODEC_FLAG_FRAME_DURATION
	AV_CODEC_FLAG_PASS1          AVCodecFlag  = C.AV_CODEC_FLAG_PASS1
	AV_CODEC_FLAG_PASS2          AVCodecFlag  = C.AV_CODEC_FLAG_PASS2
	AV_CODEC_FLAG_LOOP_FILTER    AVCodecFlag  = C.AV_CODEC_FLAG_LOOP_FILTER
	AV_CODEC_FLAG_GRAY           AVCodecFlag  = C.AV_CODEC_FLAG_GRAY
	AV_CODEC_FLAG_PSNR           AVCodecFlag  = C.AV_CODEC_FLAG_PSNR
	AV_CODEC_FLAG_INTERLACED_DCT AVCodecFlag  = C.AV_CODEC_FLAG_INTERLACED_DCT
	AV_CODEC_FLAG_LOW_DELAY      AVCodecFlag  = C.AV_CODEC_FLAG_LOW_DELAY
	AV_CODEC_FLAG_GLOBAL_HEADER  AVCodecFlag  = C.AV_CODEC_FLAG_GLOBAL_HEADER
	AV_CODEC_FLAG_BITEXACT       AVCodecFlag  = C.AV_CODEC_FLAG_BITEXACT
	AV_CODEC_FLAG_AC_PRED        AVCodecFlag  = C.AV_CODEC_FLAG_AC_PRED
	AV_CODEC_FLAG_INTERLACED_ME  AVCodecFlag  = C.AV_CODEC_FLAG_INTERLACED_ME
	AV_CODEC_FLAG_CLOSED_GOP     AVCodecFlag  = C.AV_CODEC_FLAG_CLOSED_GOP
	AV_CODEC_FLAG2_FAST          AVCodecFlag2 = C.AV_CODEC_FLAG2_FAST
	AV_CODEC_FLAG2_NO_OUTPUT     AVCodecFlag2 = C.AV_CODEC_FLAG2_NO_OUTPUT
	AV_CODEC_FLAG2_LOCAL_HEADER  AVCodecFlag2 = C.AV_CODEC_FLAG2_LOCAL_HEADER
	AV_CODEC_FLAG2_CHUNKS        AVCodecFlag2 = C.AV_CODEC_FLAG2_CHUNKS
	AV_CODEC_FLAG2_IGNORE_CROP   AVCodecFlag2 = C.AV_CODEC_FLAG2_IGNORE_CROP
	AV_CODEC_FLAG2_SHOW_ALL      AVCodecFlag2 = C.AV_CODEC_FLAG2_SHOW_ALL
	AV_CODEC_FLAG2_EXPORT_MVS    AVCodecFlag2 = C.AV_CODEC_FLAG2_EXPORT_MVS
	AV_CODEC_FLAG2_SKIP_MANUAL   AVCodecFlag2 = C.AV_CODEC_FLAG2_SKIP_MANUAL
	AV_CODEC_FLAG2_RO_FLUSH_NOOP AVCodecFlag2 = C.AV_CODEC_FLAG2_RO_FLUSH_NOOP
	AV_CODEC_FLAG2_ICC_PROFILES  AVCodecFlag2 = C.AV_CODEC_FLAG2_ICC_PROFILES
)

const (
	AV_CODEC_CAP_NONE                     AVCodecCap = 0
	AV_CODEC_CAP_DRAW_HORIZ_BAND          AVCodecCap = C.AV_CODEC_CAP_DRAW_HORIZ_BAND
	AV_CODEC_CAP_DR1                      AVCodecCap = C.AV_CODEC_CAP_DR1
	AV_CODEC_CAP_DELAY                    AVCodecCap = C.AV_CODEC_CAP_DELAY
	AV_CODEC_CAP_SMALL_LAST_FRAME         AVCodecCap = C.AV_CODEC_CAP_SMALL_LAST_FRAME
	AV_CODEC_CAP_EXPERIMENTAL             AVCodecCap = C.AV_CODEC_CAP_EXPERIMENTAL
	AV_CODEC_CAP_CHANNEL_CONF             AVCodecCap = C.AV_CODEC_CAP_CHANNEL_CONF
	AV_CODEC_CAP_FRAME_THREADS            AVCodecCap = C.AV_CODEC_CAP_FRAME_THREADS
	AV_CODEC_CAP_SLICE_THREADS            AVCodecCap = C.AV_CODEC_CAP_SLICE_THREADS
	AV_CODEC_CAP_PARAM_CHANGE             AVCodecCap = C.AV_CODEC_CAP_PARAM_CHANGE
	AV_CODEC_CAP_OTHER_THREADS            AVCodecCap = C.AV_CODEC_CAP_OTHER_THREADS
	AV_CODEC_CAP_VARIABLE_FRAME_SIZE      AVCodecCap = C.AV_CODEC_CAP_VARIABLE_FRAME_SIZE
	AV_CODEC_CAP_AVOID_PROBING            AVCodecCap = C.AV_CODEC_CAP_AVOID_PROBING
	AV_CODEC_CAP_HARDWARE                 AVCodecCap = C.AV_CODEC_CAP_HARDWARE
	AV_CODEC_CAP_HYBRID                   AVCodecCap = C.AV_CODEC_CAP_HYBRID
	AV_CODEC_CAP_ENCODER_REORDERED_OPAQUE AVCodecCap = C.AV_CODEC_CAP_ENCODER_REORDERED_OPAQUE
	AV_CODEC_CAP_ENCODER_FLUSH            AVCodecCap = C.AV_CODEC_CAP_ENCODER_FLUSH
	AV_CODEC_CAP_ENCODER_RECON_FRAME      AVCodecCap = C.AV_CODEC_CAP_ENCODER_RECON_FRAME
	AV_CODEC_CAP_MAX                                 = AV_CODEC_CAP_ENCODER_RECON_FRAME
)

////////////////////////////////////////////////////////////////////////////////
// JSON OUTPUT

func (ctx *AVCodec) MarshalJSON() ([]byte, error) {
	type jsonAVCodec struct {
		Type           AVMediaType       `json:"type"`
		Name           string            `json:"name,omitempty"`
		LongName       string            `json:"long_name,omitempty"`
		ID             AVCodecID         `json:"id,omitempty"`
		Capabilities   AVCodecCap        `json:"capabilities,omitempty"`
		Framerates     []AVRational      `json:"supported_framerates,omitempty"`
		SampleFormats  []AVSampleFormat  `json:"sample_formats,omitempty"`
		PixelFormats   []AVPixelFormat   `json:"pixel_formats,omitempty"`
		Samplerates    []int             `json:"samplerates,omitempty"`
		Profiles       []AVProfile       `json:"profiles,omitempty"`
		ChannelLayouts []AVChannelLayout `json:"channel_layouts,omitempty"`
	}
	return json.Marshal(jsonAVCodec{
		Name:           C.GoString(ctx.name),
		LongName:       C.GoString(ctx.long_name),
		Type:           AVMediaType(ctx._type),
		ID:             AVCodecID(ctx.id),
		Capabilities:   AVCodecCap(ctx.capabilities),
		Framerates:     ctx.SupportedFramerates(),
		SampleFormats:  ctx.SampleFormats(),
		PixelFormats:   ctx.PixelFormats(),
		Samplerates:    ctx.SupportedSamplerates(),
		Profiles:       ctx.Profiles(),
		ChannelLayouts: ctx.ChannelLayouts(),
	})
}

func (ctx *AVCodecContext) MarshalJSON() ([]byte, error) {
	type jsonAVCodecContext struct {
		CodecType         AVMediaType     `json:"codec_type,omitempty"`
		Codec             *AVCodec        `json:"codec,omitempty"`
		BitRate           int64           `json:"bit_rate,omitempty"`
		BitRateTolerance  int             `json:"bit_rate_tolerance,omitempty"`
		PixelFormat       AVPixelFormat   `json:"pix_fmt,omitempty"`
		Width             int             `json:"width,omitempty"`
		Height            int             `json:"height,omitempty"`
		SampleAspectRatio AVRational      `json:"sample_aspect_ratio,omitzero"`
		Framerate         AVRational      `json:"framerate,omitzero"`
		SampleFormat      AVSampleFormat  `json:"sample_fmt,omitempty"`
		SampleRate        int             `json:"sample_rate,omitempty"`
		ChannelLayout     AVChannelLayout `json:"channel_layout,omitzero"`
		FrameSize         int             `json:"frame_size,omitempty"`
		TimeBase          AVRational      `json:"time_base,omitempty"`
	}
	switch ctx.codec_type {
	case C.AVMEDIA_TYPE_VIDEO:
		return json.Marshal(jsonAVCodecContext{
			CodecType:         AVMediaType(ctx.codec_type),
			Codec:             (*AVCodec)(ctx.codec),
			BitRate:           int64(ctx.bit_rate),
			BitRateTolerance:  int(ctx.bit_rate_tolerance),
			PixelFormat:       AVPixelFormat(ctx.pix_fmt),
			Width:             int(ctx.width),
			Height:            int(ctx.height),
			SampleAspectRatio: AVRational(ctx.sample_aspect_ratio),
			Framerate:         AVRational(ctx.framerate),
			TimeBase:          (AVRational)(ctx.time_base),
		})
	case C.AVMEDIA_TYPE_AUDIO:
		return json.Marshal(jsonAVCodecContext{
			CodecType:        AVMediaType(ctx.codec_type),
			Codec:            (*AVCodec)(ctx.codec),
			BitRate:          int64(ctx.bit_rate),
			BitRateTolerance: int(ctx.bit_rate_tolerance),
			TimeBase:         (AVRational)(ctx.time_base),
			SampleFormat:     AVSampleFormat(ctx.sample_fmt),
			SampleRate:       int(ctx.sample_rate),
			ChannelLayout:    AVChannelLayout(ctx.ch_layout),
			FrameSize:        int(ctx.frame_size),
		})
	default:
		return json.Marshal(jsonAVCodecContext{
			CodecType:        AVMediaType(ctx.codec_type),
			Codec:            (*AVCodec)(ctx.codec),
			BitRate:          int64(ctx.bit_rate),
			BitRateTolerance: int(ctx.bit_rate_tolerance),
			TimeBase:         (AVRational)(ctx.time_base),
		})
	}
}

func (ctx AVProfile) MarshalJSON() ([]byte, error) {
	return json.Marshal(ctx.Name())
}

func (v AVCodecCap) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v AVCodecID) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVCodec) String() string {
	return marshalToString(ctx)
}

func (ctx *AVCodecContext) String() string {
	return marshalToString(ctx)
}

func (ctx AVProfile) String() string {
	return marshalToString(ctx)
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

func (c *AVCodec) Capabilities() AVCodecCap {
	return AVCodecCap(c.capabilities)
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
		if v.profile == C.AV_PROFILE_UNKNOWN {
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

func (ctx *AVCodecContext) Codec() *AVCodec {
	return (*AVCodec)(ctx.codec)
}

func (ctx *AVCodecContext) CodecType() AVMediaType {
	return AVMediaType(ctx.codec_type)
}

func (ctx *AVCodecContext) CodecID() AVCodecID {
	return AVCodecID(ctx.codec_id)
}

func (ctx *AVCodecContext) CodecTag() uint32 {
	return uint32(ctx.codec_tag)
}

func (ctx *AVCodecContext) BitRate() int64 {
	return int64(ctx.bit_rate)
}

func (ctx *AVCodecContext) SetBitRate(bit_rate int64) {
	ctx.bit_rate = C.int64_t(bit_rate)
}

func (ctx *AVCodecContext) Delay() int {
	return int(ctx.delay)
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

func (ctx *AVCodecContext) CodedWidth() int {
	return int(ctx.coded_width)
}

func (ctx *AVCodecContext) CodedHeight() int {
	return int(ctx.coded_height)
}

func (ctx *AVCodecContext) SampleAspectRatio() AVRational {
	return (AVRational)(ctx.sample_aspect_ratio)
}

func (ctx *AVCodecContext) SetSampleAspectRatio(sample_aspect_ratio AVRational) {
	ctx.sample_aspect_ratio = C.struct_AVRational(sample_aspect_ratio)
}

func (ctx *AVCodecContext) Framerate() AVRational {
	return (AVRational)(ctx.framerate)
}

func (ctx *AVCodecContext) SetFramerate(framerate AVRational) {
	ctx.framerate = C.struct_AVRational(framerate)
}

func (ctx *AVCodecContext) TimeBase() AVRational {
	return (AVRational)(ctx.time_base)
}

func (ctx *AVCodecContext) SetTimeBase(time_base AVRational) {
	ctx.time_base = C.struct_AVRational(time_base)
}

func (ctx *AVCodecContext) SampleFormat() AVSampleFormat {
	return AVSampleFormat(ctx.sample_fmt)
}

func (ctx *AVCodecContext) SetSampleFormat(sample_fmt AVSampleFormat) {
	ctx.sample_fmt = C.enum_AVSampleFormat(sample_fmt)
}

func (ctx *AVCodecContext) FrameNum() int64 {
	return int64(ctx.frame_num)
}

func (ctx *AVCodecContext) SampleRate() int {
	return int(ctx.sample_rate)
}

func (ctx *AVCodecContext) SetSampleRate(sample_rate int) {
	ctx.sample_rate = C.int(sample_rate)
}

func (ctx *AVCodecContext) FrameSize() int {
	return int(ctx.frame_size)
}

func (ctx *AVCodecContext) ChannelLayout() AVChannelLayout {
	return AVChannelLayout(ctx.ch_layout)
}

func (ctx *AVCodecContext) SetChannelLayout(src AVChannelLayout) error {
	// Copy the new layout first to avoid leaving struct in invalid state on error
	var temp C.struct_AVChannelLayout
	if ret := AVError(C.av_channel_layout_copy(&temp, (*C.struct_AVChannelLayout)(&src))); ret != 0 {
		return ret
	}
	// Now free the existing layout and replace it
	C.av_channel_layout_uninit((*C.struct_AVChannelLayout)(&ctx.ch_layout))
	ctx.ch_layout = temp
	return nil
}

func (ctx *AVCodecContext) GopSize() int {
	return int(ctx.gop_size)
}

func (ctx *AVCodecContext) SetGopSize(gop_size int) {
	ctx.gop_size = C.int(gop_size)
}

func (ctx *AVCodecContext) MaxBFrames() int {
	return int(ctx.max_b_frames)
}

func (ctx *AVCodecContext) SetMaxBFrames(max_b_frames int) {
	ctx.max_b_frames = C.int(max_b_frames)
}

func (ctx *AVCodecContext) PixFmt() AVPixelFormat {
	return AVPixelFormat(ctx.pix_fmt)
}

func (ctx *AVCodecContext) SetPixFmt(pix_fmt AVPixelFormat) {
	ctx.pix_fmt = C.enum_AVPixelFormat(pix_fmt)
}

func (ctx *AVCodecContext) SetPrivDataKV(name, value string) error {
	cName, cValue := C.CString(name), C.CString(value)
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cValue))
	if ret := AVError(C.av_opt_set(ctx.priv_data, cName, cValue, 0)); ret != 0 {
		return ret
	}
	return nil
}

func (ctx *AVCodecContext) Flags() AVCodecFlag {
	return AVCodecFlag(ctx.flags)
}

func (ctx *AVCodecContext) SetFlags(flags AVCodecFlag) {
	ctx.flags = C.int(flags)
}

func (ctx *AVCodecContext) Flags2() AVCodecFlag2 {
	return AVCodecFlag2(ctx.flags2)
}

func (ctx *AVCodecContext) SetFlags2(flags2 AVCodecFlag2) {
	ctx.flags2 = C.int(flags2)
}

////////////////////////////////////////////////////////////////////////////////
// AVProfile

func (c *AVProfile) ID() int {
	return int(c.profile)
}

func (c *AVProfile) Name() string {
	return C.GoString(c.name)
}

////////////////////////////////////////////////////////////////////////////////
// AVCodecCap

func (v AVCodecCap) Is(cap AVCodecCap) bool {
	return v&cap == cap
}

func (v AVCodecCap) String() string {
	if v == AV_CODEC_CAP_NONE {
		return v.FlagString()
	}
	str := ""
	for i := AVCodecCap(C.int(1)); i <= AV_CODEC_CAP_MAX; i <<= 1 {
		if v&i == i {
			str += "|" + i.FlagString()
		}
	}
	return str[1:]
}

func (v AVCodecCap) FlagString() string {
	switch v {
	case AV_CODEC_CAP_NONE:
		return "AV_CODEC_CAP_NONE"
	case AV_CODEC_CAP_DRAW_HORIZ_BAND:
		return "AV_CODEC_CAP_DRAW_HORIZ_BAND"
	case AV_CODEC_CAP_DR1:
		return "AV_CODEC_CAP_DR1"
	case AV_CODEC_CAP_DELAY:
		return "AV_CODEC_CAP_DELAY"
	case AV_CODEC_CAP_SMALL_LAST_FRAME:
		return "AV_CODEC_CAP_SMALL_LAST_FRAME"
	case AV_CODEC_CAP_EXPERIMENTAL:
		return "AV_CODEC_CAP_EXPERIMENTAL"
	case AV_CODEC_CAP_CHANNEL_CONF:
		return "AV_CODEC_CAP_CHANNEL_CONF"
	case AV_CODEC_CAP_FRAME_THREADS:
		return "AV_CODEC_CAP_FRAME_THREADS"
	case AV_CODEC_CAP_SLICE_THREADS:
		return "AV_CODEC_CAP_SLICE_THREADS"
	case AV_CODEC_CAP_PARAM_CHANGE:
		return "AV_CODEC_CAP_PARAM_CHANGE"
	case AV_CODEC_CAP_OTHER_THREADS:
		return "AV_CODEC_CAP_OTHER_THREADS"
	case AV_CODEC_CAP_VARIABLE_FRAME_SIZE:
		return "AV_CODEC_CAP_VARIABLE_FRAME_SIZE"
	case AV_CODEC_CAP_AVOID_PROBING:
		return "AV_CODEC_CAP_AVOID_PROBING"
	case AV_CODEC_CAP_HARDWARE:
		return "AV_CODEC_CAP_HARDWARE"
	case AV_CODEC_CAP_HYBRID:
		return "AV_CODEC_CAP_HYBRID"
	case AV_CODEC_CAP_ENCODER_REORDERED_OPAQUE:
		return "AV_CODEC_CAP_ENCODER_REORDERED_OPAQUE"
	case AV_CODEC_CAP_ENCODER_FLUSH:
		return "AV_CODEC_CAP_ENCODER_FLUSH"
	case AV_CODEC_CAP_ENCODER_RECON_FRAME:
		return "AV_CODEC_CAP_ENCODER_RECON_FRAME"
	default:
		return fmt.Sprintf("AVCodecCap(0x%08X)", uint32(v))
	}
}

////////////////////////////////////////////////////////////////////////////////
// AVCodecID

func (v AVCodecID) String() string {
	return v.Name()
}

func (v AVCodecID) Name() string {
	return C.GoString(C.avcodec_get_name(C.enum_AVCodecID(v)))
}

func (v AVCodecID) Type() AVMediaType {
	return AVMediaType(C.avcodec_get_type(C.enum_AVCodecID(v)))
}
