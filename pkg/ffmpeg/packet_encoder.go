package ffmpeg

import (
	"errors"
	"io"
	"syscall"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type encoder struct {
	codec  *ff.AVCodecContext
	packet *ff.AVPacket
	stream int
}

// EncodeFn is a function which is called for each packet encoded. It should
// return nil to continue encoding or io.EOF to stop decoding.
type EncodeFn func(*ff.AVPacket) error

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewPacketEncoder(codec *ff.AVCodecContext, stream int) (*encoder, error) {
	encoder := new(encoder)
	encoder.codec = codec

	// Create a packet
	packet := ff.AVCodec_packet_alloc()
	if packet == nil {
		return nil, errors.New("failed to allocate packet")
	} else {
		encoder.packet = packet
		encoder.stream = stream
	}

	// Return success
	return encoder, nil
}

func (e *encoder) Close() error {
	if e.packet != nil {
		ff.AVCodec_packet_free(e.packet)
		e.packet = nil
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (e *encoder) Encode(frame *ff.AVFrame, fn EncodeFn) error {
	// Encode a frame
	if err := e.encode(frame, fn); err != nil {
		return err
	}
	// Flush
	return e.encode(nil, fn)
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (e *encoder) encode(frame *ff.AVFrame, fn EncodeFn) error {
	// Send the frame to the encoder
	if err := ff.AVCodec_send_frame(e.codec, frame); err != nil {
		return err
	}

	// Write out the packets
	var result error
	for {
		// Receive the packet
		if err := ff.AVCodec_receive_packet(e.codec, e.packet); errors.Is(err, syscall.EAGAIN) || errors.Is(err, io.EOF) {
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
