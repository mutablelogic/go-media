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

// Probe a media stream from any reader and return information about its
// container format and streams. The probe runs as a task tracked by the
// Media's task manager, so it's cancelled along with any other running task
// if Run's context is cancelled while the probe is in flight.
func (m *Media) Probe(ctx context.Context, req task.ProbeRequest) (_ *task.ProbeResponse, err error) {
	ctx, endSpan := otel.StartSpan(m.opt.tracer, ctx, "Probe",
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	// Create a task to probe the input stream.
	t, err := task.NewProbeTask(req)
	if err != nil {
		return nil, err
	}

	// Run the task and wait for it to complete, returning the result.
	if id, err := m.tasks.Add(ctx, "probe", t); err != nil {
		return nil, err
	} else if err := m.tasks.Run(ctx, id); err != nil {
		return nil, err
	} else if status, err := m.tasks.Wait(ctx, id); err != nil {
		return nil, err
	} else if result, ok := status.Result.(*task.ProbeResponse); !ok || result == nil {
		return nil, gomedia.ErrInternalError.With("probe task returned an unexpected result type")
	} else {
		return result, nil
	}
}
