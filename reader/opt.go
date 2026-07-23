package reader

import (
	// Packages
	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type opts struct {
	input   *ff.AVInputFormat
	options []string // These are key=value pairs
}

type Opt func(o *opts) error

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (o *opts) apply(opts ...Opt) error {
	o.options = make([]string, 0, len(opts))
	for _, optFunc := range opts {
		if err := optFunc(o); err != nil {
			return err
		}
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// WithInput sets the input format and format options for reading.
// The format parameter specifies the input format name (e.g., "s16le", "mp3").
// If format is empty, only the options are set without changing the format.
// Additional options are key=value pairs (e.g., "sample_rate=22050", "channels=1").
//
// Example:
//
//	WithInput("s16le", "sample_rate=22050", "channels=1", "sample_fmt=s16")
//	WithInput("", "analyzeduration=1000000") // Only set options
func WithInput(format string, options ...string) Opt {
	return func(o *opts) error {
		// Set input format if provided
		if format != "" {
			if iformat := ff.AVFormat_find_input_format(format); iformat != nil {
				o.input = iformat
			} else {
				return gomedia.ErrBadParameter.Withf("invalid input format %q", format)
			}
		}
		// Append format options
		o.options = append(o.options, options...)
		return nil
	}
}
