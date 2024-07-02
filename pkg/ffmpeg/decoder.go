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
	re       *Re           // Resample/resize
	timeBase ff.AVRational // Timebase for the stream
	frame    *ff.AVFrame   // Destination frame
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a stream decoder which can decode packets from the input stream
func NewDecoder(stream *ff.AVStream, dest *Par, force bool) (*Decoder, error) {
	decoder := new(Decoder)
	decoder.stream = stream.Id()
	decoder.dest = dest
	decoder.timeBase = stream.TimeBase()

	// Create a frame for decoder output - before resize/resample
	frame := ff.AVUtil_frame_alloc()
	if frame == nil {
		return nil, errors.New("failed to allocate frame")
	}

	// Create a codec context for the decoder
	codec := ff.AVCodec_find_decoder(stream.CodecPar().CodecID())
	if codec == nil {
		ff.AVUtil_frame_free(frame)
		return nil, fmt.Errorf("failed to find decoder for codec %q", stream.CodecPar().CodecID())
	} else if ctx := ff.AVCodec_alloc_context(codec); ctx == nil {
		ff.AVUtil_frame_free(frame)
		return nil, fmt.Errorf("failed to allocate codec context for codec %q", codec.Name())
	} else {
		decoder.codec = ctx
		decoder.frame = frame
	}

	// If the destination codec parameters are not nil, then create a resample/resizer
	if dest != nil {
		if re, err := NewRe(dest, force); err != nil {
			return nil, errors.Join(err, decoder.Close())
		} else {
			decoder.re = re
		}
	}

	// Copy codec parameters from input stream to output codec context
	if err := ff.AVCodec_parameters_to_context(decoder.codec, stream.CodecPar()); err != nil {
		return nil, errors.Join(decoder.Close(), fmt.Errorf("failed to copy codec parameters to decoder context for codec %q", codec.Name()))
	}

	// Init the decoder
	if err := ff.AVCodec_open(decoder.codec, codec, nil); err != nil {
		return nil, errors.Join(decoder.Close(), err)
	}

	// Return success
	return decoder, nil
}

// Close the decoder and free any resources
func (d *Decoder) Close() error {
	var result error

	// Free resampler/resizer
	if d.re != nil {
		result = errors.Join(result, d.re.Close())
	}

	// Free the codec context
	if d.codec != nil {
		ff.AVCodec_free_context(d.codec)
	}

	// Free destination frame
	if d.frame != nil {
		ff.AVUtil_frame_free(d.frame)
	}

	// Reset fields
	d.re = nil
	d.codec = nil
	d.frame = nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Decode a packet into a set of frames to pass back to the
// DecoderFrameFn. If the packet is nil, then the decoder will
// flush any remaining frames.
// TODO: Optionally use the user defined packet function
// if they want to use AVParser
// TODO: The frame sent to DecoderFrameFn needs to have the
// correct timebase, etc set
func (d *Decoder) decode(packet *ff.AVPacket, fn DecoderFrameFn) error {
	if fn == nil {
		return errors.New("DecoderFrameFn is nil")
	}

	// Submit the packet to the decoder (nil packet will flush the decoder)
	if err := ff.AVCodec_send_packet(d.codec, packet); err != nil {
		return err
	}

	// get all the available frames from the decoder
	var result error
	var dest *Frame
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

		// Obtain the output frame. If a new frame is returned, it is
		// managed by the rescaler/resizer and no need to unreference it
		// later.
		if d.re != nil {
			if frame, err := d.re.Frame((*Frame)(d.frame)); err != nil {
				result = errors.Join(result, err)
			} else {
				dest = frame
			}
		} else {
			dest = (*Frame)(d.frame)
		}

		// TODO: Modify Pts?
		// What else do we need to copy across?
		fmt.Println("TODO", d.timeBase, dest.TimeBase(), ff.AVTimestamp(dest.Pts()))
		if dest.Pts() == PTS_UNDEFINED {
			(*ff.AVFrame)(dest).SetPts(d.frame.Pts())
		}

		// Pass back to the caller
		if err := fn(d.stream, dest); errors.Is(err, io.EOF) {
			// End early, return EOF
			result = io.EOF
		} else if err != nil {
			return err
		}

		// Re-allocate frame for next iteration
		ff.AVUtil_frame_unref(d.frame)
	}

	// Return success or EOF
	return result
}
