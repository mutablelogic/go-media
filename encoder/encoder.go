package encoder

import (
	"errors"

	// Packages

	"github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Encoder struct {
	ctx *ff.AVCodecContext
}

type Profile interface {
	Type() ff.AVMediaType
	Codec() *ff.AVCodec
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create an encoder with the given parameters
func New(ctx *ff.AVFormatContext, stream int, codec Profile) (*Encoder, error) {
	self := new(Encoder)

	codecctx := ff.AVCodec_alloc_context(codec.Codec())
	if codecctx == nil {
		return nil, errors.New("could not allocate codec context")
	} else {
		self.ctx = codecctx
	}

	// Create the stream
	streamctx := ff.AVFormat_new_stream(ctx, codec)
	if streamctx == nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, errors.New("could not allocate stream")
	}
	// Set stream identifier and timebase from parameters
	streamctx.SetId(stream)
	encoder.stream = streamctx

	// Some formats want stream headers to be separate.
	if ctx.Output().Flags().Is(ff.AVFMT_GLOBALHEADER) {
		encoder.ctx.SetFlags(encoder.ctx.Flags() | ff.AV_CODEC_FLAG_GLOBAL_HEADER)
	}

	// Get the options
	opts := par.optionsToDict()
	if opts == nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, errors.New("could not allocate options dictionary")
	}
	defer ff.AVUtil_dict_free(opts)

	// Open it
	if err := ff.AVCodec_open(encoder.ctx, codec, opts); err != nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, err
	}

	// If there are any non-consumed options, then error
	var result error
	for _, key := range ff.AVUtil_dict_keys(opts) {
		result = errors.Join(result, media.ErrBadParameter.With("invalid codec option: "+key))
	}
	if result != nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, result
	}

	// Copy parameters to stream
	if err := ff.AVCodec_parameters_from_context(encoder.stream.CodecPar(), encoder.ctx); err != nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, err
	}

	// Hint what timebase we want to encode at. This will change when writing the
	// headers for the encoding process
	tb := par.timebase
	if tb.Num() == 0 || tb.Den() == 0 {
		// For audio, Par.timebase may be unset; fall back to codec context timebase.
		tb = encoder.ctx.TimeBase()
	}
	encoder.stream.SetTimeBase(tb)

	// Return success
	return encoder, nil
}

// Close the encoder and release resources
func (e *encoder) Close() error {
	// NOTE: Do NOT free e.ctx here! The AVCodecContext is owned by the AVStream,
	// and will be freed automatically when avformat_free_context() is called.
	// Freeing it here causes a double-free crash.
	//
	// The encoder context is stored in the stream's internal FFStream->avctx field,
	// and ff_free_stream() (called by avformat_free_context) will call
	// avcodec_free_context() on it.

	// Just nil out our references
	e.ctx = nil
	e.stream = nil

	// Return success
	return nil
}
