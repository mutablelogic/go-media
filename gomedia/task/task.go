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

	// Progress reports how far the task has got, in task-defined units (e.g.
	// bytes, frames, streams) - total is 0 if not known in advance. Can be
	// nil; a task should treat a nil Progress as "nobody's listening" and
	// skip reporting.
	Progress func(current, total int64)

	// Result sets the task's output, retrievable afterwards via the
	// Manager's Status. Can be nil; a task should treat a nil Result as
	// "nobody's listening" and skip reporting.
	Result func(any)
}

type Task interface {
	// Run the task
	Run(ctx Context) error
}
