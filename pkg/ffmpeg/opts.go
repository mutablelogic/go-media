package ffmpeg

import (
	// Namespace imports

	. "github.com/djthorpe/go-errors"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Opt func(*opts) error

type opts struct {
	// Resample/resize options
	force bool

	// Format options
	oformat *ffmpeg.AVOutputFormat

	// Audio options
	sample_fmt ffmpeg.AVSampleFormat
	ch         ffmpeg.AVChannelLayout
	samplerate int

	// Video options
	pix_fmt       ffmpeg.AVPixelFormat
	width, height int
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Output format from name or url
func OptOutputFormat(name string) Opt {
	return func(o *opts) error {
		// By name
		if oformat := ffmpeg.AVFormat_guess_format(name, name, name); oformat != nil {
			o.oformat = oformat
		} else {
			return ErrBadParameter.Withf("invalid output format %q", name)
		}
		return nil
	}
}

// Force resampling and resizing on decode, even if the input and output
// parameters are the same
func OptForce() Opt {
	return func(o *opts) error {
		o.force = true
		return nil
	}
}

// Pixel format of the output frame
func OptPixFormat(format string) Opt {
	return func(o *opts) error {
		fmt := ffmpeg.AVUtil_get_pix_fmt(format)
		if fmt == ffmpeg.AV_PIX_FMT_NONE {
			return ErrBadParameter.Withf("invalid pixel format %q", format)
		}
		o.pix_fmt = fmt
		return nil
	}
}

// Width and height of the output frame
func OptWidthHeight(w, h int) Opt {
	return func(o *opts) error {
		if w <= 0 || h <= 0 {
			return ErrBadParameter.Withf("invalid width %v or height %v", w, h)
		}
		o.width = w
		o.height = h
		return nil
	}
}

// Frame size
func OptFrameSize(size string) Opt {
	return func(o *opts) error {
		w, h, err := ffmpeg.AVUtil_parse_video_size(size)
		if err != nil {
			return ErrBadParameter.Withf("invalid frame size %q", size)
		}
		o.width = w
		o.height = h
		return nil
	}
}

// Channel layout
func OptChannelLayout(layout string) Opt {
	return func(o *opts) error {
		return ffmpeg.AVUtil_channel_layout_from_string(&o.ch, layout)
	}
}

// Nuumber of channels
func OptChannels(ch int) Opt {
	return func(o *opts) error {
		if ch <= 0 || ch > 64 {
			return ErrBadParameter.Withf("invalid number of channels %v", ch)
		}
		ffmpeg.AVUtil_channel_layout_default(&o.ch, ch)
		return nil
	}
}

// Sample Rate
func OptSampleRate(rate int) Opt {
	return func(o *opts) error {
		if rate <= 0 {
			return ErrBadParameter.Withf("invalid sample rate %v", rate)
		}
		o.samplerate = rate
		return nil
	}
}

// Sample format
func OptSampleFormat(format string) Opt {
	return func(o *opts) error {
		fmt := ffmpeg.AVUtil_get_sample_fmt(format)
		if fmt == ffmpeg.AV_SAMPLE_FMT_NONE {
			return ErrBadParameter.Withf("invalid sample format %q", format)
		}
		o.sample_fmt = fmt
		return nil
	}
}
