package ffmpeg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"syscall"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Encoder struct {
	ctx    *ff.AVCodecContext
	stream *ff.AVStream
	packet *ff.AVPacket
	//next_pts int64
}

// EncoderFrameFn is a function which is called to receive a frame to encode. It should
// return nil to continue encoding or io.EOF to stop encoding.
type EncoderFrameFn func(int) (*ff.AVFrame, error)

// EncoderPacketFn is a function which is called for each packet encoded. It should
// return nil to continue encoding or io.EOF to stop encoding immediately.
type EncoderPacketFn func(*ff.AVPacket) error

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
	if err := par.ValidateFromCodec(encoder.ctx); err != nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, err
	} else if err := par.CopyToCodec(encoder.ctx); err != nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, err
	}

	// Create the stream
	if streamctx := ff.AVFormat_new_stream(ctx, nil); streamctx == nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, ErrInternalAppError.With("could not allocate stream")
	} else {
		streamctx.SetId(stream)
		encoder.stream = streamctx
	}

	// Copy parameters to stream
	if err := ff.AVCodec_parameters_from_context(encoder.stream.CodecPar(), encoder.ctx); err != nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, err
	}

	// Some formats want stream headers to be separate.
	if ctx.Output().Flags().Is(ff.AVFMT_GLOBALHEADER) {
		encoder.ctx.SetFlags(encoder.ctx.Flags() | ff.AV_CODEC_FLAG_GLOBAL_HEADER)
	}

	// Open it
	if err := ff.AVCodec_open(encoder.ctx, codec, nil); err != nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, ErrInternalAppError.Withf("codec_open: %v", err)
	}

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
func (e *Encoder) Encode(frame *ff.AVFrame, fn EncoderPacketFn) error {
	if fn == nil {
		return ErrBadParameter.With("nil fn")
	}
	// Encode a frame (or flush the encoder)
	return e.encode(frame, fn)
}

// Return the codec parameters
func (e *Encoder) Par() *Par {
	par := new(Par)
	if err := ff.AVCodec_parameters_from_context(&par.AVCodecParameters, e.ctx); err != nil {
		return nil
	} else {
		return par
	}
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (e *Encoder) encode(frame *ff.AVFrame, fn EncoderPacketFn) error {
	// Send the frame to the encoder
	fmt.Println("Sending frame", frame)
	if err := ff.AVCodec_send_frame(e.ctx, frame); err != nil {
		if errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	// Write out the packets
	var result error
	for {
		// Receive the packet
		fmt.Println("Receiving packet")
		if err := ff.AVCodec_receive_packet(e.ctx, e.packet); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
			// Finished receiving packet or EOF
			break
		} else if err != nil {
			return err
		}

		// Pass back to the caller
		if err := fn(e.packet); errors.Is(err, io.EOF) {
			// End early, return EOF
			result = io.EOF
			break
		} else if err != nil {
			return err
		}

		// Re-allocate frames for next iteration
		ff.AVCodec_packet_unref(e.packet)
	}

	// Flush
	if result == nil {
		result = fn(nil)
	}

	// Return success or EOF
	return result
}
