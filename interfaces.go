/*
This is a package for reading, writing and inspecting media files. In
order to operate on media, call NewManager() and then use the manager
functions to determine capabilities and manage media files and devices.
*/
package media

import (
	"context"
	"image"
	"io"
	"time"
)

// Manager represents a manager for media formats and devices.
// Create a new manager object using the NewManager function.
type Manager interface {
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
	// specified, then the format will be used to create the file or else
	// the format is guessed from the path. If no parameters are provided,
	// then the default parameters for the format are used.
	Create(string, Format, []Metadata, ...Parameters) (Media, error)

	// Create a media stream for writing. The format will be used to
	// determine the formar type and one or more CodecParameters used to
	// create the streams. If no parameters are provided, then the
	// default parameters for the format are used. It is the responsibility
	// of the caller to also close the writer when done.
	Write(io.Writer, Format, []Metadata, ...Parameters) (Media, error)

	// Return supported input formats which match any filter, which can be
	// a name, extension (with preceeding period) or mimetype. The MediaType
	// can be NONE (for any) or combinations of DEVICE and STREAM.
	InputFormats(MediaType, ...string) []Format

	// Return supported output formats which match any filter, which can be
	// a name, extension (with preceeding period) or mimetype. The MediaType
	// can be NONE (for any) or combinations of DEVICE and STREAM.
	OutputFormats(MediaType, ...string) []Format

	// Return supported devices for a given format.
	// Not all devices may be supported on all platforms or listed
	// if the device does not support enumeration.
	Devices(Format) []Device

	// Return all supported channel layouts
	ChannelLayouts() []Metadata

	// Return all supported sample formats
	SampleFormats() []Metadata

	// Return all supported  pixel formats
	PixelFormats() []Metadata

	// Return all supported codecs
	Codecs() []Metadata

	// Return audio parameters for encoding
	// ChannelLayout, SampleFormat, Samplerate
	AudioParameters(string, string, int) (Parameters, error)

	// Return video parameters for encoding
	// Width, Height, PixelFormat
	VideoParameters(int, int, string) (Parameters, error)

	// Return codec parameters for audio encoding
	// Codec name and AudioParameters
	//AudioCodecParameters(string, AudioParameters) (Parameters, error)

	// Return codec parameters for video encoding
	// Codec name, Profile name, Framerate (fps) and VideoParameters
	//VideoCodecParameters(string, string, float64, VideoParameters) (Parameters, error)

	// Return version information for the media manager as a set of
	// metadata
	Version() []Metadata

	// Log error messages with arguments
	Errorf(string, ...any)

	// Log warning messages with arguments
	Warningf(string, ...any)

	// Log info messages  with arguments
	Infof(string, ...any)
}

// Device represents a device for input or output of media streams.
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

	// Return the metadata for the media, filtering by keys if any
	// are included. Use the "artwork" key to return only artwork.
	Metadata(...string) []Metadata
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

	// Return number of planes for a specific PixelFormat
	// or SampleFormat and ChannelLayout combination
	NumPlanes() int
}

// Audio parameters for encoding or decoding audio data.
type AudioParameters interface {
	// Return the channel layout
	ChannelLayout() string

	// Return the sample format
	SampleFormat() string

	// Return the sample rate (Hz)
	Samplerate() int
}

// Video parameters for encoding or decoding video data.
type VideoParameters interface {
	// Return the width of the video frame
	Width() int

	// Return the height of the video frame
	Height() int

	// Return the pixel format
	PixelFormat() string
}

// DecoderFunc is a function that decodes a packet. Return
// io.EOF if you want to stop processing the packets early.
type DecoderFunc func(Packet) error

// FrameFunc is a function that processes a frame of audio
// or video data.  Return io.EOF if you want to stop
// processing the frames early.
type FrameFunc func(Frame) error

// Codec represents a codec for encoding or decoding media streams.
type Codec interface {
	// Return the codec name
	Name() string

	// Return the codec description
	Description() string

	// Return the codec type (AUDIO, VIDEO, SUBTITLE, DATA, INPUT, OUTPUT)
	Type() MediaType
}

// Packet represents a packet of demultiplexed data.
// Currently this is quite opaque!
type Packet interface{}

// Frame represents a frame of audio or video data.
type Frame interface {
	Parameters

	// Number of samples in the frame, for an audio frame.
	NumSamples() int

	// Return stride for each plane, for a video frame.
	Stride(int) int

	// Return a frame plane as a byte slice.
	Bytes(int) []byte

	// Return the presentation timestamp for the frame or
	// a negative number if not set
	Time() time.Duration

	// Return a frame as an image, which supports the following
	// pixel formats: AV_PIX_FMT_GRAY8, AV_PIX_FMT_RGBA,
	// AV_PIX_FMT_RGB24, AV_PIX_FMT_YUV420P
	Image() (image.Image, error)
}

// Metadata represents a metadata entry for a media stream.
type Metadata interface {
	// Return the metadata key for the entry
	Key() string

	// Return the metadata value for the entry
	Value() any
}
