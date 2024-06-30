package generator

import (
	"io"

	// Packages
	media "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////
// INTERFACE

// Generator is an interface for generating frames of audio or video
type Generator interface {
	io.Closer

	// Return a generated frame
	Frame() media.Frame
}
