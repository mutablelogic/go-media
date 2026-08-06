package manager

import (
	"context"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	gomedia "github.com/mutablelogic/go-media"
	taskencoder "github.com/mutablelogic/go-media/task/encoder"
	taskschema "github.com/mutablelogic/go-media/task/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Encode starts the encoder task on the new task manager (m.opt.taskManager)
// and returns immediately with its current status - encoding is long-running,
// so this doesn't wait for it to complete. The task stays tracked by the
// task manager afterwards (see task/httphandler), for a caller to poll,
// stream events for, or cancel by its UUID.
func (m *Media) Encode(ctx context.Context, req taskencoder.EncodeRequest) (_ *taskschema.Status, err error) {
	ctx, endSpan := otel.StartSpan(m.opt.tracer, ctx, "gomedia.encode",
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	if m.opt.taskManager == nil {
		return nil, gomedia.ErrNotImplemented.With("no task manager configured")
	}

	// Create a new task in the task manager and start it. Add validates the
	// task itself, so a malformed request is rejected here rather than only
	// once the caller polls or watches the task. Unlike Metadata, the task
	// isn't waited on or removed here - it's left running and tracked by
	// the task manager.
	id, err := m.opt.taskManager.Add(ctx, "encoder", &req)
	if err != nil {
		return nil, err
	}
	if err := m.opt.taskManager.Start(ctx, id); err != nil {
		return nil, err
	}

	return m.opt.taskManager.GetTask(ctx, id)
}
