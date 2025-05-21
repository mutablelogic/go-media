package generator

import (
	"io"

	// Packages

	"github.com/mutablelogic/go-media/pkg/ffmpeg"
)

////////////////////////////////////////////////////////////////////////////
// INTERFACE

// Generator is an interface for generating frames of audio or video
type Generator interface {
	io.Closer

	// Return the next generated frame
	Frame() *ffmpeg.Frame
}
