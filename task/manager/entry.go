package manager

import (
	"context"
	"sync"

	// Packages
	schema "github.com/mutablelogic/go-media/task/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// entry is a task tracked by a Manager, plus the state Manager needs to run,
// cancel, and report on it.
type entry struct {
	sync.Mutex
	task   schema.Task
	status schema.Status
	cancel context.CancelFunc
	done   chan struct{} // closed once the task's Run goroutine returns
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Cancel if it's currently running, and reports whether it
// did so. The caller must hold e's lock.
func (e *entry) Cancel() bool {
	if e.cancel == nil || !e.status.Finished.IsZero() {
		return false
	}
	e.status.Cancelled = true
	e.cancel()
	return true
}
