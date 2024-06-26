package ffmpeg

import (
	// Package imports
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Opt func(*opts) error

type opts struct {
	// Resize/resample options
	force bool

	// Writer options
	oformat  *ffmpeg.AVOutputFormat
	streams  map[int]*Par
	metadata []*Metadata

	// Reader options
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
