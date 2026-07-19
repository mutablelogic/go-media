package manager

import (
	"context"
	"net/url"

	// Packages
	uuid "github.com/google/uuid"
	otel "github.com/mutablelogic/go-client/pkg/otel"
	schema "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	pg "github.com/mutablelogic/go-pg"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (profile *Profile) ListAudioCodecs(ctx context.Context) (_ *schema.AudioCodecList, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "ListAudioCodecs")
	defer func() { endSpan(err) }()

	// Match helper
	matches := func(c *ff.AVCodec) bool {
		if !ff.AVCodec_is_encoder(c) {
			return false
		}
		if c.Type() != ff.AVMEDIA_TYPE_AUDIO {
			return false
		}
		return true
	}

	// Get the list of audio codecs
	var opaque uintptr
	var result schema.AudioCodecList
	for {
		codec := ff.AVCodec_iterate(&opaque)
		if codec == nil {
			break
		}
		if !matches(codec) {
			continue
		}
		result.Body = append(result.Body, schema.AudioCodec{
			Name:        codec.Name(),
			Description: codec.LongName(),
		})
	}
	result.Count = uint64(len(result.Body))

	// Return success
	return types.Ptr(result), nil
}

func (profile *Profile) CreateAudioProfile(ctx context.Context, codec string, opts url.Values) (_ *schema.AudioProfile, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "CreateAudioProfile",
		attribute.String("codec", codec),
		attribute.String("opts", opts.Encode()),
	)
	defer func() { endSpan(err) }()

	var result schema.AudioProfile
	if err := profile.Tx(ctx, func(conn pg.Conn) error {
		// Create the audio profile
		audioProfile, err := schema.NewAudioProfile(codec)
		if err != nil {
			return err
		}

		// Apply options from the URL values
		if err := audioProfile.Set(opts); err != nil {
			return err
		}

		// Insert the audio profile into the database
		if err := conn.Insert(ctx, &result, audioProfile); err != nil {
			return err
		}

		// Return success
		return nil
	}); err != nil {
		return nil, pg.NormalizeError(err)
	}

	return types.Ptr(result), nil
}

func (profile *Profile) GetAudioProfile(ctx context.Context, uuid uuid.UUID) (_ *schema.AudioProfile, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "GetAudioProfile",
		attribute.String("uuid", uuid.String()),
	)
	defer func() { endSpan(err) }()

	var result schema.AudioProfile
	if err := profile.PoolConn.Get(ctx, &result, schema.AudioProfileUUID(uuid)); err != nil {
		return nil, pg.NormalizeError(err)
	}

	return types.Ptr(result), nil
}

func (profile *Profile) DeleteAudioProfile(ctx context.Context, uuid uuid.UUID) (_ *schema.AudioProfile, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "DeleteAudioProfile",
		attribute.String("uuid", uuid.String()),
	)
	defer func() { endSpan(err) }()

	var result schema.AudioProfile
	if err := profile.PoolConn.Delete(ctx, &result, schema.AudioProfileUUID(uuid)); err != nil {
		return nil, pg.NormalizeError(err)
	}

	return types.Ptr(result), nil
}
