package media

import (
	"fmt"
	"net/url"
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

func NewInputFile(path string, format MediaFormat, cb func(Media) error) (*input, error) {
	// Check for path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, ErrNotFound.Withf("%q", path)
	} else if err != nil {
		return nil, err
	} else {
		return newInput(path, format, cb)
	}
}

func NewInputDevice(device MediaFormat, cb func(Media) error) (*input, error) {
	format, ok := device.(*format_in)
	if !ok || format == nil || format.ctx == nil {
		return nil, ErrBadParameter.With("device")
	} else {
		return newInput("", device, cb)
	}
}

func NewInputURL(path string, format MediaFormat, cb func(Media) error) (*input, error) {
	url, err := url.Parse(path)
	if err != nil {
		return nil, ErrBadParameter.With(path)
	} else {
		return newInput(url.String(), format, cb)
	}
}

func newInput(path string, format MediaFormat, cb func(Media) error) (*input, error) {
	this := new(input)

	// Create a context - detect format or use format argument
	var format_ctx *ffmpeg.AVInputFormat
	if format_in, ok := format.(*format_in); ok && format_in != nil {
		format_ctx = format_in.ctx
	}
	ctx, err := ffmpeg.AVFormat_open_input(path, format_ctx, nil)
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
