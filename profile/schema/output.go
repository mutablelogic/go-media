package schema

import (
	"encoding/json"
	"net/url"

	// Packages
	uuid "github.com/google/uuid"
	gomedia "github.com/mutablelogic/go-media"
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
// MARSHALING

// UnmarshalJSON is required because ctx is unexported (see OutputWithName) -
// without it, a client decoding an Output from JSON would get one with a
// valid Format but a nil Context(), which every consumer (e.g. writer.Create)
// treats as "no such format". Resolves ctx from the decoded Format, exactly
// as OutputWithName does.
func (o *Output) UnmarshalJSON(data []byte) error {
	type alias Output
	aux := (*alias)(o)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	ctx := ff.AVFormat_guess_format(o.Format, "", "")
	if ctx == nil {
		return gomedia.ErrBadParameter.Withf("output format %q is not found", o.Format)
	}
	o.ctx = ctx

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (o *OutputMeta) Context() *ff.AVOutputFormat {
	return o.ctx
}
