/*
This is a package for reading, writing and inspecting media files. In
order to operate on media, call NewManager() and then use the manager
functions to determine capabilities and manage media files and devices.
*/
package media

import (
	"context"
	"io"
)

// Manager represents a manager for media formats and devices.
// Create a new manager object using the NewManager function.
type Manager interface {
	// Return supported input formats which match any filter, which can be
	// a name, extension (with preceeding period) or mimetype. The MediaType
	// can be NONE (for any) or combinations of DEVICE and STREAM.
	InputFormats(MediaType, ...string) []Format

	// Return supported output formats which match any filter, which can be
	// a name, extension (with preceeding period) or mimetype. The MediaType
	// can be NONE (for any) or combinations of DEVICE and STREAM.
	OutputFormats(MediaType, ...string) []Format

	// Return supported input devices for a given format name
	// Not all devices may be supported on all platforms or listed
	// if the device does not support enumeration.
	InputDevices(string) []Device

	// Return supported output devices for a given format name
	// Not all devices may be supported on all platforms or listed
	// if the device does not support enumeration.
	OutputDevices(string) []Device

	// Open a media file or device for reading, from a path or url.
	// If a format is specified, then the format will be used to open
	// the file. Close the media object when done.
	Open(string, Format, ...string) (Media, error)

	// Open a media stream for reading.  If a format is
	// specified, then the format will be used to open the file. Close the
	// media object when done. It is the responsibility of the caller to
	// also close the reader when done.
	Read(io.Reader, Format, ...string) (Media, error)

	// Create a media file or device for writing, from a path. If a format is
	// specified, then the format will be used to create the file. Close
	// the media object when done.
	// TODO
	Create(string, Format) (Media, error)

	// Create a media stream for writing. If a format is
	// specified, then the format will be used to create the file.
	// Close the media object when done. It is the responsibility of the caller to
	// also close the writer when done.
	// TODO
	Write(io.Writer, Format) (Media, error)

	// Return version information for the media manager as a set of
	// metadata
	Version() []Metadata

	// Return all supported channel layouts
	ChannelLayouts() []Metadata

	// Return all supported sample formats
	SampleFormats() []Metadata

	// Return all supported  pixel formats
	PixelFormats() []Metadata

	// Return audio parameters for encoding
	// ChannelLayout, SampleFormat, SampleRate
	AudioParameters(string, string, int) (AudioParameters, error)

	// Return video parameters for encoding
	// Width, Height, PixelFormat, FrameRate
	VideoParameters(int, int, string, int) (VideoParameters, error)
}

// Device represents a device for input or output of media streams.
// TODO
type Device interface {
	// Device name, format depends on the device
	Name() string

	// Description of the device
	Description() string

	// Flags indicating the type INPUT or OUTPUT, AUDIO or VIDEO
	Type() MediaType

	// Whether this is the default device
	Default() bool
}

// Format represents a container format for input or output of media streams.
// Use the manager object to get a list of supported formats.
type Format interface {
	// Name(s) of the format
	Name() []string

	// Description of the format
	Description() string

	// Extensions associated with the format, if a stream. Each extension
	// should have a preceeding period (ie, ".mp4")
	Extensions() []string

	// MimeTypes associated with the format
	MimeTypes() []string

	// Flags indicating the type. INPUT for a demuxer or source, OUTPUT for a muxer or
	// sink, DEVICE for a device, FILE for a file. Plus AUDIO, VIDEO, DATA, SUBTITLE.
	Type() MediaType
}

// Media represents a media stream, which can be input or output. A new media
// object is created using the Manager object
type Media interface {
	io.Closer

	// Return a decoding context for the media stream, and
	// map the streams to decoders. If no function is provided
	// (ie, the argument is nil) then all streams are demultiplexed.
	Decoder(DecoderMapFunc) (Decoder, error)

	// Return INPUT for a demuxer or source, OUTPUT for a muxer or
	// sink, DEVICE for a device, FILE for a file or stream.
	Type() MediaType

	// Return the metadata for the media.
	Metadata() []Metadata
}

// Return parameters if a the stream should be decoded
// and either resampled or resized. Return nil if you
// want to ignore the stream, or pass identical stream
// parameters (stream.Parameters()) if you want to copy
// the stream without any changes.
type DecoderMapFunc func(Stream) (Parameters, error)

// Stream represents a audio, video, subtitle or data stream
// within a media file
type Stream interface {
	// Return AUDIO, VIDEO, SUBTITLE or DATA
	Type() MediaType

	// Return the stream parameters
	Parameters() Parameters
}

// Decoder represents a demuliplexer and decoder for a media stream.
// You can call either Demux or Decode to process the media stream,
// but not both.
type Decoder interface {
	// Demultiplex media into packets. Pass a packet to a decoder function.
	// Stop when the context is cancelled or the end of the media stream is
	// reached.
	Demux(context.Context, DecoderFunc) error

	// Decode media into frames, and resample or resize the frame.
	// Stop when the context is cancelled or the end of the media stream is
	// reached.
	Decode(context.Context, FrameFunc) error
}

// Parameters represents a set of parameters for encoding
type Parameters interface {
	AudioParameters
	VideoParameters

	// Return the media type (AUDIO, VIDEO, SUBTITLE, DATA)
	Type() MediaType
}

// Audio parameters for encoding or decoding audio data.
type AudioParameters interface {
	// Return the channel layout
	ChannelLayout() string

	// Return the sample format
	SampleFormat() string

	// Return the sample rate (Hz)
	SampleRate() int

	// TODO:
	// Planar, number of planes, bits and bytes per sample
}

// Video parameters for encoding or decoding video data.
type VideoParameters interface {
	// Return the width of the video frame
	Width() int

	// Return the height of the video frame
	Height() int

	// Return the pixel format
	PixelFormat() string

	// Return the frame rate (fps)
	FrameRate() int

	// TODO:
	// Planar, number of planes, names of the planes, bits and bytes per pixel
}

// DecoderFunc is a function that decodes a packet. Return
// io.EOF if you want to stop processing the packets early.
type DecoderFunc func(Packet) error

// FrameFunc is a function that processes a frame of audio
// or video data.  Return io.EOF if you want to stop
// processing the frames early.
type FrameFunc func(Frame) error

// Packet represents a packet of demultiplexed data.
// Currently this is quite opaque!
type Packet interface{}

// Frame represents a frame of audio or picture data.
// Currently this is quite opaque - should allow access to
// the audio sample data, or the individual pixel data!
type Frame interface{}

// Metadata represents a metadata entry for a media stream.
// Currently this is quite opaque!
type Metadata interface{}
