package avcodec

import (
	"encoding/json"

	// Packages

	media "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/pkg/metadata"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Codec struct {
	codec   *ff.AVCodec
	context *ff.AVCodecContext
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Return all codecs. Codecs can be either encoders or decoders. The argument
// can be ANY or any combination of INPUT, OUTPUT, VIDEO, AUDIO and SUBTITLE,
// in order to return a subset of codecs.
func Codecs(t media.Type) []media.Metadata {
	result := make([]media.Metadata, 0, 100)
	var opaque uintptr
	for {
		codec := ff.AVCodec_iterate(&opaque)
		if codec == nil {
			break
		}
		// Filter codecs by type
		if t.Is(media.INPUT) && !ff.AVCodec_is_decoder(codec) {
			continue
		}
		if t.Is(media.OUTPUT) && !ff.AVCodec_is_encoder(codec) {
			continue
		}
		if t.Is(media.VIDEO) && codec.Type() != ff.AVMEDIA_TYPE_VIDEO {
			continue
		}
		if t.Is(media.AUDIO) && codec.Type() != ff.AVMEDIA_TYPE_AUDIO {
			continue
		}
		if t.Is(media.SUBTITLE) && codec.Type() != ff.AVMEDIA_TYPE_SUBTITLE {
			continue
		}
		if codec.Capabilities().Is(ff.AV_CODEC_CAP_EXPERIMENTAL) {
			// Skip experimental codecs
			continue
		}
		result = append(result, metadata.New(codec.Name(), &Codec{codec, nil}))
	}
	return result
}

// Return an encoder by name, with additional options. Call Close() to
// release the codec context. Codec options are listed at
// <https://ffmpeg.org/ffmpeg-codecs.html>
func NewEncoder(name string, opts ...Opt) (*Codec, error) {
	ctx := new(Codec)

	// Options
	o, err := applyOptions(opts)
	if err != nil {
		return nil, err
	}

	// Codec context
	if codec := ff.AVCodec_find_encoder_by_name(name); codec == nil {
		return nil, media.ErrBadParameter.Withf("unknown codec %q", name)
	} else if context := ff.AVCodec_alloc_context(codec); context == nil {
		return nil, media.ErrInternalError.Withf("failed to allocate codec context for %q", name)
	} else if err := set_par(context, codec, o); err != nil {
		ff.AVCodec_free_context(context)
		return nil, err
	} else if ff.AVCodec_open(context, codec, nil); err != nil {
		ff.AVCodec_free_context(context)
		return nil, err
	} else {
		ctx.context = context
	}

	// Return success
	return ctx, nil
}

// Return a decoder by name, with additional options. Call Close() to
// release the codec context. Codec options are listed at
// <https://ffmpeg.org/ffmpeg-codecs.html>
func NewDecoder(name string, opts ...Opt) (*Codec, error) {
	ctx := new(Codec)

	// Options
	o, err := applyOptions(opts)
	if err != nil {
		return nil, err
	}

	// Codec context
	if codec := ff.AVCodec_find_decoder_by_name(name); codec == nil {
		return nil, media.ErrBadParameter.Withf("unknown codec %q", name)
	} else if context := ff.AVCodec_alloc_context(codec); context == nil {
		return nil, media.ErrInternalError.Withf("failed to allocate codec context for %q", name)
	} else if err := set_par(context, codec, o); err != nil {
		ff.AVCodec_free_context(context)
		return nil, err
	} else if ff.AVCodec_open(context, codec, nil); err != nil {
		ff.AVCodec_free_context(context)
		return nil, err
	} else {
		ctx.context = context
	}

	// Return success
	return ctx, nil
}

// Release the codec resources
func (ctx *Codec) Close() error {
	ctx.codec = nil
	if ctx != nil && ctx.context != nil {
		ff.AVCodec_free_context(ctx.context)
		ctx.context = nil
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *Codec) MarshalJSON() ([]byte, error) {
	if ctx != nil && ctx.codec != nil {
		return ctx.codec.MarshalJSON()
	}
	if ctx != nil && ctx.context != nil {
		return ctx.context.MarshalJSON()
	}
	return []byte("null"), nil
}

func (ctx *Codec) String() string {
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (ctx *Codec) Type() media.Type {
	var t media.Type
	switch {
	case ctx != nil && ctx.codec != nil:
		if ff.AVCodec_is_decoder(ctx.codec) {
			t |= media.INPUT
		}
		if ff.AVCodec_is_encoder(ctx.codec) {
			t |= media.OUTPUT
		}
		t |= type2type(ctx.codec.Type())
	case ctx != nil && ctx.context != nil:
		if ff.AVCodec_is_decoder(ctx.context.Codec()) {
			t |= media.INPUT
		}
		if ff.AVCodec_is_encoder(ctx.context.Codec()) {
			t |= media.OUTPUT
		}
		t |= type2type(ctx.context.Codec().Type())
	}
	return t
}

func (ctx *Codec) Name() string {
	switch {
	case ctx != nil && ctx.codec != nil:
		return ctx.codec.Name()
	case ctx != nil && ctx.context != nil:
		return ctx.context.Codec().Name()
	}
	return ""
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func type2type(t ff.AVMediaType) media.Type {
	switch t {
	case ff.AVMEDIA_TYPE_AUDIO:
		return media.AUDIO
	case ff.AVMEDIA_TYPE_VIDEO:
		return media.VIDEO
	case ff.AVMEDIA_TYPE_SUBTITLE:
		return media.SUBTITLE
	case ff.AVMEDIA_TYPE_ATTACHMENT:
		return media.DATA
	case ff.AVMEDIA_TYPE_DATA:
		return media.DATA
	}
	return media.NONE
}

func set_par(ctx *ff.AVCodecContext, codec *ff.AVCodec, opt *opt) error {
	switch codec.Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		return set_audio_par(ctx, codec, opt)
	case ff.AVMEDIA_TYPE_VIDEO:
		return set_video_par(ctx, codec, opt)
	case ff.AVMEDIA_TYPE_SUBTITLE:
		return set_subtitle_par(ctx, codec, opt)
	}
	return nil
}

func set_audio_par(ctx *ff.AVCodecContext, codec *ff.AVCodec, opt *opt) error {
	// Channel layout
	if ff.AVUtil_channel_layout_check(&opt.channel_layout) {
		ctx.SetChannelLayout(opt.channel_layout)
	} else if supported_layouts := codec.ChannelLayouts(); len(supported_layouts) > 0 {
		ctx.SetChannelLayout(supported_layouts[0])
	} else {
		ctx.SetChannelLayout(ff.AV_CHANNEL_LAYOUT_MONO)
	}

	// Sample format
	if opt.sample_format != ff.AV_SAMPLE_FMT_NONE {
		ctx.SetSampleFormat(opt.sample_format)
	} else if supported_formats := codec.SampleFormats(); len(supported_formats) > 0 {
		ctx.SetSampleFormat(supported_formats[0])
	}

	// Sample rate
	if opt.sample_rate > 0 {
		ctx.SetSampleRate(opt.sample_rate)
	} else if supported_rates := codec.SupportedSamplerates(); len(supported_rates) > 0 {
		ctx.SetSampleRate(supported_rates[0])
	}

	// TODO: Time base
	ctx.SetTimeBase(ff.AVUtil_rational(1, ctx.SampleRate()))

	return nil
}

func set_video_par(ctx *ff.AVCodecContext, codec *ff.AVCodec, opt *opt) error {
	// Pixel Format
	if opt.pixel_format != ff.AV_PIX_FMT_NONE {
		ctx.SetPixFmt(opt.pixel_format)
	} else if supported_formats := codec.PixelFormats(); len(supported_formats) > 0 {
		ctx.SetPixFmt(supported_formats[0])
	} else {
		ctx.SetPixFmt(ff.AV_PIX_FMT_YUV420P)
	}

	// Frame size
	if opt.width > 0 {
		ctx.SetWidth(opt.width)
	}
	if opt.height > 0 {
		ctx.SetHeight(opt.height)
	}

	// Frame rate
	if !opt.frame_rate.IsZero() {
		ctx.SetFramerate(opt.frame_rate)
	} else if supported_rates := codec.SupportedFramerates(); len(supported_rates) > 0 {
		ctx.SetFramerate(supported_rates[0])
	}

	// Time base
	if frame_rate := ctx.Framerate(); !frame_rate.IsZero() {
		ctx.SetTimeBase(ff.AVUtil_rational_invert(frame_rate))
	}

	return nil
}

func set_subtitle_par(ctx *ff.AVCodecContext, codec *ff.AVCodec, opt *opt) error {
	return nil
}
