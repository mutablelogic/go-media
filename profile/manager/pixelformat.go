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

// ListPixelFormats returns all supported pixel formats.
func (profile *Profile) ListPixelFormats(ctx context.Context, req schema.PixelFormatListRequest) (_ schema.PixelFormatList, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "ListPixelFormats",
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	var opaque uintptr
	result := make(schema.PixelFormatList, 0, 256)

	matches := func(pf *schema.PixelFormat) bool {
		if req.Name != nil && pf.Name != *req.Name {
			return false
		}
		if req.NumPlanes != nil && uint64(pf.NumPlanes) != *req.NumPlanes {
			return false
		}
		return true
	}

	for {
		pixfmt := ff.AVUtil_next_pixel_fmt(&opaque)
		if pixfmt == ff.AV_PIX_FMT_NONE {
			break
		}
		if pixelformat := schema.NewPixelFormat(pixfmt); pixelformat != nil && matches(pixelformat) {
			result = append(result, *pixelformat)
		}
	}

	return result, nil
}
