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

func (profile *Profile) ListCodecs(ctx context.Context, req schema.CodecListRequest) (_ *schema.CodecList, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "ListCodecs",
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	// Match helper
	matches := func(c *ff.AVCodec) bool {
		if !ff.AVCodec_is_encoder(c) {
			return false
		}
		if req.Type != nil && c.Type() != ff.AVMediaType(types.Value(req.Type)) {
			return false
		}
		return true
	}

	// Get the list of audio codecs, applying offset and limit as we iterate.
	// A limit of zero means return the count only.
	var opaque uintptr
	var result schema.CodecList
	for {
		codec := ff.AVCodec_iterate(&opaque)
		if codec == nil {
			break
		}
		if !matches(codec) {
			continue
		}
		result.Count += 1
		if result.Count <= req.Offset {
			continue
		}
		if req.Limit != nil && uint64(len(result.Body)) >= types.Value(req.Limit) {
			continue
		}
		result.Body = append(result.Body, schema.NewCodec(codec))
	}

	// Copy the request offset/limit into the result, then clamp the limit to
	// reflect the number of items actually available after the offset.
	result.CodecListRequest = req
	result.OffsetLimit.Clamp(result.Count)

	// Return success
	return types.Ptr(result), nil
}

func (profile *Profile) GetCodec(ctx context.Context, name string) (_ *schema.Codec, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "GetCodec",
		attribute.String("name", name),
	)
	defer func() { endSpan(err) }()

	// Get the codec by name
	codec := ff.AVCodec_find_encoder_by_name(name)
	if codec == nil {
		return nil, gomedia.ErrNotFound.Withf("codec %q is not found", name)
	} else if codec.IsEncoder() == false {
		return nil, gomedia.ErrBadParameter.Withf("codec %q is not an encoding codec", name)
	} else if codec.Type() != ff.AVMEDIA_TYPE_AUDIO && codec.Type() != ff.AVMEDIA_TYPE_VIDEO && codec.Type() != ff.AVMEDIA_TYPE_SUBTITLE {
		return nil, gomedia.ErrBadParameter.Withf("codec %q is not an audio, video or subtitle encoding codec", name)
	}

	// Return the codec
	return schema.NewCodec(codec), nil
}
