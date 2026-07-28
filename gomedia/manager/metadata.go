package manager

import (
	"context"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	gomedia "github.com/mutablelogic/go-media"
	taskmetadata "github.com/mutablelogic/go-media/task/metadata"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Metadata runs the metadata-extraction task on the new task manager
// (m.opt.taskManager) and waits for it to complete.
func (m *Media) Metadata(ctx context.Context, req taskmetadata.MetadataRequest) (_ *taskmetadata.MetadataResponse, err error) {
	ctx, endSpan := otel.StartSpan(m.opt.tracer, ctx, "Metadata",
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

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
		// Best-effort cleanup
		defer m.opt.taskManager.Remove(ctx, id)
	}
	if err := m.opt.taskManager.Start(ctx, id); err != nil {
		return nil, err
	}
	status, waitErr := m.opt.taskManager.Wait(ctx, id)

	// A task can set a partial result and still return an error (e.g. the
	// content type and basic tags were extracted fine, but a later pass -
	// embedded artwork, say - failed). Prefer returning whatever result it
	// did produce over discarding it just because waitErr is non-nil; only
	// treat waitErr as fatal if there's no usable result to fall back on.
	if status != nil {
		if result, ok := status.Result.(*taskmetadata.MetadataResponse); ok && result != nil {
			return result, nil
		}
	}

	if waitErr != nil {
		return nil, waitErr
	}

	return nil, gomedia.ErrInternalError.With("metadata task returned an unexpected result type")
}
