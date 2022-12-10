package media

import (
	"fmt"
	"os"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type input struct {
	ctx *ffmpeg.AVFormatContext
	cb  func(Media) error
}

// Ensure *input complies with Media interface
var _ Media = (*input)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewInputFile(path string, cb func(Media) error) (*input, error) {
	media := new(input)

	// Check for path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrNotFound.With(path)
	} else if err != nil {
		return nil, err
	}

	// Create a context - detect format
	if err := ffmpeg.AVFormat_open_input(&media.ctx, path, nil, nil); err != nil {
		return nil, err
	} else if media.ctx == nil {
		return nil, ErrInternalAppError.With("AVFormat_open_input")
	}

	// Find stream info
	if err := ffmpeg.AVFormat_find_stream_info(media.ctx, nil); err != nil {
		ffmpeg.AVFormat_close_input(&media.ctx)
		return nil, err
	}

	// Set close callback
	media.cb = cb

	// Return success
	return media, nil
}

func (media *input) Close() error {
	var result error

	// Callback
	if media.cb != nil {
		if err := media.cb(media); err != nil {
			result = multierror.Append(result, err)
		}
		media.cb = nil
	}

	// Close context
	if media.ctx != nil {
		ffmpeg.AVFormat_close_input(&media.ctx)
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (media *input) String() string {
	str := "<media.input"
	if media.ctx != nil {
		str += fmt.Sprint(" ctx=", media.ctx)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS
