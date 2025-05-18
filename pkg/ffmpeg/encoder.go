package ffmpeg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"syscall"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Encoder struct {
	ctx    *ff.AVCodecContext
	stream *ff.AVStream
	packet *ff.AVPacket

	// We are flushing the encoder
	eof bool

	// The next presentation timestamp
	next_pts int64
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create an encoder with the given parameters
func NewEncoder(ctx *ff.AVFormatContext, stream int, par *Par) (*Encoder, error) {
	encoder := new(Encoder)

	// Get codec
	codec_id := ff.AV_CODEC_ID_NONE
	switch par.CodecType() {
	case ff.AVMEDIA_TYPE_AUDIO:
		codec_id = ctx.Output().AudioCodec()
	case ff.AVMEDIA_TYPE_VIDEO:
		codec_id = ctx.Output().VideoCodec()
	case ff.AVMEDIA_TYPE_SUBTITLE:
		codec_id = ctx.Output().SubtitleCodec()
	}
	if codec_id == ff.AV_CODEC_ID_NONE {
		return nil, ErrBadParameter.Withf("no codec specified for stream %v", stream)
	}

	// Allocate codec
	codec := ff.AVCodec_find_encoder(codec_id)
	if codec == nil {
		return nil, ErrBadParameter.Withf("codec %q cannot encode", codec_id)
	}
	if codecctx := ff.AVCodec_alloc_context(codec); codecctx == nil {
		return nil, ErrInternalAppError.With("could not allocate audio codec context")
	} else {
		encoder.ctx = codecctx
	}

	// Check codec against parameters and set defaults as needed, then
	// copy back to codec
	if err := par.ValidateFromCodec(encoder.ctx.Codec()); err != nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, err
	} else if err := par.CopyToCodecContext(encoder.ctx); err != nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, err
	}

	// Create the stream
	if streamctx := ff.AVFormat_new_stream(ctx, codec); streamctx == nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, ErrInternalAppError.With("could not allocate stream")
	} else {
		// Set stream identifier and timebase from parameters
		streamctx.SetId(stream)
		encoder.stream = streamctx
	}

	// Some formats want stream headers to be separate.
	if ctx.Output().Flags().Is(ff.AVFMT_GLOBALHEADER) {
		encoder.ctx.SetFlags(encoder.ctx.Flags() | ff.AV_CODEC_FLAG_GLOBAL_HEADER)
	}

	// Get the options
	opts := par.newOpts()
	if opts == nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, ErrInternalAppError.With("could not allocate options dictionary")
	}
	defer ff.AVUtil_dict_free(opts)

	// Open it
	if err := ff.AVCodec_open(encoder.ctx, codec, opts); err != nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, ErrInternalAppError.Withf("codec_open: %v", err)
	}

	// If there are any non-consumed options, then error
	var result error
	for _, key := range ff.AVUtil_dict_keys(opts) {
		result = errors.Join(result, ErrBadParameter.Withf("Stream %d: invalid codec option %q", stream, key))
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
	encoder.stream.SetTimeBase(par.timebase)

	// Create a packet
	packet := ff.AVCodec_packet_alloc()
	if packet == nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, errors.New("failed to allocate packet")
	} else {
		encoder.packet = packet
	}

	// Return it
	return encoder, nil
}

func (encoder *Encoder) Close() error {
	// Free respurces
	if encoder.ctx != nil {
		ff.AVCodec_free_context(encoder.ctx)
	}
	if encoder.packet != nil {
		ff.AVCodec_packet_free(encoder.packet)
	}

	// Release resources
	encoder.packet = nil
	encoder.stream = nil
	encoder.ctx = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e *Encoder) MarshalJSON() ([]byte, error) {
	type jsonEncoder struct {
		Codec  *ff.AVCodecContext `json:"codec"`
		Stream *ff.AVStream       `json:"stream"`
	}
	return json.Marshal(&jsonEncoder{
		Codec:  e.ctx,
		Stream: e.stream,
	})
}

func (e *Encoder) String() string {
	data, _ := json.MarshalIndent(e, "", "  ")
	return string(data)
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Encode a frame and pass packets to the EncoderPacketFn. If the frame is nil, then
// the encoder will flush any remaining packets. If io.EOF is returned then
// it indicates that the encoder has ended prematurely.
func (e *Encoder) Encode(frame *Frame, fn EncoderPacketFn) error {
	if fn == nil {
		return ErrBadParameter.With("nil fn")
	}
	// Encode a frame (or flush the encoder)
	return e.encode(frame, fn)
}

// Return the codec parameters
func (e *Encoder) Par() *Par {
	par := new(Par)
	par.timebase = e.ctx.TimeBase()
	if err := ff.AVCodec_parameters_from_context(&par.AVCodecParameters, e.ctx); err != nil {
		return nil
	} else {
		return par
	}
}

// Return the next expected timestamp after a frame has been encoded
func (e *Encoder) nextPts(frame *Frame) int64 {
	next_pts := int64(0)
	switch e.ctx.Codec().Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		next_pts = ff.AVUtil_rational_rescale_q(int64(frame.NumSamples()), frame.TimeBase(), e.stream.TimeBase())
	case ff.AVMEDIA_TYPE_VIDEO:
		next_pts = ff.AVUtil_rational_rescale_q(1, frame.TimeBase(), e.stream.TimeBase())
	default:
		// Dunno what to do with subtitle and data streams yet
		fmt.Println("TODO: next_pts for subtitle and data streams")
		return 0
	}
	return next_pts
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (e *Encoder) encode(frame *Frame, fn EncoderPacketFn) error {
	// Send the frame to the encoder
	if err := ff.AVCodec_send_frame(e.ctx, (*ff.AVFrame)(frame)); err != nil {
		return err
	}

	// Write out the packets
	var result error
	for {
		// Receive the packet
		if err := ff.AVCodec_receive_packet(e.ctx, e.packet); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
			// Finished receiving packet or EOF
			break
		} else if err != nil {
			return err
		}

		// rescale output packet timestamp values from codec to stream timebase
		ff.AVCodec_packet_rescale_ts(e.packet, e.ctx.TimeBase(), e.stream.TimeBase())

		// Set packet parameters
		e.packet.SetStreamIndex(e.stream.Index())
		e.packet.SetTimeBase(e.stream.TimeBase())

		// Pass back to the caller
		if err := fn((*Packet)(e.packet)); errors.Is(err, io.EOF) {
			// End early, return EOF
			result = io.EOF
			break
		} else if err != nil {
			return err
		}

		// Re-allocate frames for next iteration
		ff.AVCodec_packet_unref(e.packet)
	}

	// Flush packet
	if result == nil {
		result = fn(nil)
	}

	// Return success or EOF
	return result
}
