package reader

import (

	// Packages

	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// PacketFn is called for each packet from the Decoder. Returning io.EOF
// stops decoding early without being treated as an error.
type PacketFn func(*ff.AVPacket) error

type Decoder struct {
	fn PacketFn
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewDecoder creates an empty Decoder that passes every packet it produces
// to fn.
func NewDecoder(fn PacketFn) (*Decoder, error) {
	if fn == nil {
		return nil, gomedia.ErrBadParameter.With("nil callback function")
	}
	return &Decoder{fn: fn}, nil
}
