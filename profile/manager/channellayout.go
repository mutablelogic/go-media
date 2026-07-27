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

// ListChannelLayouts returns standard FFmpeg channel layouts.
func (profile *Profile) ListChannelLayouts(ctx context.Context, req schema.ChannelLayoutListRequest) (_ schema.ChannelLayoutList, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "ListChannelLayouts",
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	var opaque uintptr
	result := make(schema.ChannelLayoutList, 0, 32)

	matches := func(cl *schema.ChannelLayout) bool {
		if req.Name != nil && cl.Name != *req.Name {
			return false
		}
		if req.NumChannels != nil && uint64(cl.NumChannels) != *req.NumChannels {
			return false
		}
		return true
	}

	for {
		layout := ff.AVUtil_channel_layout_standard(&opaque)
		if layout == nil {
			break
		}
		if channellayout := schema.NewChannelLayout(layout); channellayout != nil && matches(channellayout) {
			result = append(result, *channellayout)
		}
	}

	return result, nil
}
