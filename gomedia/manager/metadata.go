package manager

import (
	"context"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	gomedia "github.com/mutablelogic/go-media"
	task "github.com/mutablelogic/go-media/gomedia/task"
	types "github.com/mutablelogic/go-server/pkg/types"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (m *Media) Metadata(ctx context.Context, req task.MetadataRequest) (_ *task.MetadataResponse, err error) {
	ctx, endSpan := otel.StartSpan(m.opt.tracer, ctx, "Metadata",
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	// Run the task and wait for it to complete, returning the result.
	if id, err := m.tasks.Add(ctx, "metadata", types.Ptr(req)); err != nil {
		return nil, err
	} else if err := m.tasks.Run(ctx, id); err != nil {
		return nil, err
	} else if status, err := m.tasks.Wait(ctx, id); err != nil {
		return nil, err
	} else if result, ok := status.Result.(*task.MetadataResponse); !ok || result == nil {
		return nil, gomedia.ErrInternalError.With("metadata task returned an unexpected result type")
	} else {
		return result, nil
	}
}
