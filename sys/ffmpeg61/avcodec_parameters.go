package ffmpeg

import (
	"encoding/json"
	"errors"
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
	Framerate         AVRational    `json:"framerate,omitempty"`
}

type jsonAVCodecParameters struct {
	CodecType AVMediaType `json:"codec_type"`
	CodecID   AVCodecID   `json:"codec_id,omitempty"`
	CodecTag  uint32      `json:"codec_tag,omitempty"`
	BitRate   int64       `json:"bit_rate,omitempty"`
	*jsonAVCodecParametersAudio
	*jsonAVCodecParameterVideo
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx AVCodecParameters) MarshalJSON() ([]byte, error) {
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
			Framerate:         AVRational(ctx.framerate),
		}
	}

	return json.Marshal(par)
}

func (ctx *AVCodecParameters) String() string {
	data, _ := json.MarshalIndent(ctx, "", "  ")
	return string(data)
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

func (ctx *AVCodecParameters) CodecTag() uint32 {
	return uint32(ctx.codec_tag)
}

func (ctx *AVCodecParameters) SetCodecTag(tag uint32) {
	ctx.codec_tag = C.uint32_t(tag)
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
func (ctx *AVCodecParameters) Samplerate() int {
	return int(ctx.sample_rate)
}

func (ctx *AVCodecParameters) SetSamplerate(rate int) {
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
	ctx.ch_layout = C.AVChannelLayout(layout)
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

func (ctx *AVCodecParameters) Framerate() AVRational {
	return AVRational(ctx.framerate)
}

func (ctx *AVCodecParameters) SetFramerate(rate AVRational) {
	ctx.framerate = C.AVRational(rate)
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
