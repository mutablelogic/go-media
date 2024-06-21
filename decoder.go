package media

import (
	"context"
	"errors"
	"fmt"
	"io"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

// demuxer context - deconstructs media into packets
type demuxer struct {
	input    *ff.AVFormatContext
	decoders map[int]*decoder
	frame    *ff.AVFrame // Source frame
}

// decoder context - decodes packets into frames
type decoder struct {
	stream int
	codec  *ff.AVCodecContext
	frame  *ff.AVFrame // Destination frame
}

var _ Decoder = (*demuxer)(nil)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newDemuxer(input *ff.AVFormatContext, mapfn DecoderMapFunc) (*demuxer, error) {
	demuxer := new(demuxer)
	demuxer.input = input
	demuxer.decoders = make(map[int]*decoder)

	// Get all the streams
	streams := input.Streams()

	// Use standard map function if none provided
	if mapfn == nil {
		mapfn = func(stream Stream) (Parameters, error) {
			return stream.Parameters(), nil
		}
	}

	// Create a decoder for each stream
	// The decoder map function should be returning the parameters for the
	// destination frame. If it's nil then it's mostly a copy.
	var result error
	for _, stream := range streams {
		// Get decoder parameters
		parameters, err := mapfn(newStream(stream))
		if err != nil {
			result = errors.Join(result, err)
		} else if parameters == nil {
			continue
		}

		// Create the decoder with the parameters
		if decoder, err := demuxer.newDecoder(stream, parameters); err != nil {
			result = errors.Join(result, err)
		} else {
			streamNum := stream.Index()
			demuxer.decoders[streamNum] = decoder
		}
	}

	// Return any errors
	if result != nil {
		return nil, errors.Join(result, demuxer.close())
	}

	// Create a frame for encoding - after resampling and resizing
	if frame := ff.AVUtil_frame_alloc(); frame == nil {
		return nil, errors.Join(demuxer.close(), errors.New("failed to allocate frame"))
	} else {
		demuxer.frame = frame
	}

	// Return success
	return demuxer, nil
}

func (d *demuxer) newDecoder(stream *ff.AVStream, parameters Parameters) (*decoder, error) {
	decoder := new(decoder)
	decoder.stream = stream.Id()

	// TODO: Use parameters to create the decoder

	// Create a codec context for the decoder
	codec := ff.AVCodec_find_decoder(stream.CodecPar().CodecID())
	if codec == nil {
		return nil, fmt.Errorf("failed to find decoder for codec %q", stream.CodecPar().CodecID())
	} else if ctx := ff.AVCodec_alloc_context(codec); ctx == nil {
		return nil, fmt.Errorf("failed to allocate codec context for codec %q", codec.Name())
	} else {
		decoder.codec = ctx
	}

	// Copy codec parameters from input stream to output codec context
	if err := ff.AVCodec_parameters_to_context(decoder.codec, stream.CodecPar()); err != nil {
		return nil, errors.Join(decoder.close(), fmt.Errorf("failed to copy codec parameters to decoder context for codec %q", codec.Name()))
	}

	// Init the decoder
	if err := ff.AVCodec_open(decoder.codec, codec, nil); err != nil {
		return nil, errors.Join(decoder.close(), err)
	}

	// Create a frame for decoder output - after resize/resample
	if frame := ff.AVUtil_frame_alloc(); frame == nil {
		return nil, errors.Join(decoder.close(), errors.New("failed to allocate frame"))
	} else {
		decoder.frame = frame
	}

	// Return success
	return decoder, nil
}

func (d *demuxer) close() error {
	var result error

	// Free decoded frame
	if d.frame != nil {
		ff.AVUtil_frame_free(d.frame)
	}

	// Free resources
	for _, decoder := range d.decoders {
		result = errors.Join(result, decoder.close())
	}
	d.decoders = nil

	// Return any errors
	return result
}

func (d *decoder) close() error {
	var result error

	// Free the codec context
	if d.codec != nil {
		ff.AVCodec_free_context(d.codec)
	}

	// Free destination frame
	if d.frame != nil {
		ff.AVUtil_frame_free(d.frame)
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (d *demuxer) Demux(ctx context.Context, fn DecoderFunc) error {
	// If the decoder is nil then set it to default - which is to send the
	// packet to the appropriate decoder
	if fn == nil {
		fn = func(packet Packet) error {
			fmt.Println("TODO", packet)
			return nil
		}
	}

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
			if err := ff.AVFormat_read_frame(d.input, packet); errors.Is(err, io.EOF) {
				break FOR_LOOP
			} else if err != nil {
				return err
			}
			stream := packet.StreamIndex()
			if decoder := d.decoders[stream]; decoder != nil {
				if err := decoder.decode(fn, packet); errors.Is(err, io.EOF) {
					break FOR_LOOP
				} else if err != nil {
					return err
				}
			}
			// Unreference the packet
			ff.AVCodec_packet_unref(packet)
		}
	}

	// Flush the decoders
	for _, decoder := range d.decoders {
		if err := decoder.decode(fn, nil); err != nil {
			return err
		}
	}

	// Return the context error - will be cancelled, perhaps, or nil if the
	// demuxer finished successfully without cancellation
	return ctx.Err()
}

func (d *demuxer) Decode(context.Context, FrameFunc) error {
	// TODO
	return errors.New("not implemented")
}

////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (d *decoder) decode(fn DecoderFunc, packet *ff.AVPacket) error {
	// Send the packet to the user defined packet function or
	// to the default version
	return fn(packet)
}

/*
	// Get the codec
	stream.

		// Find the decoder for the stream
		dec := ff.AVCodec_find_decoder(codec.ID())
	if dec == nil {
		return nil, fmt.Errorf("failed to find decoder for codec %q", codec.Name())
	}

	// Allocate a codec context for the decoder
	dec_ctx := ff.AVCodec_alloc_context(dec)
	if dec_ctx == nil {
		return nil, fmt.Errorf("failed to allocate codec context for codec %q", codec.Name())
	}

	// Create a frame for encoding - after resampling and resizing
	if frame := ff.AVUtil_frame_alloc(); frame == nil {
		return nil, errors.New("failed to allocate frame")
	} else {
		decoder.frame = frame
	}

	// Return success
	return decoder, nil
}

// Close the demuxer
func (d *demuxer) close() error {

}

// Close the decoder
func (d *decoder) close() error {

}

// Demultiplex streams from the reader
func (d *demuxer) Demux(ctx context.Context, fn DecoderFunc) error {
	// If the decoder is nil then set it to default
	if fn == nil {
		fn = func(packet Packet) error {
			if packet == nil {
				return d.decodePacket(nil)
			}
			return d.decodePacket(packet.(*ff.AVPacket))
		}
	}

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
			if err := ff.AVFormat_read_frame(d.input, packet); errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return err
			}
			stream := packet.StreamIndex()
			if decoder := d.decoders[stream]; decoder != nil {
				if err := decoder.decode(fn, packet); errors.Is(err, io.EOF) {
					break FOR_LOOP
				} else if err != nil {
					return err
				}
			}
			// Unreference the packet
			ff.AVCodec_packet_unref(packet)
		}
	}

	// Flush the decoders
	for _, decoder := range d.decoders {
		if err := decoder.decode(fn, nil); err != nil {
			return err
		}
	}

	// Return success
	return nil
}

func (d *decoder) decode(fn DecoderFunc, packet *ff.AVPacket) error {
	// Send the packet to the user defined packet function or
	// to the default version
	return fn(packet)
}

func (d *demuxer) decodePacket(packet *ff.AVPacket) error {
	// Submit the packet to the decoder. If nil then flush
	if err := ff.AVCodec_send_packet(d.codec, packet); err != nil {
		return err
	}

	// get all the available frames from the decoder
	for {
		if err := ff.AVCodec_receive_frame(d.codec, d.frame); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
			// Finished decoding packet or EOF
			break
		} else if err != nil {
			return err
		}

		fmt.Println("TODO", d.frame)
	}
	return nil
}

// Resample or resize the frame, then pass back
/*
	if frame, err := d.re(d.frame); err != nil {
		return err
	} else if err := fn(frame); errors.Is(err, io.EOF) {
		// End early
		break
	} else if err != nil {
		return err
	}*/

// Flush
/*
	if frame, err := d.re(nil); err != nil {
		return err
	} else if frame == nil {
		// NOOP
	} else if err := fn(frame); errors.Is(err, io.EOF) {
		// NOOP
	} else if err != nil {
		return err
	}
*/

/*

// Return a function to decode packets from the streams into frames
func (r *reader) Decode(fn FrameFunc) DecoderFunc {
	return func(codec Decoder, packet Packet) error {
		if packet != nil {
			// Submit the packet to the decoder
			if err := ff.AVCodec_send_packet(codec.(*decoder).codec, packet.(*ff.AVPacket)); err != nil {
				return err
			}
		} else {
			// Flush remaining frames
			if err := ff.AVCodec_send_packet(codec.(*decoder).codec, nil); err != nil {
				return err
			}
		}

		// get all the available frames from the decoder
		for {
			if err := ff.AVCodec_receive_frame(codec.(*decoder).codec, r.frame); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
				// Finished decoding packet or EOF
				break
			} else if err != nil {
				return err
			}

			// Resample or resize the frame, then pass back
			if frame, err := codec.(*decoder).re(r.frame); err != nil {
				return err
			} else if err := fn(frame); errors.Is(err, io.EOF) {
				// End early
				break
			} else if err != nil {
				return err
			}
		}

		// Flush
		if frame, err := codec.(*decoder).re(nil); err != nil {
			return err
		} else if frame == nil {
			// NOOP
		} else if err := fn(frame); errors.Is(err, io.EOF) {
			// NOOP
		} else if err != nil {
			return err
		}

		// Success
		return nil
	}
}
*/
