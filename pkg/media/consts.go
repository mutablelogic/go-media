package media

import (
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// INIT

func init() {
	for v := SampleFormat(0); v <= SAMPLE_FORMAT_MAX; v++ {
		mapSampleFormat[toSampleFormat(v)] = v
	}
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	mapSampleFormat = make(map[ffmpeg.AVSampleFormat]SampleFormat)
)

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func fromSampleFormat(v ffmpeg.AVSampleFormat) SampleFormat {
	if v, ok := mapSampleFormat[v]; ok {
		return v
	} else {
		return SAMPLE_FORMAT_NONE
	}
}

func fromChannelLayout(v ffmpeg.AVChannelLayout) ChannelLayout {
	for layout := ChannelLayout(1); layout <= CHANNEL_LAYOUT_MAX; layout++ {
		if toChannelLayout(layout) == v {
			return layout
		}
	}
	return CHANNEL_LAYOUT_NONE
}

func fromPixelFormat(v ffmpeg.AVPixelFormat) PixelFormat {
	// TODO: Implement
	return PixelFormat(v)
}

func toChannelLayout(v ChannelLayout) ffmpeg.AVChannelLayout {
	switch v {
	case CHANNEL_LAYOUT_MONO:
		return ffmpeg.AV_CHANNEL_LAYOUT_MONO
	case CHANNEL_LAYOUT_STEREO:
		return ffmpeg.AV_CHANNEL_LAYOUT_STEREO
	case CHANNEL_LAYOUT_2POINT1:
		return ffmpeg.AV_CHANNEL_LAYOUT_2POINT1
	case CHANNEL_LAYOUT_2_1:
		return ffmpeg.AV_CHANNEL_LAYOUT_2_1
	case CHANNEL_LAYOUT_SURROUND:
		return ffmpeg.AV_CHANNEL_LAYOUT_SURROUND
	case CHANNEL_LAYOUT_3POINT1:
		return ffmpeg.AV_CHANNEL_LAYOUT_3POINT1
	case CHANNEL_LAYOUT_4POINT0:
		return ffmpeg.AV_CHANNEL_LAYOUT_4POINT0
	case CHANNEL_LAYOUT_4POINT1:
		return ffmpeg.AV_CHANNEL_LAYOUT_4POINT1
	case CHANNEL_LAYOUT_2_2:
		return ffmpeg.AV_CHANNEL_LAYOUT_2_2
	case CHANNEL_LAYOUT_QUAD:
		return ffmpeg.AV_CHANNEL_LAYOUT_QUAD
	case CHANNEL_LAYOUT_5POINT0:
		return ffmpeg.AV_CHANNEL_LAYOUT_5POINT0
	case CHANNEL_LAYOUT_5POINT1:
		return ffmpeg.AV_CHANNEL_LAYOUT_5POINT1
	case CHANNEL_LAYOUT_5POINT0_BACK:
		return ffmpeg.AV_CHANNEL_LAYOUT_5POINT0_BACK
	case CHANNEL_LAYOUT_5POINT1_BACK:
		return ffmpeg.AV_CHANNEL_LAYOUT_5POINT1_BACK
	case CHANNEL_LAYOUT_6POINT0:
		return ffmpeg.AV_CHANNEL_LAYOUT_6POINT0
	case CHANNEL_LAYOUT_6POINT0_FRONT:
		return ffmpeg.AV_CHANNEL_LAYOUT_6POINT0_FRONT
	case CHANNEL_LAYOUT_HEXAGONAL:
		return ffmpeg.AV_CHANNEL_LAYOUT_HEXAGONAL
	case CHANNEL_LAYOUT_6POINT1:
		return ffmpeg.AV_CHANNEL_LAYOUT_6POINT1
	case CHANNEL_LAYOUT_6POINT1_BACK:
		return ffmpeg.AV_CHANNEL_LAYOUT_6POINT1_BACK
	case CHANNEL_LAYOUT_6POINT1_FRONT:
		return ffmpeg.AV_CHANNEL_LAYOUT_6POINT1_FRONT
	case CHANNEL_LAYOUT_7POINT0:
		return ffmpeg.AV_CHANNEL_LAYOUT_7POINT0
	case CHANNEL_LAYOUT_7POINT0_FRONT:
		return ffmpeg.AV_CHANNEL_LAYOUT_7POINT0_FRONT
	case CHANNEL_LAYOUT_7POINT1:
		return ffmpeg.AV_CHANNEL_LAYOUT_7POINT1
	case CHANNEL_LAYOUT_7POINT1_WIDE:
		return ffmpeg.AV_CHANNEL_LAYOUT_7POINT1_WIDE
	case CHANNEL_LAYOUT_7POINT1_WIDE_BACK:
		return ffmpeg.AV_CHANNEL_LAYOUT_7POINT1_WIDE_BACK
	case CHANNEL_LAYOUT_OCTAGONAL:
		return ffmpeg.AV_CHANNEL_LAYOUT_OCTAGONAL
	case CHANNEL_LAYOUT_STEREO_DOWNMIX:
		return ffmpeg.AV_CHANNEL_LAYOUT_STEREO_DOWNMIX
	case CHANNEL_LAYOUT_22POINT2:
		return ffmpeg.AV_CHANNEL_LAYOUT_22POINT2
	case CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER:
		return ffmpeg.AV_CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER
	default:
		return ffmpeg.AV_CHANNEL_LAYOUT_MONO
	}
}

func toSampleFormat(v SampleFormat) ffmpeg.AVSampleFormat {
	switch v {
	case SAMPLE_FORMAT_NONE:
		return ffmpeg.AV_SAMPLE_FMT_NONE
	case SAMPLE_FORMAT_U8:
		return ffmpeg.AV_SAMPLE_FMT_U8
	case SAMPLE_FORMAT_S16:
		return ffmpeg.AV_SAMPLE_FMT_S16
	case SAMPLE_FORMAT_S32:
		return ffmpeg.AV_SAMPLE_FMT_S32
	case SAMPLE_FORMAT_FLT:
		return ffmpeg.AV_SAMPLE_FMT_FLT
	case SAMPLE_FORMAT_DBL:
		return ffmpeg.AV_SAMPLE_FMT_DBL
	case SAMPLE_FORMAT_U8P:
		return ffmpeg.AV_SAMPLE_FMT_U8P
	case SAMPLE_FORMAT_S16P:
		return ffmpeg.AV_SAMPLE_FMT_S16P
	case SAMPLE_FORMAT_S32P:
		return ffmpeg.AV_SAMPLE_FMT_S32P
	case SAMPLE_FORMAT_FLTP:
		return ffmpeg.AV_SAMPLE_FMT_FLTP
	case SAMPLE_FORMAT_DBLP:
		return ffmpeg.AV_SAMPLE_FMT_DBLP
	case SAMPLE_FORMAT_S64:
		return ffmpeg.AV_SAMPLE_FMT_S64
	case SAMPLE_FORMAT_S64P:
		return ffmpeg.AV_SAMPLE_FMT_S64P
	default:
		return ffmpeg.AV_SAMPLE_FMT_NB
	}
}
