package media

import (
	"encoding/json"

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
	planepar
}

type codecpar struct {
	Framerate ff.AVRational
}

type audiopar struct {
	Ch           ff.AVChannelLayout `json:"channel_layout,omitempty"`
	SampleFormat ff.AVSampleFormat  `json:"sample_format,omitempty"`
	Samplerate   int                `json:"samplerate,omitempty"`
}

type videopar struct {
	PixelFormat ff.AVPixelFormat `json:"pixel_format,omitempty"`
	Width       int              `json:"width,omitempty"`
	Height      int              `json:"height,omitempty"`
}

type planepar struct {
	NumPlanes int `json:"num_video_planes,omitempty"`
}

type timingpar struct {
	Framerate ff.AVRational `json:"framerate,omitempty"`
	Pts       int64         `json:"pts,omitempty"`
	TimeBase  ff.AVRational `json:"time_base,omitempty"`
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
		return nil, ErrBadParameter.Withf("unknown sample format %q", samplefmt)
	} else {
		par.audiopar.SampleFormat = fmt
	}
	if samplerate <= 0 {
		return nil, ErrBadParameter.Withf("negative or zero samplerate %v", samplerate)
	} else {
		par.audiopar.Samplerate = samplerate
	}
	par.planepar.NumPlanes = par.NumPlanes()

	// Return success
	return par, nil
}

func newCodecAudioParameters(codec *ff.AVCodecParameters) *par {
	par := new(par)
	par.t = AUDIO
	par.audiopar.Ch = codec.ChannelLayout()
	par.audiopar.SampleFormat = codec.SampleFormat()
	par.audiopar.Samplerate = codec.Samplerate()
	par.planepar.NumPlanes = par.NumPlanes()
	return par
}

func newCodecVideoParameters(codec *ff.AVCodecParameters) *par {
	par := new(par)
	par.t = VIDEO
	par.videopar.Width = codec.Width()
	par.videopar.Height = codec.Height()
	par.videopar.PixelFormat = codec.PixelFormat()
	par.planepar.NumPlanes = par.NumPlanes()
	return par
}

// Create new parameters for video from a width, height and pixel format
func newVideoParametersEx(width int, height int, pixelfmt string) (*par, error) {
	par := new(par)
	par.t = VIDEO

	// Set the  parameters
	if width <= 0 {
		// Negative widths might mean "flip" but not tested yet
		return nil, ErrBadParameter.Withf("negative or zero width %v", width)
	} else {
		par.videopar.Width = width
	}
	if height <= 0 {
		// Negative heights might mean "flip" but not tested yet
		return nil, ErrBadParameter.Withf("negative or zero height %v", height)
	} else {
		par.videopar.Height = height
	}
	if fmt := ff.AVUtil_get_pix_fmt(pixelfmt); fmt == ff.AV_PIX_FMT_NONE {
		return nil, ErrBadParameter.Withf("unknown pixel format %q", pixelfmt)
	} else {
		par.videopar.PixelFormat = fmt
	}
	par.planepar.NumPlanes = par.NumPlanes()

	// Return success
	return par, nil
}

// Create new parameters for video from a frame size, pixel format
func newVideoParameters(frame string, pixelfmt string) (*par, error) {
	// Parse the frame size
	w, h, err := ff.AVUtil_parse_video_size(frame)
	if err != nil {
		return nil, err
	}
	return newVideoParametersEx(w, h, pixelfmt)
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

// Return number of planes for a specific PixelFormat
// or SampleFormat and ChannelLayout combination
func (par *par) NumPlanes() int {
	switch par.t {
	case AUDIO:
		if ff.AVUtil_sample_fmt_is_planar(par.audiopar.SampleFormat) {
			return par.audiopar.Ch.NumChannels()
		} else {
			return 1
		}
	case VIDEO:
		return ff.AVUtil_pix_fmt_count_planes(par.videopar.PixelFormat)
	}
	return 0
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
