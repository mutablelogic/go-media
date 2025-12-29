package ffmpeg

import (
	"errors"

	// Package imports
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Option which can affect the behaviour of ffmpeg
type Opt func(*opts) error

// Logging function which is used to log messages
type LogFunc func(string)

type opts struct {
	// Logging options
	level    ffmpeg.AVLog
	callback LogFunc

	// Resize/resample options
	force bool

	// Reader options
	t       media.Type
	iformat *ffmpeg.AVInputFormat
	opts    []string // These are key=value pairs
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	bufSize = 4096
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newOpts() *opts {
	return new(opts)
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

// Input format from name or url
func OptInputFormat(name string) Opt {
	return func(o *opts) error {
		// By name
		if iformat := ffmpeg.AVFormat_find_input_format(name); iformat != nil {
			o.iformat = iformat
		} else {
			return errors.New("invalid input format")
		}
		return nil
	}
}

// Input format options
func OptInputOpt(opt ...string) Opt {
	return func(o *opts) error {
		o.opts = append(o.opts, opt...)
		return nil
	}
}
