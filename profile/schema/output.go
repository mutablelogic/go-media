package schema

import (
	"encoding/json"
	"net/url"

	// Packages
	uuid "github.com/google/uuid"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type OutputMeta struct {
	Description string             `json:"description,omitempty"` // Format description
	Opts        json.RawMessage    `json:"options,omitempty"`     // Additional format options
	ctx         *ff.AVOutputFormat `json:"-"`                     // Internal format
	opts        map[string]Option  `json:"-"`                     // Internal format options
}

type Output struct {
	Id     uuid.UUID `json:"id,omitempty"` // Unique identifier for the format profile
	Format string    `json:"format"`       // Format name, e.g. "mp4", "mkv", "flv", ...
	OutputMeta
}

type OutputUUID uuid.UUID

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func OutputWithName(format string, opt ...Option) *Output {
	ctx := ff.AVFormat_guess_format(format, "", "")
	if ctx == nil {
		return nil
	}
	return &Output{
		Format: format,
		OutputMeta: OutputMeta{
			ctx: ctx,
		},
	}
}

func OutputWithURL(url *url.URL, opt ...Option) *Output {
	if url == nil {
		return nil
	}
	ctx := ff.AVFormat_guess_format("", url.String(), "")
	if ctx == nil {
		return nil
	}
	return &Output{
		Format: ctx.Name(),
		OutputMeta: OutputMeta{
			ctx: ctx,
		},
	}
}

func OutputWithType(contenttype string, opt ...Option) *Output {
	ctx := ff.AVFormat_guess_format("", "", contenttype)
	if ctx == nil {
		return nil
	}
	return &Output{
		Format: ctx.Name(),
		OutputMeta: OutputMeta{
			ctx: ctx,
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (o *OutputMeta) Context() *ff.AVOutputFormat {
	return o.ctx
}
