package ffmpeg

import (
	"encoding/json"
	"fmt"
	"slices"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Par struct {
	ff.AVCodecParameters
	opts     []media.Metadata
	timebase ff.AVRational
}

type jsonPar struct {
	ff.AVCodecParameters
	Timebase ff.AVRational    `json:"timebase"`
	Opts     []media.Metadata `json:"options"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create new audio parameters with sample format, channel layout and sample rate
// plus any additional options which is used for creating a stream
func NewAudioPar(samplefmt string, channellayout string, samplerate int, opts ...media.Metadata) (*Par, error) {
	par := new(Par)
	par.SetCodecType(ff.AVMEDIA_TYPE_AUDIO)
	par.opts = opts

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

// Create new video parameters with pixel format, frame size, framerate
// plus any additional options which is used for creating a stream
func NewVideoPar(pixfmt string, size string, framerate float64, opts ...media.Metadata) (*Par, error) {
	par := new(Par)
	par.SetCodecType(ff.AVMEDIA_TYPE_VIDEO)
	par.opts = opts

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

	// Frame rate and timebase
	if framerate < 0 {
		return nil, ErrBadParameter.Withf("negative framerate %v", framerate)
	} else if framerate > 0 {
		par.timebase = ff.AVUtil_rational_invert(ff.AVUtil_rational_d2q(framerate, 1<<24))
	}

	// Set default sample aspect ratio
	par.SetSampleAspectRatio(ff.AVUtil_rational(1, 1))

	// Return success
	return par, nil
}

// Create audio parameters. If there is an error, then this function will panic
func AudioPar(samplefmt string, channellayout string, samplerate int, opts ...media.Metadata) *Par {
	if par, err := NewAudioPar(samplefmt, channellayout, samplerate, opts...); err != nil {
		panic(err)
	} else {
		return par
	}
}

// Create video parameters. If there is an error, then this function will panic
func VideoPar(pixfmt string, size string, framerate float64, opts ...media.Metadata) *Par {
	if par, err := NewVideoPar(pixfmt, size, framerate, opts...); err != nil {
		panic(err)
	} else {
		return par
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *Par) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonPar{
		AVCodecParameters: ctx.AVCodecParameters,
		Timebase:          ctx.timebase,
		Opts:              ctx.opts,
	})
}

func (ctx *Par) String() string {
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err.Error()
	} else {
		return string(data)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (ctx *Par) Type() media.Type {
	switch ctx.CodecType() {
	case ff.AVMEDIA_TYPE_AUDIO:
		return media.AUDIO
	case ff.AVMEDIA_TYPE_VIDEO:
		return media.VIDEO
	case ff.AVMEDIA_TYPE_SUBTITLE:
		return media.SUBTITLE
	case ff.AVMEDIA_TYPE_DATA:
		return media.DATA
	default:
		return media.UNKNOWN
	}
}

func (ctx *Par) WidthHeight() string {
	return fmt.Sprintf("%dx%d", ctx.Width(), ctx.Height())
}

func (ctx *Par) FrameRate() float64 {
	if ctx.timebase.Num() == 0 || ctx.timebase.Den() == 0 {
		return 0
	}
	return ff.AVUtil_rational_q2d(ff.AVUtil_rational_invert(ctx.timebase))
}

func (ctx *Par) ValidateFromCodec(codec *ff.AVCodec) error {
	switch codec.Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		return ctx.validateAudioCodec(codec)
	case ff.AVMEDIA_TYPE_VIDEO:
		return ctx.validateVideoCodec(codec)
	}
	return nil
}

func (ctx *Par) CopyToCodecContext(codec *ff.AVCodecContext) error {
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

// Return options as a dictionary, which needs to be freed after use
// by the caller method
func (ctx *Par) newOpts() *ff.AVDictionary {
	dict := ff.AVUtil_dict_alloc()
	for _, opt := range ctx.opts {
		if err := ff.AVUtil_dict_set(dict, opt.Key(), opt.Value(), ff.AV_DICT_APPEND); err != nil {
			ff.AVUtil_dict_free(dict)
			return nil
		}
	}
	return dict
}

func (ctx *Par) copyAudioCodec(codec *ff.AVCodecContext) error {
	codec.SetSampleFormat(ctx.SampleFormat())
	codec.SetSampleRate(ctx.Samplerate())
	codec.SetTimeBase(ff.AVUtil_rational(1, ctx.Samplerate()))
	if err := codec.SetChannelLayout(ctx.ChannelLayout()); err != nil {
		return err
	}
	return nil
}

func (ctx *Par) validateAudioCodec(codec *ff.AVCodec) error {
	sampleformats := codec.SampleFormats()
	samplerates := codec.SupportedSamplerates()
	channellayouts := codec.ChannelLayouts()

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
	codec.SetPixFmt(ctx.PixelFormat())
	codec.SetWidth(ctx.Width())
	codec.SetHeight(ctx.Height())
	codec.SetSampleAspectRatio(ctx.SampleAspectRatio())
	codec.SetTimeBase(ctx.timebase)
	return nil
}

func (ctx *Par) validateVideoCodec(codec *ff.AVCodec) error {
	pixelformats := codec.PixelFormats()
	framerates := codec.SupportedFramerates()

	// First we set params from the codec which are not already set
	if ctx.PixelFormat() == ff.AV_PIX_FMT_NONE {
		if len(pixelformats) > 0 {
			ctx.SetPixelFormat(pixelformats[0])
		}
	}
	if ctx.timebase.Num() == 0 || ctx.timebase.Den() == 0 {
		if len(framerates) > 0 {
			ctx.timebase = ff.AVUtil_rational_invert(framerates[0])
		}
	}

	// Then we check to make sure the parameters are compatible with
	// the codec
	if len(pixelformats) > 0 {
		if !slices.Contains(pixelformats, ctx.PixelFormat()) {
			return ErrBadParameter.Withf("unsupported pixel format %v", ctx.PixelFormat())
		}
	} else if ctx.PixelFormat() == ff.AV_PIX_FMT_NONE {
		return ErrBadParameter.With("pixel format not set")
	}
	if ctx.Width() == 0 || ctx.Height() == 0 {
		return ErrBadParameter.Withf("invalid width %v or height %v", ctx.Width(), ctx.Height())
	}
	if ctx.SampleAspectRatio().Num() == 0 || ctx.SampleAspectRatio().Den() == 0 {
		ctx.SetSampleAspectRatio(ff.AVUtil_rational(1, 1))
	}
	if ctx.timebase.Num() == 0 || ctx.timebase.Den() == 0 {
		return ErrBadParameter.With("framerate not set")
	} else if len(framerates) > 0 {
		valid := false
		for _, fr := range framerates {
			if ff.AVUtil_rational_equal(fr, ff.AVUtil_rational_invert(ctx.timebase)) {
				valid = true
				break
			}
		}
		if !valid {
			return ErrBadParameter.Withf("unsupported framerate %v", ff.AVUtil_rational_invert(ctx.timebase))
		}
	}

	// Return success
	return nil
}
