package media

import (
	"io"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type MediaFlag uint

////////////////////////////////////////////////////////////////////////////////
// INTERFACES

// Manager is an interface to the ffmpeg media library for media manipulation
type Manager interface {
	io.Closer

	// Open media for reading and return it
	OpenFile(path string) (Media, error)

	// Create media for writing and return it
	CreateFile(path string) (Media, error)
}

// Media is a source or destination of media
type Media interface {
	io.Closer

	// Return enumeration of streams
	Streams() []Stream

	// Return media flags for the media
	Flags() MediaFlag
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

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	MEDIA_FLAG_ALBUM             MediaFlag = (1 << iota) // Is part of an album
	MEDIA_FLAG_ALBUM_TRACK                               // Is an album track
	MEDIA_FLAG_ALBUM_COMPILATION                         // Album is a compilation
	MEDIA_FLAG_TVSHOW                                    // Is part of a TV Show
	MEDIA_FLAG_TVSHOW_EPISODE                            // Is a TV Show episode
	MEDIA_FLAG_FILE                                      // Is a file
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
