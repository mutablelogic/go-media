package media

import (
	"context"
	"io"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// MediaFlag is a bitfield of flags for media, including type of media
type MediaFlag uint

// MediaKey is a string which is used for media metadata
type MediaKey string

// Demux is a function which is called for each packet in the media, which
// is associated with a single stream. The function should return an error if
// the decode should be terminated.
type DemuxFn func(context.Context, Packet) error

// DecodeFn is a function which is called for each frame in the media, which
// is associated with a single stream. The function should return an error if
// the decode should be terminated.
type DecodeFn func(context.Context, Frame) error

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Manager is an interface to the ffmpeg media library for media manipulation
type Manager interface {
	io.Closer

	// Enumerate formats with MEDIA_FLAG_ENCODER, MEDIA_FLAG_DECODER,
	// MEDIA_FLAG_FILE and MEDIA_FLAG_DEVICE flags to filter.
	// Lookups can be further filtered by name, mimetype and extension
	MediaFormats(MediaFlag, ...string) []MediaFormat

	// Open media file for reading and return it. A format can be specified
	// to "force" a specific format
	OpenFile(string, MediaFormat) (Media, error)

	// Open media URL for reading and return it. A format can be specified
	// to "force" a specific format
	OpenURL(string, MediaFormat) (Media, error)

	// Open media device with a specific name for reading and return it.
	OpenDevice(string) (Media, error)

	// Create file for writing and return it
	CreateFile(string) (Media, error)

	// Create an output device with a specific name for writing and return it
	CreateDevice(string) (Media, error)

	// Create a map of input media. If MediaFlag is MEDIA_FLAG_NONE, then
	// all audio, video and subtitle streams are mapped, or else a
	// combination of MEDIA_FLAG_AUDIO,
	// MEDIA_FLAG_VIDEO, MEDIA_FLAG_SUBTITLE and MEDIA_FLAG_DATA
	// can be used to map specific types of streams.
	Map(Media, MediaFlag) (Map, error)

	// Demux a media file, passing packets to a callback function
	Demux(context.Context, Map, DemuxFn) error

	// Decode a packet into a series of frames, passing decoded frames to
	// a callback function
	Decode(context.Context, Map, Packet, DecodeFn) error

	// Log messages from ffmpeg
	SetDebug(bool)
}

// MediaFormat is an input or output format for media items
type MediaFormat interface {
	// Return the names of the media format
	Name() []string

	// Return a longer description of the media format
	Description() string

	// Return MEDIA_FLAG_ENCODER, MEDIA_FLAG_DECODER, MEDIA_FLAG_FILE
	// and MEDIA_FLAG_DEVICE flags
	Flags() MediaFlag

	// Return mimetypes handled
	MimeType() []string

	// Return file extensions handled
	Ext() []string

	// Return the default audio codec for the format
	DefaultAudioCodec() Codec

	// Return the default video codec for the format
	DefaultVideoCodec() Codec

	// Return the default subtitle codec for the format
	DefaultSubtitleCodec() Codec
}

// Map is a mapping of input media, potentially to output media
type Map interface {
	// Return input media
	Input() Media

	// Return a single stream which is mapped for decoding, filtering by
	// stream type. Returns nil if there is no selection of that type
	Streams(MediaFlag) []Stream

	// Print a summary of the mapping
	PrintMap(w io.Writer)

	// Resample an audio stream
	Resample(AudioFormat, Stream) error

	// Encode to output media using default codec from a specific stream
	//Encode(Media, Stream) error
}

// Media is a source or destination of media
type Media interface {
	io.Closer

	// URL for the media
	URL() string

	// Return enumeration of streams
	Streams() []Stream

	// Return media flags for the media
	Flags() MediaFlag

	// Return the format of the media
	Format() MediaFormat

	// Return metadata for the media
	Metadata() Metadata

	// Set metadata value by key, or remove it if the value is nil
	Set(MediaKey, any) error
}

// Stream of data multiplexed in the media
type Stream interface {
	// Return index of stream in the media
	Index() int

	// Return media flags for the stream
	Flags() MediaFlag

	// Return artwork for the stream - if MEDIA_FLAG_ARTWORK is set
	Artwork() []byte
}

// Metadata embedded in the media
type Metadata interface {
	// Return enumeration of keys
	Keys() []MediaKey

	// Return value for key
	Value(MediaKey) any
}

// Packet is a single unit of data in the media
type Packet interface {
	// Flags returns the flags for the packet from the stream
	Flags() MediaFlag

	// Stream returns the stream which the packet belongs to
	Stream() Stream

	// IsKeyFrame returns true if the packet contains a key frame
	IsKeyFrame() bool

	// Pos returns the byte position of the packet in the media
	Pos() int64

	// Duration returns the duration of the packet
	Duration() time.Duration

	// Size of the packet in bytes
	Size() int

	// Bytes returns the raw bytes of the packet
	Bytes() []byte
}

// Frame is a decoded video or audio frame
type Frame interface {
	AudioFrame
	VideoFrame

	// Returns MEDIA_FLAG_VIDEO, MEDIA_FLAG_AUDIO
	Flags() MediaFlag

	// Returns true if planar format
	//IsPlanar() bool

	// Returns the samples for a specified channel, as array of bytes. For packed
	// audio format, the channel should be 0.
	//Bytes(channel int) []byte
}

type AudioFrame interface {
	// Returns the audio format, if MEDIA_FLAG_AUDIO is set
	AudioFormat() AudioFormat

	// Number of samples, if MEDIA_FLAG_AUDIO is set
	NumSamples() int

	// Audio channels, if MEDIA_FLAG_AUDIO is set
	Channels() []AudioChannel

	// Duration of the frame, if MEDIA_FLAG_AUDIO is set
	Duration() time.Duration
}

type VideoFrame interface {
	// Returns the audio format, if MEDIA_FLAG_VIDEO is set
	PixelFormat() PixelFormat

	// Return frame width and height, if MEDIA_FLAG_VIDEO is set
	Size() (int, int)
}

// Codec is an encoder or decoder for a specific media type
type Codec interface {
	// Name returns the unique name for the codec
	Name() string

	// Description returns the long description for the codec
	Description() string

	// Flags for the codec (Audio, Video, Encoder, Decoder)
	Flags() MediaFlag
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	MEDIA_FLAG_ALBUM             MediaFlag = (1 << iota) // Is part of an album
	MEDIA_FLAG_ALBUM_TRACK                               // Is an album track
	MEDIA_FLAG_ALBUM_COMPILATION                         // Album is a compilation
	MEDIA_FLAG_TVSHOW                                    // Is part of a TV Show
	MEDIA_FLAG_TVSHOW_EPISODE                            // Is a TV Show episode
	MEDIA_FLAG_FILE                                      // Is a file
	MEDIA_FLAG_DEVICE                                    // Is a device
	MEDIA_FLAG_VIDEO                                     // Contains video
	MEDIA_FLAG_AUDIO                                     // Contains audio
	MEDIA_FLAG_SUBTITLE                                  // Contains subtitles
	MEDIA_FLAG_DATA                                      // Contains data stream
	MEDIA_FLAG_ATTACHMENT                                // Contains attachment
	MEDIA_FLAG_ARTWORK                                   // Contains artwork
	MEDIA_FLAG_CAPTIONS                                  // Contains captions
	MEDIA_FLAG_ENCODER                                   // Is an encoder
	MEDIA_FLAG_DECODER                                   // Is an decoder
	MEDIA_FLAG_NONE              MediaFlag = 0
	MEDIA_FLAG_MAX                         = MEDIA_FLAG_DECODER
)

const (
	MEDIA_KEY_BRAND_MAJOR      MediaKey = "major_brand"       // string
	MEDIA_KEY_BRAND_COMPATIBLE MediaKey = "compatible_brands" // string
	MEDIA_KEY_CREATED          MediaKey = "creation_time"     // time.Time
	MEDIA_KEY_ENCODER          MediaKey = "encoder"           // string
	MEDIA_KEY_ALBUM            MediaKey = "album"             // string
	MEDIA_KEY_ALBUM_ARTIST     MediaKey = "artist"            // string
	MEDIA_KEY_COMMENT          MediaKey = "comment"           // string
	MEDIA_KEY_COMPOSER         MediaKey = "composer"          // string
	MEDIA_KEY_COPYRIGHT        MediaKey = "copyright"         // string
	MEDIA_KEY_YEAR             MediaKey = "date"              // uint
	MEDIA_KEY_DISC             MediaKey = "disc"              // uint xx or xx/yy
	MEDIA_KEY_ENCODED_BY       MediaKey = "encoded_by"        // string
	MEDIA_KEY_FILENAME         MediaKey = "filename"          // string
	MEDIA_KEY_GENRE            MediaKey = "genre"             // string
	MEDIA_KEY_LANGUAGE         MediaKey = "language"          // string
	MEDIA_KEY_PERFORMER        MediaKey = "performer"         // string
	MEDIA_KEY_PUBLISHER        MediaKey = "publisher"         // string
	MEDIA_KEY_SERVICE_NAME     MediaKey = "service_name"      // string
	MEDIA_KEY_SERVICE_PROVIDER MediaKey = "service_provider"  // string
	MEDIA_KEY_TITLE            MediaKey = "title"             // string
	MEDIA_KEY_TRACK            MediaKey = "track"             // uint xx or xx/yy
	MEDIA_KEY_VERSION_MAJOR    MediaKey = "major_version"     // string
	MEDIA_KEY_VERSION_MINOR    MediaKey = "minor_version"     // string
	MEDIA_KEY_SHOW             MediaKey = "show"              // string
	MEDIA_KEY_SEASON           MediaKey = "season_number"     // uint
	MEDIA_KEY_EPISODE_SORT     MediaKey = "episode_sort"      // string
	MEDIA_KEY_EPISODE_ID       MediaKey = "episode_id"        // uint
	MEDIA_KEY_COMPILATION      MediaKey = "compilation"       // bool
	MEDIA_KEY_GAPLESS_PLAYBACK MediaKey = "gapless_playback"  // bool
	MEDIA_KEY_ACCOUNT_ID       MediaKey = "account_id"        // string
	MEDIA_KEY_DESCRIPTION      MediaKey = "description"       // string
	MEDIA_KEY_MEDIA_TYPE       MediaKey = "media_type"        // string
	MEDIA_KEY_PURCHASED        MediaKey = "purchase_date"     // time.Time
	MEDIA_KEY_ALBUM_SORT       MediaKey = "sort_album"        // string
	MEDIA_KEY_ARTIST_SORT      MediaKey = "sort_artist"       // string
	MEDIA_KEY_TITLE_SORT       MediaKey = "sort_name"         // string
	MEDIA_KEY_SYNOPSIS         MediaKey = "synopsis"          // string
	MEDIA_KEY_GROUPING         MediaKey = "grouping"          // string
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (f MediaFlag) String() string {
	if f == MEDIA_FLAG_NONE {
		return f.FlagString()
	}
	str := ""
	for v := MediaFlag(1); v <= MEDIA_FLAG_MAX; v <<= 1 {
		if f&v == v {
			str += "|" + v.FlagString()
		}
	}
	return str[1:]
}

func (f MediaFlag) FlagString() string {
	switch f {
	case MEDIA_FLAG_NONE:
		return "MEDIA_FLAG_NONE"
	case MEDIA_FLAG_ALBUM:
		return "MEDIA_FLAG_ALBUM"
	case MEDIA_FLAG_ALBUM_TRACK:
		return "MEDIA_FLAG_ALBUM_TRACK"
	case MEDIA_FLAG_ALBUM_COMPILATION:
		return "MEDIA_FLAG_ALBUM_COMPILATION"
	case MEDIA_FLAG_TVSHOW:
		return "MEDIA_FLAG_TVSHOW"
	case MEDIA_FLAG_TVSHOW_EPISODE:
		return "MEDIA_FLAG_TVSHOW_EPISODE"
	case MEDIA_FLAG_FILE:
		return "MEDIA_FLAG_FILE"
	case MEDIA_FLAG_DEVICE:
		return "MEDIA_FLAG_DEVICE"
	case MEDIA_FLAG_VIDEO:
		return "MEDIA_FLAG_VIDEO"
	case MEDIA_FLAG_AUDIO:
		return "MEDIA_FLAG_AUDIO"
	case MEDIA_FLAG_SUBTITLE:
		return "MEDIA_FLAG_SUBTITLE"
	case MEDIA_FLAG_DATA:
		return "MEDIA_FLAG_DATA"
	case MEDIA_FLAG_ATTACHMENT:
		return "MEDIA_FLAG_ATTACHMENT"
	case MEDIA_FLAG_ARTWORK:
		return "MEDIA_FLAG_ARTWORK"
	case MEDIA_FLAG_CAPTIONS:
		return "MEDIA_FLAG_CAPTIONS"
	case MEDIA_FLAG_ENCODER:
		return "MEDIA_FLAG_ENCODER"
	case MEDIA_FLAG_DECODER:
		return "MEDIA_FLAG_DECODER"
	default:
		return "[?? Invalid MediaFlag]"
	}
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

func (f MediaFlag) Is(v MediaFlag) bool {
	return f&v == v
}
