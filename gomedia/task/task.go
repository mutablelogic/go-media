package task

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Task Context composes an underlying context.Context
type Context struct {
	context.Context

	// Otel Tracer (can be nil)
	Tracer trace.Tracer
}

type Task interface {
	// Run the task
	Run(ctx Context) error
}
