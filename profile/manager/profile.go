package manager

import (
	"context"
	"encoding/json"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	schema "github.com/mutablelogic/go-media/profile/schema"
	pg "github.com/mutablelogic/go-pg"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (profile *Profile) CreateProfile(ctx context.Context, name string, req schema.ProfileMeta) (_ *schema.Profile, err error) {
	ctx, endSpan := otel.StartSpan(profile.tracer, ctx, "CreateProfile",
		attribute.String("name", name),
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	var result schema.Profile
	if err := profile.Tx(ctx, func(conn pg.Conn) error {
		// Create the profile
		profile, err := schema.NewProfile(name, req)
		if err != nil {
			return err
		}

		// Unmarshal the options JSON into a map
		if req.Opts != nil {
			var opts map[string]any
			if err := json.Unmarshal(req.Opts, &opts); err != nil {
				return err
			}
			for name, value := range opts {
				if err := profile.Set(name, value); err != nil {
					return err
				}
			}
		}

		// Insert the profile into the database
		if err := conn.Insert(ctx, &result, profile); err != nil {
			return err
		}

		// Return success
		return nil
	}); err != nil {
		return nil, pg.NormalizeError(err)
	}

	return types.Ptr(result), nil
}
