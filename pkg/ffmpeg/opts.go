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

	// Writer options
	oformat  *ffmpeg.AVOutputFormat
	streams  map[int]*Par
	metadata []*Metadata
	copy     bool // If true, copy streams without encoding
}

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	bufSize = 4096
)

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

// WithInput sets the input format and format options for reading.
// The format parameter specifies the input format name (e.g., "s16le", "mp3").
// If format is empty, only the options are set without changing the format.
// Additional options are key=value pairs (e.g., "sample_rate=22050", "channels=1").
//
// Example:
//   WithInput("s16le", "sample_rate=22050", "channels=1", "sample_fmt=s16")
//   WithInput("", "analyzeduration=1000000") // Only set options
func WithInput(format string, options ...string) Opt {
	return func(o *opts) error {
		// Set input format if provided
		if format != "" {
			if iformat := ffmpeg.AVFormat_find_input_format(format); iformat != nil {
				o.iformat = iformat
			} else {
				return errors.New("invalid input format")
			}
		}
		// Set format options
		o.opts = append(o.opts, options...)
		return nil
	}
}

// Output format from name or url
func OptOutputFormat(name string) Opt {
	return func(o *opts) error {
		// By name
		if oformat := ffmpeg.AVFormat_guess_format(name, name, ""); oformat != nil {
			o.oformat = oformat
		} else {
			return errors.New("invalid output format")
		}
		return nil
	}
}

// New stream with parameters
func OptStream(stream int, par *Par) Opt {
	return func(o *opts) error {
		if par == nil {
			return errors.New("invalid parameters")
		}
		if stream == 0 {
			stream = len(o.streams) + 1
		}
		if _, exists := o.streams[stream]; exists {
			return errors.New("duplicate stream")
		}
		if stream < 0 {
			return errors.New("invalid stream")
		}
		o.streams[stream] = par
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

// Enable stream copying mode (remuxing without encoding)
func OptCopy() Opt {
	return func(o *opts) error {
		o.copy = true
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
