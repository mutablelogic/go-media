package media

import (
	"fmt"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	//. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type output struct {
	ctx      *ffmpeg.AVFormatContext
	cb       func(Media) error
	streams  []Stream
	metadata Metadata
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

	// Set properties
	media.cb = cb
	media.ctx = ctx
	media.metadata = NewMetadata(media.ctx.Metadata())

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

	// Close context
	if ctx := media.ctx; ctx != nil {
		if ioctx := ctx.PB(); ioctx != nil {
			if err := ffmpeg.AVFormat_avio_close(ioctx); err != nil {
				result = multierror.Append(result, err)
			}
		}
		ffmpeg.AVFormat_free_context(ctx)
	}

	// Release resources
	media.ctx = nil
	media.streams = nil
	media.metadata = nil

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
	if media.ctx != nil {
		str += fmt.Sprint(" ctx=", media.ctx)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (media *output) URL() string {
	if media.ctx == nil {
		return ""
	} else {
		return media.ctx.Url()
	}
}

func (media *output) Streams() []Stream {
	if media.ctx == nil {
		return nil
	} else {
		return media.streams
	}
}

func (media *output) Metadata() Metadata {
	if media.ctx == nil {
		return nil
	} else {
		return media.metadata
	}
}

func (media *output) Flags() MediaFlag {
	if media.ctx == nil {
		return MEDIA_FLAG_NONE
	}
	flags := MEDIA_FLAG_ENCODER

	// Check for file flag
	if !media.ctx.Output().Format().Is(ffmpeg.AVFMT_NOFILE) {
		flags |= MEDIA_FLAG_FILE
	}

	// Append streams
	for _, stream := range media.Streams() {
		flags |= stream.Flags()
	}

	// Add other flags with likely media file type
	metadata := media.Metadata()
	if metadata != nil {
		if flags&MEDIA_FLAG_AUDIO != 0 && metadata.Value(MEDIA_KEY_ALBUM) != nil {
			flags |= MEDIA_FLAG_ALBUM
		}
		if flags&MEDIA_FLAG_ALBUM != 0 && metadata.Value(MEDIA_KEY_ALBUM_ARTIST) != nil && metadata.Value(MEDIA_KEY_TITLE) != nil {
			flags |= MEDIA_FLAG_ALBUM_TRACK
		}
		if flags&MEDIA_FLAG_ALBUM != 0 {
			if compilation, ok := metadata.Value(MEDIA_KEY_COMPILATION).(bool); ok && compilation {
				flags |= MEDIA_FLAG_ALBUM_COMPILATION
			}
		}
	}

	return flags
}
