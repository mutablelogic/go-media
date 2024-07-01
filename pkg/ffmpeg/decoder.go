package ffmpeg

import (
	"errors"
	"fmt"
	"io"
	"syscall"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Decoder struct {
	stream   int
	codec    *ff.AVCodecContext
	dest     *Par          // Destination parameters
	timeBase ff.AVRational // Timebase for the stream
	frame    *ff.AVFrame   // Destination frame
}

// DecoderFrameFn is a function which is called to send a frame after decoding. It should
// return nil to continue decoding or io.EOF to stop.
type DecoderFrameFn func(int, *Frame) error

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a stream decoder which can decode packets from the input stream
// TODO: resample and resize frames to the destination parameters
func NewDecoder(stream *ff.AVStream, dest *Par, force bool) (*Decoder, error) {
	decoder := new(Decoder)
	decoder.stream = stream.Id()
	decoder.dest = dest
	decoder.timeBase = stream.TimeBase()

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
		return nil, errors.Join(decoder.Close(), fmt.Errorf("failed to copy codec parameters to decoder context for codec %q", codec.Name()))
	}

	// Init the decoder
	if err := ff.AVCodec_open(decoder.codec, codec, nil); err != nil {
		return nil, errors.Join(decoder.Close(), err)
	}

	// Create a frame for decoder output - before resize/resample
	if frame := ff.AVUtil_frame_alloc(); frame == nil {
		return nil, errors.Join(decoder.Close(), errors.New("failed to allocate frame"))
	} else {
		decoder.frame = frame
	}

	// Return success
	return decoder, nil
}

// Close the decoder and free any resources
func (d *Decoder) Close() error {
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

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (d *Decoder) decode(packet *ff.AVPacket, fn DecoderFrameFn) error {
	if fn == nil {
		return errors.New("DecoderFrameFn is nil")
	}

	//if demuxfn != nil {
	// Send the packet (or a nil to flush) to the user defined packet function
	//	return demuxfn(newPacket(packet, d.stream, d.codec.Codec().Type(), d.timeBase))
	//}

	// Submit the packet to the decoder (nil packet will flush the decoder)
	if err := ff.AVCodec_send_packet(d.codec, packet); err != nil {
		return err
	}

	// get all the available frames from the decoder
	var result error
	for {
		// End early if we've received an EOF
		if result != nil {
			break
		}

		// Receive the next frame from the decoder
		if err := ff.AVCodec_receive_frame(d.codec, d.frame); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
			// Finished decoding packet or EOF
			break
		} else if err != nil {
			return err
		}

		// Resample or resize the frame, then pass to the frame function
		//frame, err := d.re(d.frame)
		//if err != nil {
		//	return err
		//}

		// Copy over the timebase and ptr from the stream
		d.frame.SetTimeBase(d.timeBase)
		d.frame.SetPts(d.frame.Pts())

		// Pass back to the caller
		if err := fn(d.stream, (*Frame)(d.frame)); errors.Is(err, io.EOF) {
			// End early, return EOF
			result = io.EOF
		} else if err != nil {
			return err
		}

		// Re-allocate frames for next iteration
		ff.AVUtil_frame_unref(d.frame)
		//		ff.AVUtil_frame_unref(d.reframe)
	}

	// Flush the resizer or resampler if we haven't received an EOF
	/*
		if result == nil {
			finished := false
			for {
				if finished {
					break
				}
				if frame, err := d.reflush(d.frame); err != nil {
					return err
				} else if frame == nil {
					finished = true
				} else if err := framefn(newFrame(frame)); errors.Is(err, io.EOF) {
					finished = true
				} else if err != nil {
					return err
				}

				// Re-allocate frames for next iteration
				ff.AVUtil_frame_unref(d.frame)
				ff.AVUtil_frame_unref(d.reframe)
			}
		}
	*/

	// Return success or EOF
	return result
}
