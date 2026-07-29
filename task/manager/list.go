package manager

import (
	"context"
	"slices"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	schema "github.com/mutablelogic/go-media/task/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListTasks returns a snapshot of the tasks matching req, in the order they
// were added.
func (m *Manager) ListTasks(ctx context.Context, req schema.TaskListRequest) (_ *schema.TaskList, err error) {
	_, endSpan := otel.StartSpan(m.tracer, ctx, "ListTasks",
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	m.Lock()
	order := slices.Clone(m.order)
	m.Unlock()

	var result schema.TaskList
	for _, id := range order {
		m.Lock()
		e, exists := m.tasks[id]
		m.Unlock()
		if !exists {
			continue
		}

		e.Lock()
		status := e.status
		e.Unlock()

		if req.State != nil && status.State() != *req.State {
			continue
		}

		result.Count += 1
		if result.Count <= req.Offset {
			continue
		}
		if req.Limit != nil && uint64(len(result.Body)) >= types.Value(req.Limit) {
			continue
		}
		result.Body = append(result.Body, status)
	}

	// Copy the request offset/limit into the result, then clamp the limit to
	// reflect the number of items actually available after the offset.
	result.TaskListRequest = req
	result.OffsetLimit.Clamp(result.Count)

	// Return success
	return types.Ptr(result), nil
}
