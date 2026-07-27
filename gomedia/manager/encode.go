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

// EncodeAudio re-encodes an audio stream read from req.Reader into
// req.Profile, writing the result to a temporary file (see
// AudioEncodeMediaResponse). The encode runs as a task tracked by the
// Media's task manager, so it's cancelled along with any other running task
// if ctx is cancelled while the encode is in flight.
func (m *Media) EncodeAudio(ctx context.Context, req task.AudioEncodeMediaRequest) (_ *task.AudioEncodeMediaResponse, err error) {
	ctx, endSpan := otel.StartSpan(m.opt.tracer, ctx, "EncodeAudio",
		attribute.String("req", types.Stringify(req)),
	)
	defer func() { endSpan(err) }()

	// Run the task and wait for it to complete, returning the result.
	if id, err := m.tasks.Add(ctx, "encode", &req); err != nil {
		return nil, err
	} else if err := m.tasks.Run(ctx, id); err != nil {
		return nil, err
	} else if status, err := m.tasks.Wait(ctx, id); err != nil {
		return nil, err
	} else if result, ok := status.Result.(*task.AudioEncodeMediaResponse); !ok || result == nil {
		return nil, gomedia.ErrInternalError.With("encode task returned an unexpected result type")
	} else {
		return result, nil
	}
}
