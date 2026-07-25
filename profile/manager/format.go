package manager

import (
	"context"
	"strings"

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

	// Match helper. Name/MimeTypes/Extensions are each a single string that
	// FFmpeg comma-joins when a format has more than one (e.g. mp4's
	// extensions are "mp4,m4a,m4v"), so filters match against any one token
	// rather than the field as a whole.
	matches := func(name, mimeTypes, extensions string, isInput, isOutput bool) bool {
		if req.Name != nil && !hasToken(name, types.Value(req.Name)) {
			return false
		}
		if req.Type != nil && !hasToken(mimeTypes, types.Value(req.Type)) {
			return false
		}
		if req.Ext != nil && !hasToken(extensions, types.Value(req.Ext)) {
			return false
		}
		if req.IsInput != nil && isInput != types.Value(req.IsInput) {
			return false
		}
		if req.IsOutput != nil && isOutput != types.Value(req.IsOutput) {
			return false
		}
		return true
	}

	// Get the list of input and output formats, applying offset and limit as
	// we iterate. A limit of zero means return the count only.
	var result schema.FormatList
	add := func(f *schema.Format) {
		if f == nil {
			return
		}
		result.Count += 1
		if result.Count <= req.Offset {
			return
		}
		if req.Limit != nil && uint64(len(result.Body)) >= types.Value(req.Limit) {
			return
		}
		result.Body = append(result.Body, f)
	}

	var inOpaque uintptr
	for {
		format := ff.AVFormat_demuxer_iterate(&inOpaque)
		if format == nil {
			break
		}
		if !matches(format.Name(), format.MimeTypes(), format.Extensions(), true, false) {
			continue
		}
		add(schema.NewInputFormat(format))
	}

	var outOpaque uintptr
	for {
		format := ff.AVFormat_muxer_iterate(&outOpaque)
		if format == nil {
			break
		}
		if !matches(format.Name(), format.MimeTypes(), format.Extensions(), false, true) {
			continue
		}
		add(schema.NewOutputFormat(format))
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

	// Get the format by name - prefer an output (muxer) match, since this is
	// primarily used to look up an encoding target, then fall back to an
	// input (demuxer) match.
	if format := ff.AVFormat_guess_format(name, "", ""); format != nil {
		return schema.NewOutputFormat(format), nil
	}
	if format := ff.AVFormat_find_input_format(name); format != nil {
		return schema.NewInputFormat(format), nil
	}

	return nil, gomedia.ErrNotFound.Withf("format %q is not found", name)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// hasToken reports whether value equals (case-insensitively) one of the
// comma-separated tokens in field.
func hasToken(field, value string) bool {
	for _, token := range strings.Split(field, ",") {
		if strings.EqualFold(strings.TrimSpace(token), value) {
			return true
		}
	}
	return false
}
