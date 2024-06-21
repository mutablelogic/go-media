package media

import (
	"encoding/json"
	"math"

	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type par struct {
	t MediaType
	audiopar
	videopar
}

type audiopar struct {
	Ch           ff.AVChannelLayout
	SampleFormat ff.AVSampleFormat
	Samplerate   int
}

type videopar struct {
	PixelFormat ff.AVPixelFormat
	Width       int
	Height      int
	Framerate   ff.AVRational
}

var _ Parameters = (*par)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create new parameters for audio sampling from number of channels, sample format and sample rate in Hz
func newAudioParameters(numchannels int, samplefmt string, samplerate int) (*par, error) {
	// Get channel layout from number of channels
	var ch ff.AVChannelLayout
	ff.AVUtil_channel_layout_default(&ch, numchannels)
	if name, err := ff.AVUtil_channel_layout_describe(&ch); err != nil {
		return nil, err
	} else {
		return newAudioParametersEx(name, samplefmt, samplerate)
	}
}

// Create new parameters for audio sampling from a channel layout name, sample format and sample rate in Hz
func newAudioParametersEx(channels string, samplefmt string, samplerate int) (*par, error) {
	par := new(par)
	par.t = AUDIO

	// Set the  parameters
	if err := ff.AVUtil_channel_layout_from_string(&par.audiopar.Ch, channels); err != nil {
		return nil, err
	}
	if fmt := ff.AVUtil_get_sample_fmt(samplefmt); fmt == ff.AV_SAMPLE_FMT_NONE {
		return nil, ErrBadParameter.Withf("sample format %q", samplefmt)
	} else {
		par.audiopar.SampleFormat = fmt
	}
	if samplerate <= 0 {
		return nil, ErrBadParameter.Withf("samplerate %v", samplerate)
	} else {
		par.audiopar.Samplerate = samplerate
	}

	// Return success
	return par, nil
}

// Create new parameters for video from a width, height, pixel format and framerate in fps
func newVideoParametersEx(width int, height int, pixelfmt string, framerate float64) (*par, error) {
	par := new(par)
	par.t = VIDEO

	// Set the  parameters
	if width <= 0 {
		// Negative widths might mean "flip" but not tested yet
		return nil, ErrBadParameter.Withf("width %v", width)
	} else {
		par.videopar.Width = width
	}
	if height <= 0 {
		// Negative heights might mean "flip" but not tested yet
		return nil, ErrBadParameter.Withf("height %v", height)
	} else {
		par.videopar.Height = height
	}
	if fmt := ff.AVUtil_get_pix_fmt(pixelfmt); fmt == ff.AV_PIX_FMT_NONE {
		return nil, ErrBadParameter.Withf("pixel format %q", pixelfmt)
	} else {
		par.videopar.PixelFormat = fmt
	}
	if framerate <= 0 {
		return nil, ErrBadParameter.Withf("framerate %v", framerate)
	} else {
		par.videopar.Framerate = ff.AVUtil_rational_d2q(1/framerate, 0)
	}

	// Return success
	return par, nil
}

// Create new parameters for video from a frame size, pixel format and framerate in fps
func newVideoParameters(frame string, pixelfmt string, framerate float64) (*par, error) {
	// Parse the frame size
	w, h, err := ff.AVUtil_parse_video_size(frame)
	if err != nil {
		return nil, err
	}
	return newVideoParametersEx(w, h, pixelfmt, framerate)
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (par *par) MarshalJSON() ([]byte, error) {
	if par.t == AUDIO {
		return json.Marshal(par.audiopar)
	} else {
		return json.Marshal(par.videopar)
	}
}

func (par *par) String() string {
	data, _ := json.MarshalIndent(par, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return type
func (par *par) Type() MediaType {
	return par.t
}

// Return the channel layout
func (par *par) ChannelLayout() string {
	if name, err := ff.AVUtil_channel_layout_describe(&par.audiopar.Ch); err != nil {
		return ""
	} else {
		return name
	}
}

// Return the sample format
func (par *par) SampleFormat() string {
	return ff.AVUtil_get_sample_fmt_name(par.audiopar.SampleFormat)
}

// Return the sample rate (Hz)
func (par *par) Samplerate() int {
	return par.audiopar.Samplerate
}

// Return the sample format

// Return the width of the video frame
func (par *par) Width() int {
	return par.videopar.Width
}

// Return the height of the video frame
func (par *par) Height() int {
	return par.videopar.Height
}

// Return the pixel format
func (par *par) PixelFormat() string {
	return ff.AVUtil_get_pix_fmt_name(par.videopar.PixelFormat)
}

// Return the frame rate (fps)
func (par *par) Framerate() float64 {
	if v := par.videopar.Framerate.Float(1); v == 0 {
		return math.Inf(1)
	} else {
		return 1 / v
	}
}
