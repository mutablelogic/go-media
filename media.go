package media

import (
	"io"
)

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
}
