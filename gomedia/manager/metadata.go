package manager

import (
	"context"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	gomedia "github.com/mutablelogic/go-media"
	task "github.com/mutablelogic/go-media/task/metadata"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Metadata runs the metadata-extraction task on the new task manager
// (m.opt.taskManager) and waits for it to complete.
func (m *Media) Metadata(ctx context.Context, req task.MetadataRequest) (_ *task.MetadataResponse, err error) {
	ctx, endSpan := otel.StartSpan(m.opt.tracer, ctx, "gomedia.metadata",
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	// Check that we have a task manager configured
	if m.opt.taskManager == nil {
		return nil, gomedia.ErrNotImplemented.With("no task manager configured")
	}

	// Determine the name of the input, if known (e.g. the uploaded filename).
	var name string
	if named, ok := req.Reader.(gomedia.NamedReader); ok && named != nil {
		name = named.Name()
	}

	// Create a new task in the task manager and start it. The task manager will
	// run the task asynchronously, and we will wait for it to complete, then
	// we'll delete the task from the manager.
	id, err := m.opt.taskManager.Add(ctx, name, &req)
	if err != nil {
		return nil, err
	} else {
		defer m.opt.taskManager.Remove(ctx, id)
	}

	// Start the task
	if err := m.opt.taskManager.Start(ctx, id); err != nil {
		return nil, err
	}

	// Wait for task to complete. A task can set a partial result and still return an error.
	if status, err := m.opt.taskManager.Wait(ctx, id); status != nil {
		if result, ok := status.Result.(*task.MetadataResponse); ok && result != nil {
			return result, nil
		} else {
			return nil, gomedia.ErrInternalError.With("metadata task returned an unexpected result type")
		}
	} else {
		return nil, err
	}
}
