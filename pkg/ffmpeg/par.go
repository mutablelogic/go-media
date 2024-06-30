package ffmpeg

import (
	"encoding/json"
	"fmt"
	"slices"

	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Par struct {
	ff.AVCodecParameters
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewAudioPar(samplefmt string, channellayout string, samplerate int) (*Par, error) {
	par := new(Par)
	par.SetCodecType(ff.AVMEDIA_TYPE_AUDIO)

	// Sample Format
	if samplefmt_ := ff.AVUtil_get_sample_fmt(samplefmt); samplefmt_ == ff.AV_SAMPLE_FMT_NONE {
		return nil, ErrBadParameter.Withf("unknown sample format %q", samplefmt)
	} else {
		par.SetSampleFormat(samplefmt_)
	}

	// Channel layout
	var ch ff.AVChannelLayout
	if err := ff.AVUtil_channel_layout_from_string(&ch, channellayout); err != nil {
		return nil, ErrBadParameter.Withf("channel layout %q", channellayout)
	} else if err := par.SetChannelLayout(ch); err != nil {
		return nil, err
	}

	// Sample rate
	if samplerate <= 0 {
		return nil, ErrBadParameter.Withf("negative or zero samplerate %v", samplerate)
	} else {
		par.SetSamplerate(samplerate)
	}

	// Return success
	return par, nil
}

func NewVideoPar(pixfmt string, size string) (*Par, error) {
	par := new(Par)
	par.SetCodecType(ff.AVMEDIA_TYPE_VIDEO)

	// Pixel Format
	if pixfmt_ := ff.AVUtil_get_pix_fmt(pixfmt); pixfmt_ == ff.AV_PIX_FMT_NONE {
		return nil, ErrBadParameter.Withf("unknown pixel format %q", pixfmt)
	} else {
		par.SetPixelFormat(pixfmt_)
	}

	// Frame size
	if w, h, err := ff.AVUtil_parse_video_size(size); err != nil {
		return nil, ErrBadParameter.Withf("size %q", size)
	} else {
		par.SetWidth(w)
		par.SetHeight(h)
	}

	// Set default sample aspect ratio
	par.SetSampleAspectRatio(ff.AVUtil_rational(1, 1))

	// Return success
	return par, nil
}

func AudioPar(samplefmt string, channellayout string, samplerate int) *Par {
	if par, err := NewAudioPar(samplefmt, channellayout, samplerate); err != nil {
		panic(err)
	} else {
		return par
	}
}

func VideoPar(pixfmt string, size string) *Par {
	if par, err := NewVideoPar(pixfmt, size); err != nil {
		panic(err)
	} else {
		return par
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *Par) MarshalJSON() ([]byte, error) {
	return json.Marshal(ctx.AVCodecParameters)
}

func (ctx *Par) String() string {
	data, _ := json.MarshalIndent(ctx, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (ctx *Par) ValidateFromCodec(codec *ff.AVCodecContext) error {
	switch codec.Codec().Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		return ctx.validateAudioCodec(codec)
	case ff.AVMEDIA_TYPE_VIDEO:
		return ctx.validateVideoCodec(codec)
	}
	return nil
}

func (ctx *Par) CopyToCodec(codec *ff.AVCodecContext) error {
	switch codec.Codec().Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		return ctx.copyAudioCodec(codec)
	case ff.AVMEDIA_TYPE_VIDEO:
		return ctx.copyVideoCodec(codec)
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (ctx *Par) copyAudioCodec(codec *ff.AVCodecContext) error {
	codec.SetSampleFormat(ctx.SampleFormat())
	codec.SetSampleRate(ctx.Samplerate())
	if err := codec.SetChannelLayout(ctx.ChannelLayout()); err != nil {
		return err
	}
	return nil
}

func (ctx *Par) validateAudioCodec(codec *ff.AVCodecContext) error {
	sampleformats := codec.Codec().SampleFormats()
	samplerates := codec.Codec().SupportedSamplerates()
	channellayouts := codec.Codec().ChannelLayouts()

	// First we set params from the codec which are not already set
	if ctx.SampleFormat() == ff.AV_SAMPLE_FMT_NONE {
		if len(sampleformats) > 0 {
			ctx.SetSampleFormat(sampleformats[0])
		}
	}
	if ctx.Samplerate() == 0 {
		if len(samplerates) > 0 {
			ctx.SetSamplerate(samplerates[0])
		}
	}
	if ctx.ChannelLayout().NumChannels() == 0 {
		if len(channellayouts) > 0 {
			ctx.SetChannelLayout(channellayouts[0])
		}
	}

	// Then we check to make sure the parameters are compatible with
	// the codec
	if len(sampleformats) > 0 {
		if !slices.Contains(sampleformats, ctx.SampleFormat()) {
			return ErrBadParameter.Withf("unsupported sample format %v", ctx.SampleFormat())
		}
	} else if ctx.SampleFormat() == ff.AV_SAMPLE_FMT_NONE {
		return ErrBadParameter.With("sample format not set")
	}
	if len(samplerates) > 0 {
		if !slices.Contains(samplerates, ctx.Samplerate()) {
			return ErrBadParameter.Withf("unsupported samplerate %v", ctx.Samplerate())
		}
	} else if ctx.Samplerate() == 0 {
		return ErrBadParameter.With("samplerate not set")
	}
	if len(channellayouts) > 0 {
		valid := false
		for _, ch := range channellayouts {
			chctx := ctx.ChannelLayout()
			if ff.AVUtil_channel_layout_compare(&ch, &chctx) {
				valid = true
				break
			}
		}
		if !valid {
			return ErrBadParameter.Withf("unsupported channel layout %v", ctx.ChannelLayout())
		}
	} else if ctx.ChannelLayout().NumChannels() == 0 {
		return ErrBadParameter.With("channel layout not set")
	}

	// Validated
	return nil
}

func (ctx *Par) copyVideoCodec(codec *ff.AVCodecContext) error {
	fmt.Println("TODO: copyVideoCodec")
	return nil
}

func (ctx *Par) validateVideoCodec(codec *ff.AVCodecContext) error {
	fmt.Println("TODO: validateVideoCodec")
	return nil
}
