package ffmpeg

import (
	// Package imports
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
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

	// Writer options
	oformat  *ffmpeg.AVOutputFormat
	streams  map[int]*Par
	metadata []*Metadata

	// Reader options
	t       media.Type
	iformat *ffmpeg.AVInputFormat
	opts    []string // These are key=value pairs
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newOpts() *opts {
	return &opts{
		streams: make(map[int]*Par),
	}
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

// Output format from name or url
func OptOutputFormat(name string) Opt {
	return func(o *opts) error {
		// By name
		if oformat := ffmpeg.AVFormat_guess_format(name, name, name); oformat != nil {
			o.oformat = oformat
		} else {
			return ErrBadParameter.Withf("invalid output format %q", name)
		}
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
			return ErrBadParameter.Withf("invalid input format %q", name)
		}
		return nil
	}
}

// Input format from ff.AVInputFormat
func optInputFormat(format *Format) Opt {
	return func(o *opts) error {
		if format != nil && format.Input != nil {
			o.iformat = format.Input
			o.t = format.Type()
		} else {
			return ErrBadParameter.With("invalid input format")
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

// New stream with parameters
func OptStream(stream int, par *Par) Opt {
	return func(o *opts) error {
		if par == nil {
			return ErrBadParameter.With("invalid parameters")
		}
		if stream == 0 {
			stream = len(o.streams) + 1
		}
		if _, exists := o.streams[stream]; exists {
			return ErrDuplicateEntry.Withf("stream %v", stream)
		}
		if stream < 0 {
			return ErrBadParameter.Withf("invalid stream %v", stream)
		}
		o.streams[stream] = par

		// Return success
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

// Append metadata to the output file, including artwork
func OptMetadata(entry ...*Metadata) Opt {
	return func(o *opts) error {
		o.metadata = append(o.metadata, entry...)
		return nil
	}
}
