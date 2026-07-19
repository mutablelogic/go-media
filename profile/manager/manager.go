package manager

import (
	"context"
	"strings"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	gomedia "github.com/mutablelogic/go-media"
	schema "github.com/mutablelogic/go-media/profile/schema"
	pg "github.com/mutablelogic/go-pg"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Profile struct {
	opt
	pg.PoolConn
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new profile manger
func New(ctx context.Context, pool pg.PoolConn, opts ...Opt) (_ *Profile, err error) {
	self := new(Profile)

	// Apply and set options
	if err := self.apply(opts); err != nil {
		return nil, err
	} else {
		self.PoolConn = pool
	}

	// Parse and register named queries so bind.Query(...) can resolve them.
	queries, err := pg.NewQueries(strings.NewReader(schema.Queries))
	if err != nil {
		return nil, gomedia.ErrInternalError.Withf("parse queries.sql: %v", err.Error())
	} else if pool == nil {
		return nil, gomedia.ErrInternalError.With("pg pool is required")
	} else {
		pool = pool.WithQueries(queries).With("schema", self.schema).(pg.PoolConn)
	}

	// Create objects in the database schema
	bootstrapCtx, endBootstrapSpan := otel.StartSpan(self.tracer, ctx, "gomedia.bootstrap",
		attribute.String("schema", self.schema),
	)
	defer func() {
		endBootstrapSpan(err)
	}()

	if err := bootstrap(bootstrapCtx, pool, self.schema); err != nil {
		return nil, err
	} else {
		self.PoolConn = pool
	}

	// Return the profile manager
	return self, nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func bootstrap(ctx context.Context, conn pg.Conn, schemaName string) error {
	// Get all objects
	objects, err := pg.NewQueries(strings.NewReader(schema.Objects))
	if err != nil {
		return err
	}

	// Create the schema
	if err := pg.SchemaCreate(ctx, conn, schemaName); err != nil {
		return err
	}

	// Create all objects - not in a transaction
	for _, key := range objects.Keys() {
		if err := conn.Exec(ctx, objects.Query(key)); err != nil {
			return err
		}
	}

	// Return success
	return nil
}
