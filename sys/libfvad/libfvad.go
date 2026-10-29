/*
In order to use this package, you will need to build libfvad, which
installs it with the other libraries:

% make libfvad

API Documentation Sources:
https://github.com/dpirch/libfvad/blob/master/include/fvad.h
*/
package libfvad

import (
	"fmt"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libfvad
#include <fvad.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Error   int
	Mode    int
	Context C.Fvad
)

////////////////////////////////////////////////////////////////////////////////
// CONSTS

// Modes, in increasing order of aggressiveness. A more aggressive mode is more
// restrictive in reporting speech, so is more likely to be right when it does,
// but misses more speech.
const (
	MODE_QUALITY         Mode = 0
	MODE_LOW_BITRATE     Mode = 1
	MODE_AGGRESSIVE      Mode = 2
	MODE_VERY_AGGRESSIVE Mode = 3
)

const (
	errNone Error = iota
	errMode
	errSampleRate
	errFrameLength
)

////////////////////////////////////////////////////////////////////////////////
// METHODS

// NewContext returns a new VAD instance in MODE_QUALITY, with a sample rate of
// 8000 Hz, or nil if memory could not be allocated. It should be released
// with Free. A context isn't safe for concurrent use.
func NewContext() *Context {
	return (*Context)(C.fvad_new())
}

// Free releases the VAD instance.
func (context *Context) Free() {
	C.fvad_free((*C.Fvad)(context))
}

// Reset clears all state, and resets the mode and sample rate to defaults.
func (context *Context) Reset() {
	C.fvad_reset((*C.Fvad)(context))
}

// SetMode sets the operating ("aggressiveness") mode.
func (context *Context) SetMode(mode Mode) error {
	if res := C.fvad_set_mode((*C.Fvad)(context), C.int(mode)); res < 0 {
		return errMode
	}
	return nil
}

// SetSampleRate sets the input sample rate in Hz, which is one of 8000,
// 16000, 32000 or 48000. Processing is done at 8000 Hz, so input at higher
// sample rates is downsampled first.
func (context *Context) SetSampleRate(rate int) error {
	if res := C.fvad_set_sample_rate((*C.Fvad)(context), C.int(rate)); res < 0 {
		return errSampleRate
	}
	return nil
}

// Process returns true if voice is detected in a frame of mono samples, which
// is 10, 20 or 30 ms long at the sample rate, so for example 160, 320 or 480
// samples at 16000 Hz.
func (context *Context) Process(frame []int16) (bool, error) {
	if len(frame) == 0 {
		return false, errFrameLength
	}
	res := C.fvad_process((*C.Fvad)(context), (*C.int16_t)(unsafe.Pointer(&frame[0])), C.size_t(len(frame)))
	if res < 0 {
		return false, errFrameLength
	}
	return res > 0, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e Error) Error() string {
	switch e {
	case errMode:
		return "libfvad: invalid mode"
	case errSampleRate:
		return "libfvad: invalid sample rate"
	case errFrameLength:
		return "libfvad: invalid frame length"
	default:
		return "libfvad: unknown error"
	}
}

func (m Mode) String() string {
	switch m {
	case MODE_QUALITY:
		return "MODE_QUALITY"
	case MODE_LOW_BITRATE:
		return "MODE_LOW_BITRATE"
	case MODE_AGGRESSIVE:
		return "MODE_AGGRESSIVE"
	case MODE_VERY_AGGRESSIVE:
		return "MODE_VERY_AGGRESSIVE"
	default:
		return fmt.Sprintf("Mode(%d)", int(m))
	}
}
