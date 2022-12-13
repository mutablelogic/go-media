package media

import (
	"fmt"
	"os"

	// Packages

	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type input struct {
	media
}

// Ensure *input complies with Media interface
var _ Media = (*input)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewInputFile(path string, cb func(Media) error) (*input, error) {
	this := new(input)

	// Check for path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrNotFound.With(path)
	} else if err != nil {
		return nil, err
	}

	// Create a context - detect format
	ctx, err := ffmpeg.AVFormat_open_input(path, nil, nil)
	if err != nil {
		return nil, err
	}

	// Find stream info
	if err := ffmpeg.AVFormat_find_stream_info(ctx, nil); err != nil {
		ffmpeg.AVFormat_close_input_ptr(ctx)
		return nil, err
	}

	// Initialize the media
	if err := this.new(ctx, cb); err != nil {
		ffmpeg.AVFormat_close_input_ptr(ctx)
		return nil, err
	}

	// Return success
	return this, nil
}

func (input *input) Close() error {
	return input.media.Close(input)
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (media *input) String() string {
	str := "<media.input"
	if url := media.URL(); url != "" {
		str += fmt.Sprintf(" url=%q", url)
	}
	if flags := media.Flags(); flags != MEDIA_FLAG_NONE {
		str += fmt.Sprint(" flags=", flags)
	}
	if len(media.streams) > 0 {
		for key, stream := range media.streams {
			str += fmt.Sprintf(" stream_%d=%v", key, stream)
		}
	}
	if media.metadata != nil {
		str += fmt.Sprint(" metadata=", media.metadata)
	}
	return str + ">"
}
