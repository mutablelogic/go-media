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

// Cancel marks e cancelled if it's currently running, and returns the
// context.CancelFunc to actually stop it, or nil if there's nothing to
// cancel (never started, or already finished). The caller must hold e's
// lock for this call, but must call the returned func only after releasing
// it - and, if a Cancelled event is going to be emitted, only after that
// emit has completed. Calling it cancels the task's own ctx, which
// synchronously unblocks its Run goroutine; that goroutine then races to
// re-acquire e's lock for its own Finished bookkeeping, so calling the
// func too early can let a Finished event overtake a Cancelled event that
// logically preceded it.
func (e *entry) Cancel() (context.CancelFunc, bool) {
	if e.cancel == nil || !e.status.Finished.IsZero() {
		return nil, false
	}
	e.status.Cancelled = true
	return e.cancel, true
}
