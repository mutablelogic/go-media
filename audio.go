package media

import (
	"fmt"
	"io"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// SampleFormat specifies the type of a single sample
type SampleFormat uint

// ChannelLayout specifies the layout of channels
type ChannelLayout uint

// AudioChannel specifies a single audio channel
type AudioChannel uint

// AudioFormat specifies the interface for audio format
type AudioFormat struct {
	// Sample rate in Hz
	Rate uint

	// Sample format
	Format SampleFormat

	// Channel layout
	Layout ChannelLayout
}

// SWResampleFn is a function that accepts an "output" audio frame,
// which can be nil if the conversion has not started yet, and should
// fill the audio frame provided to the Convert function. Should return
// io.EOF on end of conversion, or any other error to stop the conversion.
type SWResampleFn func(AudioFrame) error

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

/*
// AudioFrame is a slice of audio samples
type AudioFrame interface {
	io.Closer

	// Audio format
	AudioFormat() AudioFormat

	// Number of samples in a single channel
	Samples() int

	// Audio channels
	Channels() []AudioChannel

	// Duration of the frame
	Duration() time.Duration

	// Returns true if planar format (one set of samples per channel)
	IsPlanar() bool

	// Returns the samples for a specified channel, as array of bytes. For packed
	// audio format, the channel should be 0.
	Bytes(channel int) []byte
}
*/

// SWResample is an interface to the ffmpeg swresample library
// which resamples audio.
type SWResample interface {
	io.Closer

	// Create a new empty context object for conversion, with an input frame which
	// will be used to store the data and the target output format.
	Convert(AudioFrame, AudioFormat, SWResampleFn) error
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	SAMPLE_FORMAT_NONE SampleFormat = iota
	SAMPLE_FORMAT_U8                // Byte
	SAMPLE_FORMAT_S16               // Signed 16-bit
	SAMPLE_FORMAT_S32               // Signed 32-bit
	SAMPLE_FORMAT_S64               // Signed 64-bit
	SAMPLE_FORMAT_FLT               // Float 32-bit
	SAMPLE_FORMAT_DBL               // Float 64-bit
	SAMPLE_FORMAT_U8P               // Planar byte
	SAMPLE_FORMAT_S16P              // Planar signed 16-bit
	SAMPLE_FORMAT_S32P              // Planar signed 32-bit
	SAMPLE_FORMAT_S64P              // Planar signed 64-bit
	SAMPLE_FORMAT_FLTP              // Planar float 32-bit
	SAMPLE_FORMAT_DBLP              // Planar float 64-bit
	SAMPLE_FORMAT_MAX  = SAMPLE_FORMAT_DBLP
)

const (
	CHANNEL_LAYOUT_NONE ChannelLayout = iota
	CHANNEL_LAYOUT_MONO
	CHANNEL_LAYOUT_STEREO
	CHANNEL_LAYOUT_2POINT1
	CHANNEL_LAYOUT_2_1
	CHANNEL_LAYOUT_SURROUND
	CHANNEL_LAYOUT_3POINT1
	CHANNEL_LAYOUT_4POINT0
	CHANNEL_LAYOUT_4POINT1
	CHANNEL_LAYOUT_2_2
	CHANNEL_LAYOUT_QUAD
	CHANNEL_LAYOUT_5POINT0
	CHANNEL_LAYOUT_5POINT1
	CHANNEL_LAYOUT_5POINT0_BACK
	CHANNEL_LAYOUT_5POINT1_BACK
	CHANNEL_LAYOUT_6POINT0
	CHANNEL_LAYOUT_6POINT0_FRONT
	CHANNEL_LAYOUT_HEXAGONAL
	CHANNEL_LAYOUT_6POINT1
	CHANNEL_LAYOUT_6POINT1_BACK
	CHANNEL_LAYOUT_6POINT1_FRONT
	CHANNEL_LAYOUT_7POINT0
	CHANNEL_LAYOUT_7POINT0_FRONT
	CHANNEL_LAYOUT_7POINT1
	CHANNEL_LAYOUT_7POINT1_WIDE
	CHANNEL_LAYOUT_7POINT1_WIDE_BACK
	CHANNEL_LAYOUT_OCTAGONAL
	CHANNEL_LAYOUT_HEXADECAGONAL
	CHANNEL_LAYOUT_STEREO_DOWNMIX
	CHANNEL_LAYOUT_22POINT2
	CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER
	CHANNEL_LAYOUT_MAX = CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER
)

const (
	CHANNEL_NONE AudioChannel = iota
	CHANNEL_FRONT_LEFT
	CHANNEL_FRONT_RIGHT
	CHANNEL_FRONT_CENTER
	CHANNEL_LOW_FREQUENCY
	CHANNEL_BACK_LEFT
	CHANNEL_BACK_RIGHT
	CHANNEL_FRONT_LEFT_OF_CENTER
	CHANNEL_FRONT_RIGHT_OF_CENTER
	CHANNEL_BACK_CENTER
	CHANNEL_SIDE_LEFT
	CHANNEL_SIDE_RIGHT
	CHANNEL_TOP_CENTER
	CHANNEL_TOP_FRONT_LEFT
	CHANNEL_TOP_FRONT_CENTER
	CHANNEL_TOP_FRONT_RIGHT
	CHANNEL_TOP_BACK_LEFT
	CHANNEL_TOP_BACK_CENTER
	CHANNEL_TOP_BACK_RIGHT
	CHANNEL_STEREO_LEFT
	CHANNEL_STEREO_RIGHT
	CHANNEL_WIDE_LEFT
	CHANNEL_WIDE_RIGHT
	CHANNEL_SURROUND_DIRECT_LEFT
	CHANNEL_SURROUND_DIRECT_RIGHT
	CHANNEL_LOW_FREQUENCY_2
	CHANNEL_TOP_SIDE_LEFT
	CHANNEL_TOP_SIDE_RIGHT
	CHANNEL_BOTTOM_FRONT_CENTER
	CHANNEL_BOTTOM_FRONT_LEFT
	CHANNEL_BOTTOM_FRONT_RIGHT
	CHANNEL_MAX = CHANNEL_BOTTOM_FRONT_RIGHT
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v AudioFormat) String() string {
	str := "<AudioFormat"
	if v.Rate != 0 {
		str += fmt.Sprint(" rate=", v.Rate)
	}
	if v.Format != SAMPLE_FORMAT_NONE {
		str += fmt.Sprint(" format=", v.Format)
	}
	if v.Layout != CHANNEL_LAYOUT_NONE {
		str += fmt.Sprint(" layout=", v.Layout)
	}
	return str + ">"
}

func (v AudioChannel) String() string {
	switch v {
	case CHANNEL_NONE:
		return "CHANNEL_NONE"
	case CHANNEL_FRONT_LEFT:
		return "CHANNEL_FRONT_LEFT"
	case CHANNEL_FRONT_RIGHT:
		return "CHANNEL_FRONT_RIGHT"
	case CHANNEL_FRONT_CENTER:
		return "CHANNEL_FRONT_CENTER"
	case CHANNEL_LOW_FREQUENCY:
		return "CHANNEL_LOW_FREQUENCY"
	case CHANNEL_BACK_LEFT:
		return "CHANNEL_BACK_LEFT"
	case CHANNEL_BACK_RIGHT:
		return "CHANNEL_BACK_RIGHT"
	case CHANNEL_FRONT_LEFT_OF_CENTER:
		return "CHANNEL_FRONT_LEFT_OF_CENTER"
	case CHANNEL_FRONT_RIGHT_OF_CENTER:
		return "CHANNEL_FRONT_RIGHT_OF_CENTER"
	case CHANNEL_BACK_CENTER:
		return "CHANNEL_BACK_CENTER"
	case CHANNEL_SIDE_LEFT:
		return "CHANNEL_SIDE_LEFT"
	case CHANNEL_SIDE_RIGHT:
		return "CHANNEL_SIDE_RIGHT"
	case CHANNEL_TOP_CENTER:
		return "CHANNEL_TOP_CENTER"
	case CHANNEL_TOP_FRONT_LEFT:
		return "CHANNEL_TOP_FRONT_LEFT"
	case CHANNEL_TOP_FRONT_CENTER:
		return "CHANNEL_TOP_FRONT_CENTER"
	case CHANNEL_TOP_FRONT_RIGHT:
		return "CHANNEL_TOP_FRONT_RIGHT"
	case CHANNEL_TOP_BACK_LEFT:
		return "CHANNEL_TOP_BACK_LEFT"
	case CHANNEL_TOP_BACK_CENTER:
		return "CHANNEL_TOP_BACK_CENTER"
	case CHANNEL_TOP_BACK_RIGHT:
		return "CHANNEL_TOP_BACK_RIGHT"
	case CHANNEL_STEREO_LEFT:
		return "CHANNEL_STEREO_LEFT"
	case CHANNEL_STEREO_RIGHT:
		return "CHANNEL_STEREO_RIGHT"
	case CHANNEL_WIDE_LEFT:
		return "CHANNEL_WIDE_LEFT"
	case CHANNEL_WIDE_RIGHT:
		return "CHANNEL_WIDE_RIGHT"
	case CHANNEL_SURROUND_DIRECT_LEFT:
		return "CHANNEL_SURROUND_DIRECT_LEFT"
	case CHANNEL_SURROUND_DIRECT_RIGHT:
		return "CHANNEL_SURROUND_DIRECT_RIGHT"
	case CHANNEL_LOW_FREQUENCY_2:
		return "CHANNEL_LOW_FREQUENCY_2"
	case CHANNEL_TOP_SIDE_LEFT:
		return "CHANNEL_TOP_SIDE_LEFT"
	case CHANNEL_TOP_SIDE_RIGHT:
		return "CHANNEL_TOP_SIDE_RIGHT"
	case CHANNEL_BOTTOM_FRONT_CENTER:
		return "CHANNEL_BOTTOM_FRONT_CENTER"
	case CHANNEL_BOTTOM_FRONT_LEFT:
		return "CHANNEL_BOTTOM_FRONT_LEFT"
	case CHANNEL_BOTTOM_FRONT_RIGHT:
		return "CHANNEL_BOTTOM_FRONT_RIGHT"
	default:
		return "[?? Invalid AudioChannel value]"
	}
}

func (v SampleFormat) String() string {
	switch v {
	case SAMPLE_FORMAT_NONE:
		return "SAMPLE_FORMAT_NONE"
	case SAMPLE_FORMAT_U8:
		return "SAMPLE_FORMAT_U8"
	case SAMPLE_FORMAT_S16:
		return "SAMPLE_FORMAT_S16"
	case SAMPLE_FORMAT_S32:
		return "SAMPLE_FORMAT_S32"
	case SAMPLE_FORMAT_FLT:
		return "SAMPLE_FORMAT_FLT"
	case SAMPLE_FORMAT_DBL:
		return "SAMPLE_FORMAT_DBL"
	case SAMPLE_FORMAT_U8P:
		return "SAMPLE_FORMAT_U8P"
	case SAMPLE_FORMAT_S16P:
		return "SAMPLE_FORMAT_S16P"
	case SAMPLE_FORMAT_S32P:
		return "SAMPLE_FORMAT_S32P"
	case SAMPLE_FORMAT_FLTP:
		return "SAMPLE_FORMAT_FLTP"
	case SAMPLE_FORMAT_DBLP:
		return "SAMPLE_FORMAT_DBLP"
	case SAMPLE_FORMAT_S64:
		return "SAMPLE_FORMAT_S64"
	case SAMPLE_FORMAT_S64P:
		return "SAMPLE_FORMAT_S64P"
	default:
		return "[?? Invalid SampleFormat value]"
	}
}

func (v ChannelLayout) String() string {
	switch v {
	case CHANNEL_LAYOUT_NONE:
		return "CHANNEL_LAYOUT_NONE"
	case CHANNEL_LAYOUT_MONO:
		return "CHANNEL_LAYOUT_MONO"
	case CHANNEL_LAYOUT_STEREO:
		return "CHANNEL_LAYOUT_STEREO"
	case CHANNEL_LAYOUT_2POINT1:
		return "CHANNEL_LAYOUT_2POINT1"
	case CHANNEL_LAYOUT_2_1:
		return "CHANNEL_LAYOUT_2_1"
	case CHANNEL_LAYOUT_SURROUND:
		return "CHANNEL_LAYOUT_SURROUND"
	case CHANNEL_LAYOUT_3POINT1:
		return "CHANNEL_LAYOUT_3POINT1"
	case CHANNEL_LAYOUT_4POINT0:
		return "CHANNEL_LAYOUT_4POINT0"
	case CHANNEL_LAYOUT_4POINT1:
		return "CHANNEL_LAYOUT_4POINT1"
	case CHANNEL_LAYOUT_2_2:
		return "CHANNEL_LAYOUT_2_2"
	case CHANNEL_LAYOUT_QUAD:
		return "CHANNEL_LAYOUT_QUAD"
	case CHANNEL_LAYOUT_5POINT0:
		return "CHANNEL_LAYOUT_5POINT0"
	case CHANNEL_LAYOUT_5POINT1:
		return "CHANNEL_LAYOUT_5POINT1"
	case CHANNEL_LAYOUT_5POINT0_BACK:
		return "CHANNEL_LAYOUT_5POINT0_BACK"
	case CHANNEL_LAYOUT_5POINT1_BACK:
		return "CHANNEL_LAYOUT_5POINT1_BACK"
	case CHANNEL_LAYOUT_6POINT0:
		return "CHANNEL_LAYOUT_6POINT0"
	case CHANNEL_LAYOUT_6POINT0_FRONT:
		return "CHANNEL_LAYOUT_6POINT0_FRONT"
	case CHANNEL_LAYOUT_HEXAGONAL:
		return "CHANNEL_LAYOUT_HEXAGONAL"
	case CHANNEL_LAYOUT_6POINT1:
		return "CHANNEL_LAYOUT_6POINT1"
	case CHANNEL_LAYOUT_6POINT1_BACK:
		return "CHANNEL_LAYOUT_6POINT1_BACK"
	case CHANNEL_LAYOUT_6POINT1_FRONT:
		return "CHANNEL_LAYOUT_6POINT1_FRONT"
	case CHANNEL_LAYOUT_7POINT0:
		return "CHANNEL_LAYOUT_7POINT0"
	case CHANNEL_LAYOUT_7POINT0_FRONT:
		return "CHANNEL_LAYOUT_7POINT0_FRONT"
	case CHANNEL_LAYOUT_7POINT1:
		return "CHANNEL_LAYOUT_7POINT1"
	case CHANNEL_LAYOUT_7POINT1_WIDE:
		return "CHANNEL_LAYOUT_7POINT1_WIDE"
	case CHANNEL_LAYOUT_7POINT1_WIDE_BACK:
		return "CHANNEL_LAYOUT_7POINT1_WIDE_BACK"
	case CHANNEL_LAYOUT_OCTAGONAL:
		return "CHANNEL_LAYOUT_OCTAGONAL"
	case CHANNEL_LAYOUT_HEXADECAGONAL:
		return "CHANNEL_LAYOUT_HEXADECAGONAL"
	case CHANNEL_LAYOUT_STEREO_DOWNMIX:
		return "CHANNEL_LAYOUT_STEREO_DOWNMIX"
	case CHANNEL_LAYOUT_22POINT2:
		return "CHANNEL_LAYOUT_22POINT2"
	case CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER:
		return "CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER"
	default:
		return "[?? Invalid ChannelLayout value]"
	}
}
