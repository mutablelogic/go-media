package ffmpeg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"syscall"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type encoder struct {
	ctx    *ff.AVCodecContext
	stream *ff.AVStream
	// packet *ff.AVPacket // Removed: allocate per frame to avoid race conditions
	eof bool // We are flushing the encoder
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create an encoder with the given parameters
func newEncoder(ctx *ff.AVFormatContext, stream int, par *Par) (*encoder, error) {
	encoder := new(encoder)

	// Get codec
	codecID := ff.AV_CODEC_ID_NONE
	switch par.CodecType() {
	case ff.AVMEDIA_TYPE_AUDIO:
		codecID = ctx.Output().AudioCodec()
	case ff.AVMEDIA_TYPE_VIDEO:
		codecID = ctx.Output().VideoCodec()
	case ff.AVMEDIA_TYPE_SUBTITLE:
		codecID = ctx.Output().SubtitleCodec()
	}
	if codecID == ff.AV_CODEC_ID_NONE {
		return nil, media.ErrBadParameter.With("no codec specified for stream")
	}

	// Allocate codec
	codec := ff.AVCodec_find_encoder(codecID)
	if codec == nil {
		return nil, media.ErrBadParameter.With("codec cannot encode")
	}
	codecctx := ff.AVCodec_alloc_context(codec)
	if codecctx == nil {
		return nil, errors.New("could not allocate codec context")
	}
	encoder.ctx = codecctx

	// Check codec against parameters and set defaults as needed, then
	// copy back to codec
	if err := par.ValidateFromCodec(encoder.ctx.Codec()); err != nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, err
	}
	if err := par.CopyToCodecContext(encoder.ctx); err != nil {
		ff.AVCodec_free_context(encoder.ctx)
		return nil, err
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

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e *encoder) MarshalJSON() ([]byte, error) {
	type jsonEncoder struct {
		Codec  *ff.AVCodecContext `json:"codec"`
		Stream *ff.AVStream       `json:"stream"`
	}
	return json.Marshal(&jsonEncoder{
		Codec:  e.ctx,
		Stream: e.stream,
	})
}

func (e *encoder) String() string {
	data, _ := json.MarshalIndent(e, "", "  ")
	return string(data)
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Encode a frame and pass packets to the EncoderPacketFn. If the frame is nil, then
// the encoder will flush any remaining packets. If io.EOF is returned then
// it indicates that the encoder has ended prematurely.
func (e *encoder) Encode(frame *Frame, fn EncoderPacketFn) error {
	if fn == nil {
		return media.ErrBadParameter.With("nil callback function")
	}
	if e.ctx == nil {
		return errors.New("encoder is closed")
	}
	return e.encode(frame, fn)
}

// EncodeSubtitle encodes a subtitle and passes the packet to the EncoderPacketFn.
// Subtitles use a legacy encoding API and don't support flushing (nil subtitle will return an error).
func (e *encoder) EncodeSubtitle(sub *ff.AVSubtitle, fn EncoderPacketFn) error {
	if fn == nil {
		return media.ErrBadParameter.With("nil callback function")
	}
	if e.ctx == nil {
		return errors.New("encoder is closed")
	}
	if sub == nil {
		return media.ErrBadParameter.With("subtitle cannot be nil (no flushing support)")
	}
	if e.ctx.CodecType() != ff.AVMEDIA_TYPE_SUBTITLE {
		return media.ErrBadParameter.With("encoder is not configured for subtitles")
	}
	return e.encodeSubtitle(sub, fn)
}

// Return the codec parameters
func (e *encoder) Par() *Par {
	if e.ctx == nil {
		return nil
	}
	par := new(Par)
	par.timebase = e.ctx.TimeBase()
	if err := ff.AVCodec_parameters_from_context(&par.AVCodecParameters, e.ctx); err != nil {
		return nil
	}
	return par
}

// Return the stream index
func (e *encoder) Stream() int {
	if e.stream == nil {
		return -1
	}
	return e.stream.Index()
}

// FrameSize returns the number of samples per frame (audio only, 0 for video or variable frame size)
func (e *encoder) FrameSize() int {
	if e.ctx == nil {
		return 0
	}
	return e.ctx.FrameSize()
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (e *encoder) encode(frame *Frame, fn EncoderPacketFn) error {
	// Send the frame to the encoder
	if err := ff.AVCodec_send_frame(e.ctx, (*ff.AVFrame)(frame)); err != nil {
		return err
	}

	// Write out the packets
	var result error
	for {
		// Allocate a new packet for each iteration to avoid race conditions
		// if the callback queues the packet pointer (e.g. async muxing)
		packet := ff.AVCodec_packet_alloc()
		if packet == nil {
			return errors.New("failed to allocate packet")
		}

		// Receive the packet
		if err := ff.AVCodec_receive_packet(e.ctx, packet); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
			// Finished receiving packets or EOF
			ff.AVCodec_packet_free(packet)
			break
		} else if err != nil {
			ff.AVCodec_packet_free(packet)
			return err
		}

		// Rescale output packet timestamp values from codec to stream timebase
		ff.AVCodec_packet_rescale_ts(packet, e.ctx.TimeBase(), e.stream.TimeBase())

		// Calculate proper duration in stream timebase
		// Duration should be 1 frame in codec timebase, rescaled to stream timebase
		if packet.Duration() <= 0 {
			codecTb := e.ctx.TimeBase()
			streamTb := e.stream.TimeBase()
			// dur = 1 frame * (stream_den / stream_num) / (codec_den / codec_num)
			// = (stream_den * codec_num) / (stream_num * codec_den)
			dur := int64(streamTb.Den()) * int64(codecTb.Num()) / (int64(codecTb.Den()) * int64(streamTb.Num()))
			if dur <= 0 {
				dur = 1
			}
			packet.SetDuration(dur)
		}

		// Set packet parameters
		packet.SetStreamIndex(e.stream.Index())

		// DEBUG: Before muxer
		fmt.Printf("BEFORE MUXER: stream=%d pts=%d dts=%d duration=%d size=%d\n",
			packet.StreamIndex(), packet.Pts(), packet.Dts(), packet.Duration(), packet.Size())

		// Pass back to the caller
		err := fn((*Packet)(packet))

		// DEBUG: After muxer
		fmt.Printf("AFTER MUXER: stream=%d pts=%d dts=%d duration=%d size=%d\n",
			packet.StreamIndex(), packet.Pts(), packet.Dts(), packet.Duration(), packet.Size())

		// After av_interleaved_write_frame returns, the packet data has been
		// consumed (unreferenced). We can now safely free the packet structure.
		// av_packet_free internally calls av_packet_unref which is a no-op
		// if the packet was already unreferenced, then frees the struct.
		ff.AVCodec_packet_free(packet)

		if errors.Is(err, io.EOF) {
			// End early, return EOF
			result = io.EOF
			break
		} else if err != nil {
			return err
		}
	}

	// Flush packet (send nil to indicate end of packet batch)
	if result == nil {
		result = fn(nil)
	}

	// Return success or EOF
	return result
}

func (e *encoder) encodeSubtitle(sub *ff.AVSubtitle, fn EncoderPacketFn) error {
	// Allocate buffer for subtitle data (subtitles are typically small, 64KB should be sufficient)
	buf := make([]byte, 65536)

	// Encode subtitle using legacy API
	bytesWritten, err := ff.AVCodec_encode_subtitle(e.ctx, buf, sub)
	if err != nil {
		return err
	}
	if bytesWritten == 0 {
		// No data encoded, nothing to send
		return nil
	}

	// Allocate packet to hold the subtitle data
	packet := ff.AVCodec_packet_alloc()
	if packet == nil {
		return errors.New("failed to allocate packet")
	}

	// Copy subtitle data to packet
	if err := ff.AVCodec_packet_from_data(packet, buf[:bytesWritten]); err != nil {
		ff.AVCodec_packet_free(packet)
		return err
	}

	// Set packet parameters from subtitle
	packet.SetStreamIndex(e.stream.Index())
	packet.SetPts(sub.PTS())
	packet.SetDts(sub.PTS())

	// Calculate duration from subtitle display times (convert ms to stream timebase)
	startMs := int64(sub.StartDisplayTime())
	endMs := int64(sub.EndDisplayTime())
	durationMs := endMs - startMs
	if durationMs > 0 {
		// Convert ms to stream timebase units: duration = (durationMs * timebase.den) / (1000 * timebase.num)
		streamTb := e.stream.TimeBase()
		duration := (durationMs * int64(streamTb.Den())) / (1000 * int64(streamTb.Num()))
		if duration <= 0 {
			duration = 1
		}
		packet.SetDuration(duration)
	}

	// Set timebase for the packet
	packet.SetTimeBase(e.stream.TimeBase())

	// DEBUG: Before callback
	fmt.Printf("SUBTITLE ENCODE: stream=%d pts=%d duration=%d size=%d\n",
		packet.StreamIndex(), packet.Pts(), packet.Duration(), packet.Size())

	// Pass to callback
	err = fn((*Packet)(packet))

	// Free the packet
	ff.AVCodec_packet_free(packet)

	if errors.Is(err, io.EOF) {
		return io.EOF
	}

	// Flush packet (send nil to indicate end)
	if err == nil {
		err = fn(nil)
	}

	return err
}
