package manager

import (
	// Packages
	trace "go.opentelemetry.io/otel/trace"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is a functional option for task manager configuration.
type Opt func(*opt) error

type opt struct {
	tracer trace.Tracer
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

func (o *opt) defaults() {}

////////////////////////////////////////////////////////////////////////////////
// OPTIONS

// WithTracer sets the tracer used for tracing operations.
func WithTracer(tracer trace.Tracer) Opt {
	return func(o *opt) error {
		o.tracer = tracer
		return nil
	}
}
