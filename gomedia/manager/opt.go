package manager

import (
	// Packages
	client "github.com/mutablelogic/go-client"
	chromaprint "github.com/mutablelogic/go-media/pkg/chromaprint"
	trace "go.opentelemetry.io/otel/trace"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is a functional option for filer manager configuration.
type Opt func(*opt) error

type opt struct {
	tracer         trace.Tracer
	acoustIDClient *chromaprint.Client
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

// WithAcoustIDKey creates and stores an AcoustID lookup client.
func WithAcoustIDKey(key string, opts ...client.ClientOpt) Opt {
	return func(o *opt) error {
		c, err := chromaprint.NewClient(key, opts...)
		if err != nil {
			return err
		} else {
			o.acoustIDClient = c
		}
		return nil
	}
}
