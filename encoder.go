package media

import (

	// Packages
	"fmt"

	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type encoder struct {
	t      MediaType
	ctx    *ff.AVCodecContext
	stream *ff.AVStream
	packet *ff.AVPacket
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create an encoder with the given parameters
func newEncoder(ctx *ff.AVFormatContext, stream_id int, param Parameters) (*encoder, error) {
	encoder := new(encoder)
	par := param.(*par)

	// Get codec
	codec_id := ff.AV_CODEC_ID_NONE
	if param.Type().Is(CODEC) {
		codec_id = par.codecpar.Codec
	} else if par.Type().Is(AUDIO) {
		codec_id = ctx.Output().AudioCodec()
	} else if par.Type().Is(VIDEO) {
		codec_id = ctx.Output().VideoCodec()
	} else if par.Type().Is(SUBTITLE) {
		codec_id = ctx.Output().SubtitleCodec()
	}
	if codec_id == ff.AV_CODEC_ID_NONE {
		return nil, ErrBadParameter.With("no codec specified for stream")
	}

	// Allocate codec
	codec := ff.AVCodec_find_encoder(codec_id)
	if codec == nil {
		return nil, ErrBadParameter.Withf("codec %q cannot encode", codec_id)
	}
	codecctx := ff.AVCodec_alloc_context(codec)
	if codecctx == nil {
		return nil, ErrInternalAppError.With("could not allocate audio codec context")
	} else {
		encoder.ctx = codecctx
	}

	// Create the stream
	if stream := ff.AVFormat_new_stream(ctx, nil); stream == nil {
		ff.AVCodec_free_context(codecctx)
		return nil, ErrInternalAppError.With("could not allocate stream")
	} else {
		stream.SetId(stream_id)
		encoder.stream = stream
	}

	// Set parameters
	switch codec.Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		encoder.t = AUDIO
		// TODO: Check codec supports this configuration

		// Set codec parameters
		if err := codecctx.SetChannelLayout(par.audiopar.Ch); err != nil {
			ff.AVCodec_free_context(codecctx)
			return nil, err
		}
		codecctx.SetSampleFormat(par.audiopar.SampleFormat)
		codecctx.SetSampleRate(par.audiopar.Samplerate)

		// Set stream parameters
		encoder.stream.SetTimeBase(ff.AVUtil_rational(1, par.audiopar.Samplerate))
	case ff.AVMEDIA_TYPE_VIDEO:
		encoder.t = VIDEO
		// TODO: Check codec supports this configuration

		// Set codec parameters
		codecctx.SetPixFmt(par.videopar.PixelFormat)
		codecctx.SetWidth(par.videopar.Width)
		codecctx.SetHeight(par.videopar.Height)

		// Set stream parameters
		encoder.stream.SetTimeBase(ff.AVUtil_rational_d2q(1/par.codecpar.Framerate, 1<<24))
	case ff.AVMEDIA_TYPE_SUBTITLE:
		encoder.t = SUBTITLE
		fmt.Println("TODO: Set encoding subtitle parameters")
	default:
		encoder.t = DATA
	}
	encoder.t |= OUTPUT

	// copy parameters to the stream
	if err := ff.AVCodec_parameters_from_context(encoder.stream.CodecPar(), codecctx); err != nil {
		ff.AVCodec_free_context(codecctx)
		return nil, err
	}

	// Some formats want stream headers to be separate.
	if ctx.Flags().Is(ff.AVFMT_GLOBALHEADER) {
		codecctx.SetFlags(codecctx.Flags() | ff.AV_CODEC_FLAG_GLOBAL_HEADER)
	}

	// Open it
	if err := ff.AVCodec_open(codecctx, codec, nil); err != nil {
		ff.AVCodec_free_context(codecctx)
		return nil, ErrInternalAppError.Withf("codec_open: %v", err)
	}

	// Allocate packet
	if packet := ff.AVCodec_packet_alloc(); packet == nil {
		ff.AVCodec_free_context(codecctx)
		return nil, ErrInternalAppError.With("could not allocate packet")
	} else {
		encoder.packet = packet
	}

	// Return it
	return encoder, nil
}

func (encoder *encoder) Close() error {
	// Free respurces
	if encoder.packet != nil {
		ff.AVCodec_packet_free(encoder.packet)
	}
	if encoder.ctx != nil {
		ff.AVCodec_free_context(encoder.ctx)
	}

	// Release resources
	encoder.stream = nil
	encoder.packet = nil
	encoder.ctx = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (encoder *encoder) encode(fn MuxFunc) (*ff.AVPacket, error) {
	packet, err := fn(encoder.stream.Id())
	if packet != nil {
	}

}
