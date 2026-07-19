package manager

import (
	// Packages
	"strings"

	"github.com/hala-systems/fabric-service/pkg/types"
	gomedia "github.com/mutablelogic/go-media"
	schema "github.com/mutablelogic/go-media/profile/schema"
	trace "go.opentelemetry.io/otel/trace"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is a functional option for filer manager configuration.
type Opt func(*opt) error

type opt struct {
	tracer trace.Tracer
	schema string
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (o *opt) apply(opt []Opt) error {
	o.defaults()

	// Apply options
	for _, fn := range opt {
		if err := fn(o); err != nil {
			return err
		}
	}

	// Return success
	return nil
}

func (o *opt) defaults() {
	o.schema = schema.DefaultSchema
}

////////////////////////////////////////////////////////////////////////////////
// OPTIONS

// WithTracer sets the tracer used for tracing operations.
func WithTracer(tracer trace.Tracer) Opt {
	return func(o *opt) error {
		o.tracer = tracer
		return nil
	}
}

// WithSchema sets the database schema used for profile persistence
func WithSchema(schema string) Opt {
	return func(o *opt) error {
		if schema = strings.TrimSpace(schema); !types.IsIdentifier(schema) {
			return gomedia.ErrBadParameter.Withf("invalid schema name %q", schema)
		} else {
			o.schema = schema
		}

		// Return success
		return nil
	}
}
