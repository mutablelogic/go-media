package ffmpeg

import (
	"encoding/json"
	"errors"
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
// MEMORY MANAGEMENT

// Allocate a new AVCodecParameters and set its fields to default values.
func AVCodec_parameters_alloc() *AVCodecParameters {
	return (*AVCodecParameters)(C.avcodec_parameters_alloc())
}

// Free an AVCodecParameters instance and everything associated with it.
func AVCodec_parameters_free(par *AVCodecParameters) {
	C.avcodec_parameters_free((**C.struct_AVCodecParameters)(unsafe.Pointer(&par)))
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx AVCodecParameters) MarshalJSON() ([]byte, error) {

	type jsonAVCodecParametersAudio struct {
		SampleFormat  AVSampleFormat  `json:"sample_format"`
		SampleRate    int             `json:"sample_rate"`
		ChannelLayout AVChannelLayout `json:"channel_layout"`
		FrameSize     int             `json:"frame_size,omitempty"`
	}

	type jsonAVCodecParameterVideo struct {
		PixelFormat       AVPixelFormat `json:"pixel_format"`
		Width             int           `json:"width"`
		Height            int           `json:"height"`
		SampleAspectRatio AVRational    `json:"sample_aspect_ratio,omitempty"`
	}

	type jsonAVCodecParameters struct {
		CodecType AVMediaType `json:"codec_type"`
		CodecID   AVCodecID   `json:"codec_id,omitempty"`
		CodecTag  uint32      `json:"codec_tag,omitempty"`
		BitRate   int64       `json:"bit_rate,omitempty"`
		*jsonAVCodecParametersAudio
		*jsonAVCodecParameterVideo
	}

	par := jsonAVCodecParameters{
		CodecType: AVMediaType(ctx.codec_type),
		CodecID:   AVCodecID(ctx.codec_id),
		CodecTag:  uint32(ctx.codec_tag),
		BitRate:   int64(ctx.bit_rate),
	}
	switch ctx.CodecType() {
	case AVMEDIA_TYPE_AUDIO:
		par.jsonAVCodecParametersAudio = &jsonAVCodecParametersAudio{
			SampleFormat:  AVSampleFormat(ctx.format),
			SampleRate:    int(ctx.sample_rate),
			ChannelLayout: AVChannelLayout(ctx.ch_layout),
			FrameSize:     int(ctx.frame_size),
		}
	case AVMEDIA_TYPE_VIDEO:
		par.jsonAVCodecParameterVideo = &jsonAVCodecParameterVideo{
			PixelFormat:       AVPixelFormat(ctx.format),
			Width:             int(ctx.width),
			Height:            int(ctx.height),
			SampleAspectRatio: AVRational(ctx.sample_aspect_ratio),
		}
	}

	return json.Marshal(par)
}

// UnmarshalJSON implements the json.Unmarshaler interface, reconstructing
// ctx from the shape produced by MarshalJSON. String identifiers (codec
// type/ID, sample/pixel format, channel layout) are resolved back to their
// ffmpeg enum values via the corresponding lookup functions/UnmarshalJSON
// implementations - encoding/json auto-allocates whichever of the embedded
// audio/video pointer structs has matching keys present, mirroring
// MarshalJSON's own type-switch.
func (ctx *AVCodecParameters) UnmarshalJSON(data []byte) error {
	// encoding/json refuses to auto-allocate an embedded pointer field
	// whose element type is unexported (even a local, function-scoped
	// type) when unmarshaling - "cannot set embedded pointer to
	// unexported struct". MarshalJSON's lowercase locals never hit this
	// since marshaling only reads, never allocates, so these need
	// capitalized names purely to satisfy that check.
	type JSONAVCodecParametersAudio struct {
		SampleFormat  AVSampleFormat  `json:"sample_format"`
		SampleRate    int             `json:"sample_rate"`
		ChannelLayout AVChannelLayout `json:"channel_layout"`
		FrameSize     int             `json:"frame_size,omitempty"`
	}

	type JSONAVCodecParameterVideo struct {
		PixelFormat       AVPixelFormat `json:"pixel_format"`
		Width             int           `json:"width"`
		Height            int           `json:"height"`
		SampleAspectRatio AVRational    `json:"sample_aspect_ratio,omitempty"`
	}

	type jsonAVCodecParameters struct {
		CodecType AVMediaType `json:"codec_type"`
		CodecID   AVCodecID   `json:"codec_id,omitempty"`
		CodecTag  uint32      `json:"codec_tag,omitempty"`
		BitRate   int64       `json:"bit_rate,omitempty"`
		*JSONAVCodecParametersAudio
		*JSONAVCodecParameterVideo
	}

	var par jsonAVCodecParameters
	if err := json.Unmarshal(data, &par); err != nil {
		return err
	}

	ctx.SetCodecType(par.CodecType)
	ctx.SetCodecID(par.CodecID)
	ctx.SetCodecTag(par.CodecTag)
	ctx.SetBitRate(par.BitRate)

	if par.JSONAVCodecParametersAudio != nil {
		ctx.SetSampleFormat(par.SampleFormat)
		ctx.SetSampleRate(par.SampleRate)
		if par.ChannelLayout.NumChannels() > 0 {
			if err := ctx.SetChannelLayout(par.ChannelLayout); err != nil {
				return err
			}
		}
		ctx.SetFrameSize(par.FrameSize)
	}
	if par.JSONAVCodecParameterVideo != nil {
		ctx.SetPixelFormat(par.PixelFormat)
		ctx.SetWidth(par.Width)
		ctx.SetHeight(par.Height)
		ctx.SetSampleAspectRatio(par.SampleAspectRatio)
	}

	return nil
}

func (ctx *AVCodecParameters) String() string {
	return marshalToString(ctx)
}

////////////////////////////////////////////////////////////////////////////////
// PARAMETERS

func (ctx *AVCodecParameters) CodecType() AVMediaType {
	return AVMediaType(ctx.codec_type)
}

func (ctx *AVCodecParameters) SetCodecType(t AVMediaType) {
	ctx.codec_type = C.enum_AVMediaType(t)
}

func (ctx *AVCodecParameters) CodecID() AVCodecID {
	return AVCodecID(ctx.codec_id)
}

func (ctx *AVCodecParameters) SetCodecID(id AVCodecID) {
	ctx.codec_id = C.enum_AVCodecID(id)
}

func (ctx *AVCodecParameters) CodecTag() uint32 {
	return uint32(ctx.codec_tag)
}

func (ctx *AVCodecParameters) SetCodecTag(tag uint32) {
	ctx.codec_tag = C.uint32_t(tag)
}

// AV_PROFILE_UNKNOWN indicates no specific codec profile is requested; the
// encoder picks its own default. This is the zero value ffmpeg itself uses,
// not Go's zero value (0), which collides with real profile IDs (e.g. AAC's
// FF_PROFILE_AAC_MAIN) — a freshly zero-valued AVCodecParameters must have
// this set explicitly before being copied onto an AVCodecContext, or the
// encoder will reject it as an unsupported profile.
const AV_PROFILE_UNKNOWN = -99

func (ctx *AVCodecParameters) Profile() int {
	return int(ctx.profile)
}

func (ctx *AVCodecParameters) SetProfile(profile int) {
	ctx.profile = C.int(profile)
}

// Audio and Video
func (ctx *AVCodecParameters) Format() int {
	return int(ctx.format)
}

// Audio and Video
func (ctx *AVCodecParameters) BitRate() int64 {
	return int64(ctx.bit_rate)
}

func (ctx *AVCodecParameters) SetBitRate(rate int64) {
	ctx.bit_rate = C.int64_t(rate)
}

// Audio
func (ctx *AVCodecParameters) SampleFormat() AVSampleFormat {
	if AVMediaType(ctx.codec_type) == AVMEDIA_TYPE_AUDIO {
		return AVSampleFormat(ctx.format)
	} else {
		return AV_SAMPLE_FMT_NONE
	}
}

func (ctx *AVCodecParameters) SetSampleFormat(format AVSampleFormat) {
	ctx.format = C.int(format)
}

// Audio
func (ctx *AVCodecParameters) SampleRate() int {
	return int(ctx.sample_rate)
}

func (ctx *AVCodecParameters) SetSampleRate(rate int) {
	ctx.sample_rate = C.int(rate)
}

// Audio
func (ctx *AVCodecParameters) ChannelLayout() AVChannelLayout {
	return AVChannelLayout(ctx.ch_layout)
}

func (ctx *AVCodecParameters) SetChannelLayout(layout AVChannelLayout) error {
	if !AVUtil_channel_layout_check(&layout) {
		return errors.New("invalid channel layout")
	}
	// Copy the new layout first to avoid leaving struct in invalid state on error
	var temp C.struct_AVChannelLayout
	if ret := AVError(C.av_channel_layout_copy(&temp, (*C.struct_AVChannelLayout)(&layout))); ret != 0 {
		return ret
	}
	// Now free the existing layout and replace it
	C.av_channel_layout_uninit((*C.struct_AVChannelLayout)(&ctx.ch_layout))
	ctx.ch_layout = temp
	return nil
}

// Audio
func (ctx *AVCodecParameters) FrameSize() int {
	return int(ctx.frame_size)
}

func (ctx *AVCodecParameters) SetFrameSize(size int) {
	ctx.frame_size = C.int(size)
}

// Video
func (ctx *AVCodecParameters) PixelFormat() AVPixelFormat {
	if AVMediaType(ctx.codec_type) == AVMEDIA_TYPE_VIDEO {
		return AVPixelFormat(ctx.format)
	} else {
		return AV_PIX_FMT_NONE
	}
}

func (ctx *AVCodecParameters) SetPixelFormat(format AVPixelFormat) {
	ctx.format = C.int(format)
}

// Video
func (ctx *AVCodecParameters) SampleAspectRatio() AVRational {
	return AVRational(ctx.sample_aspect_ratio)
}

func (ctx *AVCodecParameters) SetSampleAspectRatio(aspect AVRational) {
	ctx.sample_aspect_ratio = C.AVRational(aspect)
}

// Video
func (ctx *AVCodecParameters) Width() int {
	return int(ctx.width)
}

func (ctx *AVCodecParameters) SetWidth(width int) {
	ctx.width = C.int(width)
}

// Video
func (ctx *AVCodecParameters) Height() int {
	return int(ctx.height)
}

func (ctx *AVCodecParameters) SetHeight(height int) {
	ctx.height = C.int(height)
}
