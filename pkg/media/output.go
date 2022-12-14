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
	media
}

// Ensure *input complies with Media interface
var _ Media = (*output)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewOutputFile(path string, cb func(Media) error) (*output, error) {
	media := new(output)

	// Create a context - detect format
	ctx, err := ffmpeg.AVFormat_alloc_output_context2(nil, "", path)
	if err != nil {
		return nil, err
	}

	// Open the actual file if required
	if !ctx.Output().Format().Is(ffmpeg.AVFMT_NOFILE) {
		if ioctx, err := ffmpeg.AVFormat_avio_open(path, ffmpeg.AVIO_FLAG_WRITE); err != nil {
			ffmpeg.AVFormat_free_context(ctx)
			return nil, err
		} else {
			ctx.SetPB(ioctx)
		}
	}

	// Initialize the media
	if err := media.new(ctx, cb); err != nil {
		if ctx.PB() != nil {
			ffmpeg.AVFormat_avio_close(ctx.PB())
		}
		ffmpeg.AVFormat_free_context(ctx)
		return nil, err
	}

	// Return success
	return media, nil
}

func NewOutputDevice(device MediaFormat, cb func(Media) error) (*output, error) {
	media := new(output)
	format, ok := device.(*format_out)
	if !ok || format == nil || format.ctx == nil {
		return nil, ErrBadParameter.With("device")
	}

	// Create a context - detect format
	ctx, err := ffmpeg.AVFormat_alloc_output_context2(format.ctx, "", "")
	if err != nil {
		return nil, err
	}

	// Initialize the media
	if err := media.new(ctx, cb); err != nil {
		if ctx.PB() != nil {
			ffmpeg.AVFormat_avio_close(ctx.PB())
		}
		ffmpeg.AVFormat_free_context(ctx)
		return nil, err
	}

	// Return success
	return media, nil
}

func (output *output) Close() error {
	var result error

	// Close output
	if output.ctx != nil {
		if ioctx := output.ctx.PB(); ioctx != nil {
			if err := ffmpeg.AVFormat_avio_close(ioctx); err != nil {
				result = multierror.Append(result, err)
			}
		}
	}

	// Callback with media
	if err := output.media.Close(output); err != nil {
		result = multierror.Append(result, err)
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (media *output) String() string {
	str := "<media.output"
	if url := media.URL(); url != "" {
		str += fmt.Sprintf(" url=%q", url)
	}
	if flags := media.Flags(); flags != MEDIA_FLAG_NONE {
		str += fmt.Sprint(" flags=", flags)
	}
	if len(media.streams) > 0 {
		str += fmt.Sprint(" streams=", media.streams)
	}
	if media.metadata != nil {
		str += fmt.Sprint(" metadata=", media.metadata)
	}
	return str + ">"
}
