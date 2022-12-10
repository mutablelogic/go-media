package media

import (
	"fmt"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type output struct {
	ctx *ffmpeg.AVFormatContext
	cb  func(Media) error
}

// Ensure *input complies with Media interface
var _ Media = (*output)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewOutputFile(path string, cb func(Media) error) (*output, error) {
	media := new(output)

	// Create a context - detect format
	if err := ffmpeg.AVFormat_alloc_output_context2(&media.ctx, nil, "", path); err != nil {
		return nil, err
	} else if media.ctx == nil {
		return nil, ErrInternalAppError.With("AVFormat_alloc_output_context2")
	}

	// Set close callback
	media.cb = cb

	// Return success
	return media, nil
}

func (media *output) Close() error {
	var result error

	// Callback
	if media.cb != nil {
		if err := media.cb(media); err != nil {
			result = multierror.Append(result, err)
		}
		media.cb = nil
	}

	// Close context - TODO

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (media *output) String() string {
	str := "<media.output"
	if media.ctx != nil {
		str += fmt.Sprint(" ctx=", media.ctx)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS
