package schema

import (
	"context"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Task Context composes an underlying context.Context
type Context struct {
	context.Context

	// Progress reports how far the task has got, in task-defined units (e.g.
	// bytes, frames, streams) - total is 0 if not known in advance.
	Progress func(current, total uint64)

	// Result sets the task's output, retrievable afterwards via the
	// Manager's Status.
	Result func(any)
}

type Task interface {
	// Run the task
	Run(ctx Context) error
}
