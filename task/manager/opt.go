package manager

import (
	// Packages
	client "github.com/mutablelogic/go-client"
	metadata "github.com/mutablelogic/go-media/metadata"
	trace "go.opentelemetry.io/otel/trace"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is a functional option for task manager configuration.
type Opt func(*opt) error

type opt struct {
	metaopts []metadata.Option
	tracer   trace.Tracer
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
		o.metaopts = append(o.metaopts, metadata.WithTracer(tracer))
		return nil
	}
}

// WithTMDB enables TMDB metadata extraction, using the given token and client options.
func WithTMDB(token string, clientopts ...client.ClientOpt) Opt {
	return func(o *opt) error {
		o.metaopts = append(o.metaopts, metadata.WithTMDB(token, clientopts...))
		return nil
	}
}
