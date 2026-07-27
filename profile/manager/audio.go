package manager

import (
	"context"
	"encoding/json"

	// Packages
	uuid "github.com/google/uuid"
	otel "github.com/mutablelogic/go-client/pkg/otel"
	schema "github.com/mutablelogic/go-media/profile/schema"
	pg "github.com/mutablelogic/go-pg"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (profile *Profile) CreateAudioProfile(ctx context.Context, req schema.AudioProfileMeta) (_ *schema.AudioProfile, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "CreateAudioProfile",
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	var result schema.AudioProfile
	if err := profile.Tx(ctx, func(conn pg.Conn) error {
		// Create the audio profile
		audioProfile, err := schema.NewAudioProfile(req.Name)
		if err != nil {
			return err
		}

		// Set options
		if req.Bitrate != nil {
			if err := audioProfile.Set(schema.OptionBitrate, types.Value(req.Bitrate)); err != nil {
				return err
			}
		}
		if req.SampleRate != nil {
			if err := audioProfile.Set(schema.OptionSampleRate, types.Value(req.SampleRate)); err != nil {
				return err
			}
		}
		if req.SampleFormat != nil {
			if err := audioProfile.Set(schema.OptionSampleFormat, types.Value(req.SampleFormat)); err != nil {
				return err
			}
		}
		if req.ChannelLayout != nil {
			if err := audioProfile.Set(schema.OptionChannelLayout, types.Value(req.ChannelLayout)); err != nil {
				return err
			}
		}

		// Unmarshal the options JSON into a map
		if req.Opts != nil {
			var opts map[string]any
			if err := json.Unmarshal(req.Opts, &opts); err != nil {
				return err
			}
			for name, value := range opts {
				if err := audioProfile.Set(name, value); err != nil {
					return err
				}
			}
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

func (profile *Profile) UpdateAudioProfile(ctx context.Context, uuid uuid.UUID, meta schema.AudioProfileMeta) (_ *schema.AudioProfile, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "UpdateAudioProfile",
		attribute.String("uuid", uuid.String()),
	)
	defer func() { endSpan(err) }()

	var result schema.AudioProfile
	// TODO: Set each option in the meta to the audio profile to validate it
	if err := profile.PoolConn.Update(ctx, &result, schema.AudioProfileUUID(uuid), meta); err != nil {
		return nil, pg.NormalizeError(err)
	}

	return types.Ptr(result), nil
}
