package media

import (
	"io"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Chromaprint is a wrapper around the chromaprint library. Create a new
// Chromaprint object by calling chromaprint.New(sample_rate, channels)
type Chromaprint interface {
	io.Closer

	// Write sample data to the fingerprinter. Expects 16-bit signed integers
	// and returns number of samples written
	Write([]int16) (int64, error)

	// Finish the fingerprinting, and compute the fingerprint, return as a
	// string
	Finish() (string, error)
}
