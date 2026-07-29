package schema

import (
	"context"
	"net/url"
	"strconv"

	// Packages
	pg "github.com/mutablelogic/go-pg"
	types "github.com/mutablelogic/go-server/pkg/types"
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
	// Task returns the task's own kind, e.g. "metadata" - fixed by the
	// implementation, unlike the caller-chosen name passed to Manager.Add.
	Task() string

	// Validate reports whether the task is well-formed and can be run at
	// all (e.g. a non-nil reader). Manager.Add calls this before the task
	// is even registered, so a malformed request is rejected immediately
	// rather than accepted as a task that's already doomed to fail once
	// started.
	Validate() error

	// Run the task
	Run(ctx Context) error
}

// TaskListRequest filters the tasks returned by Manager.ListTasks. A nil
// State means "don't filter by state".
type TaskListRequest struct {
	State *State `json:"state,omitempty" help:"Filter by lifecycle state." enum:"not_started,running,cancelled,error,done" example:"running"`
	pg.OffsetLimit
}

// TaskList is a snapshot of the tasks tracked by a Manager matching a
// TaskListRequest, in the order they were added.
type TaskList struct {
	TaskListRequest
	Count uint64   `json:"count" help:"Number of tasks matching the request." example:"1"`
	Body  []Status `json:"body,omitempty" help:"List of tasks."`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

func (r TaskListRequest) Query() url.Values {
	query := url.Values{}
	if r.State != nil {
		query.Set("state", string(*r.State))
	}
	if r.Offset > 0 {
		query.Set("offset", strconv.FormatUint(r.Offset, 10))
	}
	if r.Limit != nil {
		query.Set("limit", strconv.FormatUint(types.Value(r.Limit), 10))
	}
	return query
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t TaskList) String() string {
	return types.Stringify(t)
}
