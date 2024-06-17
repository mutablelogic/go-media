package media

import (
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////
// TYPES

// Media Types: Audio, Video, Subtitle or Data
type MediaType int

////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	AUDIO = MediaType(ff.AVMEDIA_TYPE_AUDIO) // Audio media type
	VIDEO = MediaType(ff.AVMEDIA_TYPE_VIDEO) // Video media type
)
