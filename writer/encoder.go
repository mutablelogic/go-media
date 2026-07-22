package writer

import (
	"errors"
	"io"
	"sync"
	"syscall"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	frame "github.com/mutablelogic/go-media/frame"
	profile "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// PacketFn is called for each packet an Encoder produces. Returning io.EOF
// stops encoding early without being treated as an error.
type PacketFn func(*ff.AVPacket) error

// Encoder is a standalone, muxer-agnostic collection of codec encoders keyed
// by caller-chosen stream ID. Unlike Writer, it never creates an AVStream or
// AVFormatContext, so each codec context here has no muxer-owned AVStream to
// free it — Encoder itself owns every context it opens and must free them in
// Close().
type Encoder struct {
	sync.RWMutex
	codecs map[int]*ff.AVCodecContext
	fn     PacketFn
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewEncoder creates an empty Encoder that passes every packet it produces
// to fn. Use Add to register a codec for each stream ID before calling
// Encode.
func NewEncoder(fn PacketFn) (*Encoder, error) {
	if fn == nil {
		return nil, gomedia.ErrBadParameter.With("nil callback function")
	}
	return &Encoder{codecs: make(map[int]*ff.AVCodecContext), fn: fn}, nil
}

// Add opens a codec context for streamID from the given profile. Codec-
// specific options come from profile.Options(), applied the same way
// Writer applies Output.Opts to the format context.
func (e *Encoder) Add(streamID int, p profile.Profile) error {
	e.Lock()
	defer e.Unlock()

	if p == nil {
		return gomedia.ErrBadParameter.With("nil profile")
	}

	codec := p.Codec().Context()
	if codec == nil {
		return gomedia.ErrBadParameter.With("profile has no resolved codec")
	}

	if _, exists := e.codecs[streamID]; exists {
		return gomedia.ErrBadParameter.Withf("stream %d already has a codec", streamID)
	}

	ctx := ff.AVCodec_alloc_context(codec)
	if ctx == nil {
		return gomedia.ErrInternalError.With("failed to allocate codec context")
	}

	// Copy codec parameters (bitrate, sample rate/format, channel layout, ...)
	if err := ff.AVCodec_parameters_to_context(ctx, p.Par()); err != nil {
		ff.AVCodec_free_context(ctx)
		return err
	}

	// Set timebase if specified
	if timebase := p.TimeBase(); timebase != nil {
		ctx.SetTimeBase(*timebase)
	}

	// Build the codec-specific options dictionary from the profile
	dict, err := dictFromOpts(p.Options())
	if err != nil {
		ff.AVCodec_free_context(ctx)
		return err
	}
	defer ff.AVUtil_dict_free(dict)

	// Open the codec, consuming recognized options from the dictionary
	if err := ff.AVCodec_open(ctx, codec, dict); err != nil {
		ff.AVCodec_free_context(ctx)
		return err
	}

	// Any keys left in the dictionary were not recognized by the codec
	if keys := ff.AVUtil_dict_keys(dict); len(keys) > 0 {
		ff.AVCodec_free_context(ctx)
		return gomedia.ErrBadParameter.Withf("invalid codec options for stream %d: %v", streamID, keys)
	}

	e.codecs[streamID] = ctx
	return nil
}

// Close releases every codec context this Encoder owns.
func (e *Encoder) Close() error {
	e.Lock()
	defer e.Unlock()

	for _, ctx := range e.codecs {
		ff.AVCodec_free_context(ctx)
	}
	e.codecs = nil
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Encode sends f to the codec registered for f.StreamID and passes
// resulting packets to the Encoder's callback. Packets carry the codec's own
// timebase — rescale to a muxer's stream timebase downstream if needed.
func (e *Encoder) Encode(f *frame.Frame) error {
	if f == nil {
		return gomedia.ErrBadParameter.With("nil frame")
	}
	ctx, err := e.contextFor(f.StreamID)
	if err != nil {
		return err
	} else {
		return e.encode(ctx, f.StreamID, f.AVFrame)
	}
}

// Flush signals end-of-stream to the codec registered for streamID and
// passes any remaining buffered packets to the Encoder's callback.
func (e *Encoder) Flush(streamID int) error {
	ctx, err := e.contextFor(streamID)
	if err != nil {
		return err
	} else {
		return e.encode(ctx, streamID, nil)
	}
}

// FrameSize returns the number of samples per frame the codec registered for
// streamID expects (audio only; 0 if the stream doesn't exist or the codec
// accepts variable frame sizes).
func (e *Encoder) FrameSize(streamID int) int {
	ctx, err := e.contextFor(streamID)
	if err != nil {
		return 0
	} else {
		return ctx.FrameSize()
	}
}

// Par returns codec parameters reflecting the opened codec context for
// streamID — unlike the profile originally passed to Add, this includes
// whatever the codec generated on open (e.g. OpusHead for libopus,
// AudioSpecificConfig for aac), which a muxer's AVStream needs. Caller must
// free the result with ff.AVCodec_parameters_free.
func (e *Encoder) Par(streamID int) (*ff.AVCodecParameters, error) {
	ctx, err := e.contextFor(streamID)
	if err != nil {
		return nil, err
	}

	par := ff.AVCodec_parameters_alloc()
	if par == nil {
		return nil, gomedia.ErrInternalError.With("failed to allocate codec parameters")
	}
	if err := ff.AVCodec_parameters_from_context(par, ctx); err != nil {
		ff.AVCodec_parameters_free(par)
		return nil, err
	}
	return par, nil
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (e *Encoder) contextFor(streamID int) (*ff.AVCodecContext, error) {
	e.RLock()
	defer e.RUnlock()

	ctx, exists := e.codecs[streamID]
	if !exists {
		return nil, gomedia.ErrBadParameter.Withf("no codec registered for stream %d", streamID)
	} else {
		return ctx, nil
	}
}

func (e *Encoder) encode(ctx *ff.AVCodecContext, streamID int, frame *ff.AVFrame) error {
	// Send the frame to the encoder (nil frame flushes)
	if err := ff.AVCodec_send_frame(ctx, frame); err != nil {
		return err
	}

	// Write out the packets
	var result error
	for {
		// Allocate a new packet for each iteration to avoid race conditions
		// if the callback queues the packet pointer
		packet := ff.AVCodec_packet_alloc()
		if packet == nil {
			return errors.New("failed to allocate packet")
		}

		if err := ff.AVCodec_receive_packet(ctx, packet); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
			ff.AVCodec_packet_free(packet)
			break
		} else if err != nil {
			ff.AVCodec_packet_free(packet)
			return err
		}

		packet.SetStreamIndex(streamID)

		err := e.fn(packet)
		ff.AVCodec_packet_free(packet)

		if errors.Is(err, io.EOF) {
			result = io.EOF
			break
		} else if err != nil {
			return err
		}
	}

	// Signal end of packet batch
	if result == nil {
		result = e.fn(nil)
	}

	return result
}
