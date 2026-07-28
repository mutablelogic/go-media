package manager

import (
	// Packages
	client "github.com/mutablelogic/go-client"
	chromaprint "github.com/mutablelogic/go-media/pkg/chromaprint"
	profilemanager "github.com/mutablelogic/go-media/profile/manager"
	taskmanager "github.com/mutablelogic/go-media/task/manager"
	trace "go.opentelemetry.io/otel/trace"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is a functional option for filer manager configuration.
type Opt func(*opt) error

type opt struct {
	tracer         trace.Tracer
	chromaprint    *chromaprint.Client
	profileManager *profilemanager.Profile
	taskManager    *taskmanager.Manager
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

// WithChromaprintKey creates and stores a Chromaprint client.
func WithChromaprintKey(key string, opts ...client.ClientOpt) Opt {
	return func(o *opt) error {
		c, err := chromaprint.NewClient(key, opts...)
		if err != nil {
			return err
		} else {
			o.chromaprint = c
		}
		return nil
	}
}

// WithProfileManager sets the profile manager used to resolve encoding
// profiles.
func WithProfileManager(profiles *profilemanager.Profile) Opt {
	return func(o *opt) error {
		o.profileManager = profiles
		return nil
	}
}

// WithTaskManager sets the task manager used to track tasks.
func WithTaskManager(tasks *taskmanager.Manager) Opt {
	return func(o *opt) error {
		o.taskManager = tasks
		return nil
	}
}
