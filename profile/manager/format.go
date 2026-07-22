package manager

import (
	"context"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	gomedia "github.com/mutablelogic/go-media"
	schema "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (profile *Profile) ListFormats(ctx context.Context, req schema.FormatListRequest) (_ *schema.FormatList, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "ListFormats",
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	// Match helper
	matches := func(_ *ff.AVOutputFormat) bool {
		return true
	}

	// Get the list of output formats, applying offset and limit as we iterate.
	// A limit of zero means return the count only.
	var opaque uintptr
	var result schema.FormatList
	for {
		format := ff.AVFormat_muxer_iterate(&opaque)
		if format == nil {
			break
		}
		if !matches(format) {
			continue
		}
		result.Count += 1
		if result.Count <= req.Offset {
			continue
		}
		if req.Limit != nil && uint64(len(result.Body)) >= types.Value(req.Limit) {
			continue
		}
		result.Body = append(result.Body, schema.NewOutputFormat(format))
	}

	// Copy the request offset/limit into the result, then clamp the limit to
	// reflect the number of items actually available after the offset.
	result.FormatListRequest = req
	result.OffsetLimit.Clamp(result.Count)

	// Return success
	return types.Ptr(result), nil
}

func (profile *Profile) GetFormat(ctx context.Context, name string) (_ *schema.Format, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "GetFormat",
		attribute.String("name", name),
	)
	defer func() { endSpan(err) }()

	// Get the format by name
	format := ff.AVFormat_guess_format(name, "", "")
	if format == nil {
		return nil, gomedia.ErrNotFound.Withf("format %q is not found", name)
	}

	// Return the format
	return schema.NewOutputFormat(format), nil
}
