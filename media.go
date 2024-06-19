/* media is a package for reading and writing media files. */
package media

import "io"

// Manager represents a manager for media formats. Create a new manager
// object using the NewManager function.
type Manager interface {
	// Return supported input formats which match any filter, which can be
	// a name, extension (with preceeding period) or mimetype. The MediaType
	// can be NONE (for any) or combinations of DEVICE and STREAM.
	InputFormats(MediaType, ...string) []Format

	// Return supported output formats which match any filter, which can be
	// a name, extension (with preceeding period) or mimetype. The MediaType
	// can be NONE (for any) or combinations of DEVICE and STREAM.
	OutputFormats(MediaType, ...string) []Format

	// Return supported input devices for a given name
	InputDevices(string) []Device

	// Return supported output devices for a given name
	OutputDevices(string) []Device

	// Open a media file for reading, from a path or url. If a format is
	// specified, then the format will be used to open the file. Close the
	// media object when done.
	Open(string, Format) (Media, error)

	// Open a media stream for reading.  If a format is
	// specified, then the format will be used to open the file. Close the
	// media object when done. It is the responsibility of the caller to
	// also close the reader when done.
	Read(io.Reader, Format) (Media, error)

	// Create a media file for writing, from a path. If a format is
	// specified, then the format will be used to create the file. Close
	// the media object when done.
	Create(string, Format) (Media, error)

	// Create a media stream for writing. If a format is
	// specified, then the format will be used to create the file.
	// Close the media object when done. It is the responsibility of the caller to
	// also close the writer when done.
	Write(io.Writer, Format) (Media, error)
}

// Device represents a device for input or output of media streams.
type Device interface {
	// Device name, format depends on the device
	Name() string

	// Description of the device
	Description() string

	// Flags indicating the type
	Type() MediaType

	// Whether this is the default device
	Default() bool
}

// Format represents a container format for input or output of media streams.
type Format interface {
	// Name(s) of the format
	Name() []string

	// Description of the format
	Description() string

	// Extensions associated with the format, if a stream
	Extensions() []string

	// MimeTypes associated with the format
	MimeTypes() []string

	// Flags indicating the type. INPUT for a demuxer or source, OUTPUT for a muxer or sink, DEVICE for a device, FILE for a file.
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
