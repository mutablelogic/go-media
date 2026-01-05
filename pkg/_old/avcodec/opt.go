package avcodec

import (
	// Packages
	media "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/pkg/metadata"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
	// Audio parameters
	channel_layout ff.AVChannelLayout
	sample_format  ff.AVSampleFormat
	sample_rate    int

	// Video parameters
	width, height int
	pixel_format  ff.AVPixelFormat
	frame_rate    ff.AVRational
	pixel_ratio   ff.AVRational

	// Codec options
	meta []*metadata.Metadata
}

type Opt func(*opt) error

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func applyOptions(opts []Opt) (*opt, error) {
	o := new(opt)
	o.sample_format = ff.AV_SAMPLE_FMT_NONE
	o.pixel_format = ff.AV_PIX_FMT_NONE
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}
	return o, nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func WithOpt(key string, value any) Opt {
	return func(o *opt) error {
		o.meta = append(o.meta, metadata.New(key, value))
		return nil
	}
}

func WithChannelLayout(v string) Opt {
	return func(o *opt) error {
		if err := ff.AVUtil_channel_layout_from_string(&o.channel_layout, v); err != nil {
			return media.ErrBadParameter.Withf("unknown channel layout %q", v)
		}
		return nil
	}
}

func WithSampleFormat(v string) Opt {
	return func(o *opt) error {
		if samplefmt := ff.AVUtil_get_sample_fmt(v); samplefmt == ff.AV_SAMPLE_FMT_NONE {
			return media.ErrBadParameter.Withf("unknown sample format %q", v)
		} else {
			o.sample_format = samplefmt
		}
		return nil
	}

}

func WithSampleRate(v int) Opt {
	return func(o *opt) error {
		if v <= 0 {
			return media.ErrBadParameter.Withf("negative or zero samplerate %v", v)
		} else {
			o.sample_rate = v
		}
		return nil
	}
}

func WithPixelFormat(v string) Opt {
	return func(o *opt) error {
		if pixfmt := ff.AVUtil_get_pix_fmt(v); pixfmt == ff.AV_PIX_FMT_NONE {
			return media.ErrBadParameter.Withf("unknown pixel format %q", v)
		} else {
			o.pixel_format = pixfmt
		}
		return nil
	}
}

func WithFrameSize(v string) Opt {
	return func(o *opt) error {
		if width, height, err := ff.AVUtil_parse_video_size(v); err != nil {
			return media.ErrBadParameter.Withf("size %q", v)
		} else {
			o.width = width
			o.height = height
		}
		return nil
	}
}

func WithFrameRate(num, den int) Opt {
	return func(o *opt) error {
		o.frame_rate = ff.AVUtil_rational(num, den)
		if o.frame_rate.IsZero() {
			return media.ErrBadParameter.Withf("zero frame rate %v/%v", num, den)
		}
		if o.frame_rate.Num() <= 0 || o.frame_rate.Den() <= 0 {
			return media.ErrBadParameter.Withf("negative frame rate %v/%v", num, den)
		}
		return nil
	}
}
