package ffmpeg

import (
	"errors"
	"io"
	"syscall"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

type decoder struct {
	codec  *ff.AVCodecContext
	frame  *ff.AVFrame
	stream int
}

// DecodeFn is a function which is called for each frame decoded
// with the stream id of the packet. It should return nil to continue
// decoding or io.EOF to stop decoding.
type DecodeFn func(int, *ff.AVFrame) error

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewPacketDecoder(codec *ff.AVCodecContext, stream int) (*decoder, error) {
	decoder := new(decoder)
	decoder.codec = codec

	// Create a frame
	frame := ff.AVUtil_frame_alloc()
	if frame == nil {
		return nil, errors.New("failed to allocate frame")
	} else {
		decoder.frame = frame
		decoder.stream = stream
	}

	// Return success
	return decoder, nil
}

func (d *decoder) Close() error {
	if d.frame != nil {
		ff.AVUtil_frame_free(d.frame)
		d.frame = nil
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Decode a packet using the decoder and pass frames onto the FrameFn
// function. The FrameFn function should return nil to continue decoding
// or io.EOF to stop decoding.
func (d *decoder) Decode(packet *ff.AVPacket, fn DecodeFn) error {
	// Submit the packet to the decoder (nil packet will flush the decoder)
	if err := ff.AVCodec_send_packet(d.codec, packet); err != nil {
		return err
	}

	// get all the available frames from the decoder
	var result error
	for {
		if err := ff.AVCodec_receive_frame(d.codec, d.frame); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
			// Finished decoding packet or EOF
			break
		} else if err != nil {
			return err
		}

		// Pass back to the caller
		if err := fn(d.stream, d.frame); errors.Is(err, io.EOF) {
			// End early, return EOF
			result = io.EOF
			break
		} else if err != nil {
			return err
		}

		// Re-allocate frames for next iteration
		ff.AVUtil_frame_unref(d.frame)
	}

	// Flush
	if result == nil {
		result = fn(d.stream, nil)
	}

	// Return success or EOF
	return result
}
