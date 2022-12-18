package chromaprint

import (
	"io"
	"unsafe"

	// Packages
	"github.com/mutablelogic/go-media/sys/chromaprint"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Chromaprint is a wrapper around the chromaprint library. Create a new
// Chromaprint object by calling New(rate,channels)
type Chromaprint interface {
	io.Closer

	// Write sample data to the fingerprinter. Expects 16-bit signed integers
	// and returns number of samples written
	Write([]int16) (int64, error)

	// Finish the fingerprinting, and compute the fingerprint, return as a
	// string
	Finish() (string, error)
}

type fingerprint struct {
	n   int64
	ctx *chromaprint.Context
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new fingerprint context
func New(rate, channels int) *fingerprint {
	// Create a context
	ctx := chromaprint.NewChromaprint(chromaprint.ALGORITHM_DEFAULT)
	if ctx == nil {
		return nil
	}
	// Start the fingerprinting
	if err := ctx.Start(rate, channels); err != nil {
		ctx.Free()
		return nil
	}
	// Return success
	return &fingerprint{0, ctx}
}

// Close the fingerprint to release resources
func (fingerprint *fingerprint) Close() error {
	// Free the context
	if fingerprint.ctx != nil {
		fingerprint.ctx.Free()
	}

	// Release resources
	fingerprint.ctx = nil

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Write 16-bit signed integers with little-endian format, and
// return the total number of bytes written
func (fingerprint *fingerprint) Write(data []int16) (int64, error) {
	if fingerprint.ctx == nil {
		return 0, io.ErrClosedPipe
	}
	if err := fingerprint.ctx.WritePtr(uintptr(unsafe.Pointer(&data[0])), len(data)); err != nil {
		return 0, err
	}
	fingerprint.n += int64(len(data))
	return fingerprint.n, nil
}

// Finish the fingerprinting, and compute the fingerprint, return as a string
func (fingerprint *fingerprint) Finish() (string, error) {
	if fingerprint.ctx == nil {
		return "", io.ErrClosedPipe
	}
	if err := fingerprint.ctx.Finish(); err != nil {
		return "", err
	}
	return fingerprint.ctx.GetFingerprint()
}
