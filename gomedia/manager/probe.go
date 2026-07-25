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

	probeTask, err := task.NewProbeTask(req)
	if err != nil {
		return nil, err
	}

	id, err := m.tasks.Add(ctx, "probe", probeTask)
	if err != nil {
		return nil, err
	}
	if err := m.tasks.Run(ctx, id); err != nil {
		return nil, err
	}
	status, err := m.tasks.Wait(ctx, id)
	if err != nil {
		return nil, err
	}

	result, ok := status.Result.(*task.ProbeResponse)
	if !ok {
		return nil, gomedia.ErrInternalError.With("probe task returned an unexpected result type")
	}

	return result, nil
}
