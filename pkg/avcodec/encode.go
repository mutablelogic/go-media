package avcodec

import (

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type EncodingContext struct {
	codec  *Codec
	stream *ff.AVStream
	packet *ff.AVPacket
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create an encoding context (a stream and a packet) with the given codec
func NewEncodingContext(codec *Codec) (*Encoder, error) {

}
