package chromaprint

import (
	"io"
	"time"
	"unsafe"

	// Packages
	"github.com/mutablelogic/go-media/sys/chromaprint"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type fingerprint struct {
	rate, channels int
	duration       time.Duration
	n              int64
	ctx            *chromaprint.Context
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new fingerprint context, with the expected sample rate,
// number of channels and the maximum duration of the data to put into
// the fingerprint. Returns nil if the context could not be created.
// If duration is zero, defaults to 120 seconds.
func New(rate, channels int, duration time.Duration) *fingerprint {
	// Check arguments
	if rate <= 0 || channels <= 0 {
		return nil
	}
	if duration <= 0 {
		duration = maxFingerprintDuration
	}

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
	return &fingerprint{rate, channels, duration, 0, ctx}
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
// return the total number of samples written
func (fingerprint *fingerprint) Write(data []int16) (int64, error) {
	if fingerprint.ctx == nil {
		return 0, io.ErrClosedPipe
	}
	if len(data) == 0 {
		return fingerprint.n, nil
	}
	if fingerprint.Duration() < fingerprint.duration {
		if err := fingerprint.ctx.WritePtr(uintptr(unsafe.Pointer(&data[0])), len(data)); err != nil {
			return 0, err
		}
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

// Return the duration of the sampled data
func (fingerprint *fingerprint) Duration() time.Duration {
	if fingerprint.ctx == nil {
		return 0
	}
	return time.Duration(fingerprint.n) * time.Second / time.Duration(fingerprint.rate*fingerprint.channels)
}
