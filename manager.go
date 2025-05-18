/*
This is a package for reading, writing and inspecting media files. In
order to operate on media, call NewManager() and then use the manager
functions to determine capabilities and manage media files and devices.
*/
package media

import "io"

// Manager represents a manager for media formats and devices.
// Create a new manager object using the NewManager function.
//
//		  import "github.com/mutablelogic/go-media/pkg/ffmpeg"
//
//		  manager, err := ffmpeg.NewManager()
//		  if err != nil {
//			...
//	      }
//
// Various options are available to control the manager, for
// logging and affecting decoding, that can be applied when
// creating the manager by passing them as arguments.
//
// Only one manager can be created. If NewManager is called
// a second time, the previously created manager is returned,
// but any new options are applied.
type Manager interface {
	// Open a media file or device for reading, from a path or url.
	// If a format is specified, then the format will be used to open
	// the file. Close the media object when done.
	Open(string, Format, ...string) (Media, error)

	// Open a media stream for reading.  If a format is
	// specified, then the format will be used to open the file. Close the
	// media object when done. It is the responsibility of the caller to
	// also close the reader when done.
	//Read(io.Reader, Format, ...string) (Media, error)

	// Create a media file or device for writing, from a path. If a format is
	// specified, then the format will be used to create the file or else
	// the format is guessed from the path. If no parameters are provided,
	// then the default parameters for the format are used.
	//Create(string, Format, []Metadata, ...Parameters) (Media, error)

	// Create a media stream for writing. The format will be used to
	// determine the format and one or more CodecParameters used to
	// create the streams. If no parameters are provided, then the
	// default parameters for the format are used. It is the responsibility
	// of the caller to also close the writer when done.
	//Write(io.Writer, Format, []Metadata, ...Parameters) (Media, error)

	// Return audio parameters for encoding
	// ChannelLayout, SampleFormat, Samplerate
	//AudioParameters(string, string, int) (Parameters, error)

	// Return video parameters for encoding
	// Width, Height, PixelFormat
	//VideoParameters(int, int, string) (Parameters, error)

	// Return codec parameters for audio encoding
	// Codec name and AudioParameters
	//AudioCodecParameters(string, AudioParameters) (Parameters, error)

	// Return codec parameters for video encoding
	// Codec name, Profile name, Framerate (fps) and VideoParameters
	//VideoCodecParameters(string, string, float64, VideoParameters) (Parameters, error)

	// Return supported input and output container formats which match any filter,
	// which can be a name, extension (with preceeding period) or mimetype. The Type
	// can be a combination of DEVICE, INPUT, OUTPUT or ANY to select the right kind of
	// format
	Formats(Type, ...string) []Format

	// Return all supported sample formats
	SampleFormats() []Metadata

	// Return all supported pixel formats
	PixelFormats() []Metadata

	// Return standard channel layouts which can be used for audio,
	// with the number of channels provided. If no channels are provided,
	// then all standard channel layouts are returned.
	ChannelLayouts() []Metadata

	// Return all supported codecs, of a specific type or all
	// if ANY is used. If any names is provided, then only the codecs
	// with those names are returned.
	Codecs(Type, ...string) []Metadata

	// Log error messages with arguments
	Errorf(string, ...any)

	// Log warning messages with arguments
	Warningf(string, ...any)

	// Log info messages  with arguments
	Infof(string, ...any)
}

// A container format for a media file or stream
type Format interface {
	// The type of the format, which can be combinations of
	// INPUT, OUTPUT, DEVICE, AUDIO, VIDEO and SUBTITLE
	Type() Type

	// The unique name that the format can be referenced as
	Name() string

	// Description of the format
	Description() string
}

// A container format for a media file, reader, device or
// network stream
type Media interface {
	io.Closer

	// The type of the format, which can be combinations of
	// INPUT, OUTPUT, DEVICE
	Type() Type
}
