package audio

import (
	"mime"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func init() {
	mime.AddExtensionType(".m4a", "audio/mp4")
	mime.AddExtensionType(".flac", "audio/flac")
}
