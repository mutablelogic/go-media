/* media is a package for reading and writing media files. */
package media

import "io"

// Format represents a container format for input or output of media streams.
type Format interface {
	// Name(s) of the format
	Name() []string

	// Description of the format
	Description() string

	// Extensions associated with the format
	Extensions() []string

	// MimeTypes associated with the format
	MimeTypes() []string

	// INPUT for a demuxer, OUTPUT for a muxer
	Type() MediaType
}

// Media represents a media stream, which can
// be input or output. A new media object is created
// using NewReader, Open, NewWriter or Create.
type Media interface {
	io.Closer

	// Return the metadata for the media stream.
	Metadata() []Metadata

	// Demultiplex media (when NewReader or Open has
	// been used). Pass a packet to a decoder function.
	Demux(DecoderFunc) error

	// Return a decode function, which can rescale or
	// resample a frame and then call a frame processing
	// function for encoding and multiplexing.
	Decode(FrameFunc) DecoderFunc
}

// Decoder represents a decoder for a media stream.
type Decoder interface{}

// DecoderFunc is a function that decodes a packet
type DecoderFunc func(Decoder, Packet) error

// FrameFunc is a function that processes a frame of audio
// or video data.
type FrameFunc func(Frame) error

// Packet represents a packet of demultiplexed data.
type Packet interface{}

// Frame represents a frame of audio or video data.
type Frame interface{}

// Metadata represents a metadata entry for a media stream.
type Metadata interface{}
