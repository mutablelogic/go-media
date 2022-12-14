package media

import (
	// Packages

	multierror "github.com/hashicorp/go-multierror"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	//	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type media struct {
	ctx      *ffmpeg.AVFormatContext
	streams  map[int]Stream
	metadata *metadata
	cb       func(Media) error
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (media *media) new(ctx *ffmpeg.AVFormatContext, cb func(Media) error) error {
	media.ctx = ctx
	media.metadata = NewMetadata(ctx.Metadata())
	media.cb = cb
	media.streams = make(map[int]Stream, ctx.NumStreams())
	for _, stream := range ctx.Streams() {
		key := stream.Index()
		media.streams[key] = NewStream(stream)
	}
	// Return success
	return nil
}

func (media *media) Close(parent Media) error {
	var result error

	// Callback if required
	if media.cb != nil && parent != nil {
		if err := media.cb(parent); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Close input
	if media.ctx.Input() != nil {
		ffmpeg.AVFormat_close_input_ptr(media.ctx)
	} else {
		ffmpeg.AVFormat_free_context(media.ctx)
	}

	// Release resources
	media.cb = nil
	media.ctx = nil
	media.streams = nil
	media.metadata = nil

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (media *media) URL() string {
	if media.ctx == nil {
		return ""
	} else {
		return media.ctx.Url()
	}
}

func (media *media) Streams() []Stream {
	if media.ctx == nil {
		return nil
	}
	result := make([]Stream, 0, len(media.streams))
	for _, stream := range media.streams {
		result = append(result, stream)
	}
	return result
}

func (media *media) Metadata() Metadata {
	if media.ctx == nil {
		return nil
	} else {
		return media.metadata
	}
}

func (media *media) Flags() MediaFlag {
	flags := MEDIA_FLAG_NONE
	if media.ctx == nil {
		return flags
	}

	// Add flags from stream
	for _, stream := range media.Streams() {
		flags |= stream.Flags()
	}

	// Add flags from input
	if input := media.ctx.Input(); input != nil {
		flags := MEDIA_FLAG_DECODER
		if !input.Format().Is(ffmpeg.AVFMT_NOFILE) {
			flags |= MEDIA_FLAG_FILE
		}
	}
	// Add flags from output
	if output := media.ctx.Output(); output != nil {
		flags := MEDIA_FLAG_ENCODER
		if !output.Format().Is(ffmpeg.AVFMT_NOFILE) {
			flags |= MEDIA_FLAG_FILE
		}
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
		// TODO: Flags for TV episode, etc.
	}

	// Return flags
	return flags
}

func (media *media) Set(key MediaKey, value any) error {
	return media.metadata.Set(key, value)
}
