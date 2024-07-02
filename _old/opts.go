package media

import (
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Opt func(*opts) error

// Logging function which is used to log messages
type LogFunc func(string)

type opts struct {
	level    ffmpeg.AVLog
	callback LogFunc
	force    bool
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set a logging function
func OptLog(verbose bool, fn LogFunc) Opt {
	return func(o *opts) error {
		if verbose {
			o.level = ffmpeg.AV_LOG_VERBOSE
		} else {
			o.level = ffmpeg.AV_LOG_FATAL
		}
		o.callback = fn
		return nil
	}
}

// Force resampling and resizing on decode, even if the input and output
// parameters are the same
func OptForce() Opt {
	return func(o *opts) error {
		o.force = true
		return nil
	}
}
