package ffmpeg

import (
	// Package imports
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Opt func(*opts) error

type opts struct {
	// Resample/resize options
	force bool
	par   *Par

	// Writer options
	oformat  *ffmpeg.AVOutputFormat
	streams  map[int]*Par
	metadata []*Metadata
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newOpts() *opts {
	return &opts{
		par:     new(Par),
		streams: make(map[int]*Par),
	}
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

// New audio stream with parameters
func OptAudioStream(stream int, par *Par) Opt {
	return func(o *opts) error {
		if par == nil || par.CodecType() != ffmpeg.AVMEDIA_TYPE_AUDIO {
			return ErrBadParameter.With("invalid audio parameters")
		}
		if stream == 0 {
			stream = len(o.streams) + 1
		}
		if _, exists := o.streams[stream]; exists {
			return ErrDuplicateEntry.Withf("stream %v", stream)
		}
		if stream < 0 {
			return ErrBadParameter.Withf("invalid stream %v", stream)
		}
		o.streams[stream] = par

		// Return success
		return nil
	}
}

// New video stream with parameters
func OptVideoStream(stream int, par *Par) Opt {
	return func(o *opts) error {
		if par == nil || par.CodecType() != ffmpeg.AVMEDIA_TYPE_VIDEO {
			return ErrBadParameter.With("invalid video parameters")
		}
		if stream == 0 {
			stream = len(o.streams) + 1
		}
		if _, exists := o.streams[stream]; exists {
			return ErrDuplicateEntry.Withf("stream %v", stream)
		}
		if stream < 0 {
			return ErrBadParameter.Withf("invalid stream %v", stream)
		}
		o.streams[stream] = par

		// Return success
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

// Append metadata to the output file, including artwork
func OptMetadata(entry ...*Metadata) Opt {
	return func(o *opts) error {
		o.metadata = append(o.metadata, entry...)
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
		o.par.SetCodecType(ffmpeg.AVMEDIA_TYPE_VIDEO)
		o.par.SetPixelFormat(fmt)
		return nil
	}
}

// Width and height of the output frame
func OptWidthHeight(w, h int) Opt {
	return func(o *opts) error {
		if w <= 0 || h <= 0 {
			return ErrBadParameter.Withf("invalid width %v or height %v", w, h)
		}
		o.par.SetCodecType(ffmpeg.AVMEDIA_TYPE_VIDEO)
		o.par.SetWidth(w)
		o.par.SetHeight(h)
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
		o.par.SetCodecType(ffmpeg.AVMEDIA_TYPE_VIDEO)
		o.par.SetWidth(w)
		o.par.SetHeight(h)
		return nil
	}
}

// Channel layout
func OptChannelLayout(layout string) Opt {
	return func(o *opts) error {
		var ch ffmpeg.AVChannelLayout
		if err := ffmpeg.AVUtil_channel_layout_from_string(&ch, layout); err != nil {
			return ErrBadParameter.Withf("invalid channel layout %q", layout)
		}
		o.par.SetCodecType(ffmpeg.AVMEDIA_TYPE_AUDIO)
		return o.par.SetChannelLayout(ch)
	}
}

// Nuumber of channels
func OptChannels(num int) Opt {
	return func(o *opts) error {
		var ch ffmpeg.AVChannelLayout
		if num <= 0 || num > 64 {
			return ErrBadParameter.Withf("invalid number of channels %v", num)
		}
		ffmpeg.AVUtil_channel_layout_default(&ch, num)
		o.par.SetCodecType(ffmpeg.AVMEDIA_TYPE_AUDIO)
		return o.par.SetChannelLayout(ch)
	}
}

// Sample Rate
func OptSampleRate(rate int) Opt {
	return func(o *opts) error {
		if rate <= 0 {
			return ErrBadParameter.Withf("invalid sample rate %v", rate)
		}
		o.par.SetCodecType(ffmpeg.AVMEDIA_TYPE_AUDIO)
		o.par.SetSamplerate(rate)
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
		o.par.SetCodecType(ffmpeg.AVMEDIA_TYPE_AUDIO)
		o.par.SetSampleFormat(fmt)
		return nil
	}
}
