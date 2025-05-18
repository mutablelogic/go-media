package ffmpeg

import (
	"context"
	"errors"
	"io"
	"syscall"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Decoding context
type Context struct {
	input    *ff.AVFormatContext
	decoders map[int]*Decoder
	ch       map[int]chan *Frame
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new decoding context
func newContext(r *Reader, fn DecoderMapFunc) (*Context, error) {
	ctx := new(Context)
	ctx.input = r.input
	ctx.decoders = make(map[int]*Decoder, r.input.NumStreams())
	ctx.ch = make(map[int]chan *Frame, r.input.NumStreams())

	// Do stream mapping
	if err := ctx.mapStreams(fn, r.force); err != nil {
		return nil, errors.Join(err, ctx.Close())
	}

	// Make channels for each decoder
	for stream_index := range ctx.decoders {
		ctx.ch[stream_index] = make(chan *Frame)
	}

	// Return sucess
	return ctx, nil
}

// Release resources for the decoding context
func (c *Context) Close() error {
	var result error
	for _, decoder := range c.decoders {
		if err := decoder.Close(); err != nil {
			result = errors.Join(result, err)
		}
	}
	for _, ch := range c.ch {
		close(ch)
	}

	// Release resources
	c.decoders = nil
	c.ch = nil
	c.input = nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (decoder *Context) C(stream int) chan *Frame {
	return decoder.ch[stream]
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (decoder *Context) decode(ctx context.Context, fn DecoderFrameFn) error {
	// Allocate a packet
	packet := ff.AVCodec_packet_alloc()
	if packet == nil {
		return errors.New("failed to allocate packet")
	}
	defer ff.AVCodec_packet_free(packet)

	// Read packets
FOR_LOOP:
	for {
		select {
		case <-ctx.Done():
			break FOR_LOOP
		default:
			if err := ff.AVFormat_read_frame(decoder.input, packet); errors.Is(err, io.EOF) {
				break FOR_LOOP
			} else if errors.Is(err, syscall.EAGAIN) {
				continue FOR_LOOP
			} else if err != nil {
				return ErrInternalAppError.With("AVFormat_read_frame: ", err)
			}
			stream_index := packet.StreamIndex()
			if decoder := decoder.decoders[stream_index]; decoder != nil {
				if err := decoder.decode(packet, fn); errors.Is(err, io.EOF) {
					break FOR_LOOP
				} else if err != nil {
					return err
				}
			}
		}

		// Unreference the packet
		ff.AVCodec_packet_unref(packet)
	}

	// Flush the decoders
	for _, decoder := range decoder.decoders {
		if err := decoder.decode(nil, fn); errors.Is(err, io.EOF) {
			// no-op
		} else if err != nil {
			return err
		}
	}

	// Return the context error - will be cancelled, perhaps, or nil if the
	// demuxer finished successfully without cancellation
	return ctx.Err()
}

// Map streams to decoders, and return the decoders
func (c *Context) mapStreams(fn DecoderMapFunc, force bool) error {
	// Standard decoder map function copies all streams
	if fn == nil {
		fn = func(_ int, par *Par) (*Par, error) {
			return par, nil
		}
	}

	// Create a decoder for each stream. The decoder map function
	// should be returning the parameters for the destination frame.
	var result error
	for _, stream := range c.input.Streams() {
		stream_index := stream.Index()

		// Get decoder parameters and map to a decoder
		par, err := fn(stream_index, &Par{
			AVCodecParameters: *stream.CodecPar(),
			timebase:          stream.TimeBase(),
		})
		if err != nil {
			result = errors.Join(result, err)
		} else if par == nil {
			continue
		} else if decoder, err := NewDecoder(stream, par, force); err != nil {
			result = errors.Join(result, err)
		} else if _, exists := c.decoders[stream_index]; exists {
			result = errors.Join(result, ErrDuplicateEntry.Withf("stream index %d", stream_index))
		} else {
			c.decoders[stream_index] = decoder
		}
	}

	// Check to see if we have to do something
	if len(c.decoders) == 0 {
		result = errors.Join(result, ErrBadParameter.With("no streams to decode"))
	}

	// Return any errors
	return result
}
