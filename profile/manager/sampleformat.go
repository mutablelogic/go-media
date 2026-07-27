package manager

import (
	"context"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	schema "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListSampleFormats returns all supported sample formats.
func (profile *Profile) ListSampleFormats(ctx context.Context, req schema.SampleFormatListRequest) (_ schema.SampleFormatList, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "ListSampleFormats",
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	var opaque uintptr
	result := make(schema.SampleFormatList, 0, 16)

	matches := func(sf *schema.SampleFormat) bool {
		if req.Name != nil && sf.Name != *req.Name {
			return false
		}
		if req.IsPlanar != nil && sf.IsPlanar != *req.IsPlanar {
			return false
		}
		return true
	}

	for {
		samplefmt := ff.AVUtil_next_sample_fmt(&opaque)
		if samplefmt == ff.AV_SAMPLE_FMT_NONE {
			break
		}
		if sampleformat := schema.NewSampleFormat(samplefmt); sampleformat != nil && matches(sampleformat) {
			result = append(result, *sampleformat)
		}
	}

	return result, nil
}
