package ffmpeg

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
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
		return nil, fmt.Errorf("unknown sample format %q", samplefmt)
	} else {
		par.SetSampleFormat(samplefmt_)
	}

	// Channel layout
	var ch ff.AVChannelLayout
	if err := ff.AVUtil_channel_layout_from_string(&ch, channellayout); err != nil {
		return nil, fmt.Errorf("invalid channel layout %q: %w", channellayout, err)
	}
	if err := par.SetChannelLayout(ch); err != nil {
		return nil, err
	}

	// Sample rate
	if samplerate <= 0 {
		return nil, fmt.Errorf("invalid samplerate %d: must be positive", samplerate)
	}
	par.SetSampleRate(samplerate)

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
		return nil, fmt.Errorf("unknown pixel format %q", pixfmt)
	} else {
		par.SetPixelFormat(pixfmt_)
	}

	// Frame size
	w, h, err := ff.AVUtil_parse_video_size(size)
	if err != nil {
		return nil, fmt.Errorf("invalid size %q: %w", size, err)
	}
	par.SetWidth(w)
	par.SetHeight(h)

	// Frame rate and timebase
	if framerate < 0 {
		return nil, fmt.Errorf("invalid framerate %v: must be non-negative", framerate)
	}
	if framerate > 0 {
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

func (par *Par) MarshalJSON() ([]byte, error) {
	if par == nil {
		return []byte("null"), nil
	}
	return json.Marshal(jsonPar{
		AVCodecParameters: par.AVCodecParameters,
		Timebase:          par.timebase,
		Opts:              par.opts,
	})
}

func (par *Par) String() string {
	if par == nil {
		return "<nil>"
	}
	data, err := json.MarshalIndent(par, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (par *Par) Type() media.Type {
	if par == nil {
		return media.UNKNOWN
	}
	switch par.CodecType() {
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

func (par *Par) WidthHeight() string {
	if par == nil {
		return "0x0"
	}
	return fmt.Sprintf("%dx%d", par.Width(), par.Height())
}

func (par *Par) FrameRate() float64 {
	if par == nil || par.timebase.Num() == 0 || par.timebase.Den() == 0 {
		return 0
	}
	return ff.AVUtil_rational_q2d(ff.AVUtil_rational_invert(par.timebase))
}

func (par *Par) ValidateFromCodec(codec *ff.AVCodec) error {
	if par == nil {
		return errors.New("par is nil")
	}
	if codec == nil {
		return errors.New("codec is nil")
	}
	switch codec.Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		return par.validateAudioCodec(codec)
	case ff.AVMEDIA_TYPE_VIDEO:
		return par.validateVideoCodec(codec)
	}
	return nil
}

func (par *Par) CopyToCodecContext(codec *ff.AVCodecContext) error {
	if par == nil {
		return errors.New("par is nil")
	}
	if codec == nil {
		return errors.New("codec is nil")
	}
	switch codec.Codec().Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		return par.copyAudioCodec(codec)
	case ff.AVMEDIA_TYPE_VIDEO:
		return par.copyVideoCodec(codec)
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Convert options to a dictionary, which needs to be freed after use by the caller
func (par *Par) optionsToDict() *ff.AVDictionary {
	dict := ff.AVUtil_dict_alloc()
	for _, opt := range par.opts {
		if err := ff.AVUtil_dict_set(dict, opt.Key(), opt.Value(), ff.AV_DICT_APPEND); err != nil {
			ff.AVUtil_dict_free(dict)
			return nil
		}
	}
	return dict
}

func (par *Par) copyAudioCodec(codec *ff.AVCodecContext) error {
	codec.SetSampleFormat(par.SampleFormat())
	codec.SetSampleRate(par.SampleRate())
	codec.SetTimeBase(ff.AVUtil_rational(1, par.SampleRate()))
	return codec.SetChannelLayout(par.ChannelLayout())
}

func (par *Par) validateAudioCodec(codec *ff.AVCodec) error {
	sampleformats := codec.SampleFormats()
	samplerates := codec.SupportedSamplerates()
	channellayouts := codec.ChannelLayouts()

	// First set params from the codec which are not already set
	if par.SampleFormat() == ff.AV_SAMPLE_FMT_NONE && len(sampleformats) > 0 {
		par.SetSampleFormat(sampleformats[0])
	}
	if par.SampleRate() == 0 && len(samplerates) > 0 {
		par.SetSampleRate(samplerates[0])
	}
	if par.ChannelLayout().NumChannels() == 0 && len(channellayouts) > 0 {
		par.SetChannelLayout(channellayouts[0])
	}

	// Then check parameters are compatible with the codec
	if err := par.validateSampleFormat(sampleformats); err != nil {
		return err
	}
	if err := par.validateSampleRate(samplerates); err != nil {
		return err
	}
	return par.validateChannelLayout(channellayouts)
}

// Helper methods for audio validation
func (par *Par) validateSampleFormat(supported []ff.AVSampleFormat) error {
	if len(supported) > 0 {
		if !slices.Contains(supported, par.SampleFormat()) {
			return fmt.Errorf("unsupported sample format %v", par.SampleFormat())
		}
	} else if par.SampleFormat() == ff.AV_SAMPLE_FMT_NONE {
		return errors.New("sample format not set")
	}
	return nil
}

func (par *Par) validateSampleRate(supported []int) error {
	if len(supported) > 0 {
		if !slices.Contains(supported, par.SampleRate()) {
			return fmt.Errorf("unsupported samplerate %v", par.SampleRate())
		}
	} else if par.SampleRate() == 0 {
		return errors.New("samplerate not set")
	}
	return nil
}

func (par *Par) validateChannelLayout(supported []ff.AVChannelLayout) error {
	if len(supported) > 0 {
		valid := false
		parCh := par.ChannelLayout()
		for _, ch := range supported {
			if ff.AVUtil_channel_layout_compare(&ch, &parCh) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("unsupported channel layout %v", par.ChannelLayout())
		}
	} else if par.ChannelLayout().NumChannels() == 0 {
		return errors.New("channel layout not set")
	}
	return nil
}

func (par *Par) copyVideoCodec(codec *ff.AVCodecContext) error {
	codec.SetPixFmt(par.PixelFormat())
	codec.SetWidth(par.Width())
	codec.SetHeight(par.Height())
	codec.SetSampleAspectRatio(par.SampleAspectRatio())
	codec.SetTimeBase(par.timebase)
	return nil
}

func (par *Par) validateVideoCodec(codec *ff.AVCodec) error {
	pixelformats := codec.PixelFormats()
	framerates := codec.SupportedFramerates()

	// First set params from the codec which are not already set
	if par.PixelFormat() == ff.AV_PIX_FMT_NONE && len(pixelformats) > 0 {
		par.SetPixelFormat(pixelformats[0])
	}
	if (par.timebase.Num() == 0 || par.timebase.Den() == 0) && len(framerates) > 0 {
		par.timebase = ff.AVUtil_rational_invert(framerates[0])
	}

	// Then check parameters are compatible with the codec
	if err := par.validatePixelFormat(pixelformats); err != nil {
		return err
	}
	if err := par.validateDimensions(); err != nil {
		return err
	}
	par.ensureSampleAspectRatio()
	return par.validateFrameRate(framerates)
}

// Helper methods for video validation
func (par *Par) validatePixelFormat(supported []ff.AVPixelFormat) error {
	if len(supported) > 0 {
		if !slices.Contains(supported, par.PixelFormat()) {
			return fmt.Errorf("unsupported pixel format %v", par.PixelFormat())
		}
	} else if par.PixelFormat() == ff.AV_PIX_FMT_NONE {
		return errors.New("pixel format not set")
	}
	return nil
}

func (par *Par) validateDimensions() error {
	if par.Width() == 0 || par.Height() == 0 {
		return fmt.Errorf("invalid width %v or height %v", par.Width(), par.Height())
	}
	return nil
}

func (par *Par) ensureSampleAspectRatio() {
	if par.SampleAspectRatio().Num() == 0 || par.SampleAspectRatio().Den() == 0 {
		par.SetSampleAspectRatio(ff.AVUtil_rational(1, 1))
	}
}

func (par *Par) validateFrameRate(supported []ff.AVRational) error {
	if par.timebase.Num() == 0 || par.timebase.Den() == 0 {
		return errors.New("framerate not set")
	}
	if len(supported) > 0 {
		valid := false
		parFr := ff.AVUtil_rational_invert(par.timebase)
		for _, fr := range supported {
			if ff.AVUtil_rational_equal(fr, parFr) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("unsupported framerate %v", parFr)
		}
	}
	return nil
}
