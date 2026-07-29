package manager

import (
	"context"

	// Packages
	uuid "github.com/google/uuid"
	otel "github.com/mutablelogic/go-client/pkg/otel"
	schema "github.com/mutablelogic/go-media/task/schema"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// GetTask returns a snapshot of the task registered under id.
func (m *Manager) GetTask(ctx context.Context, id uuid.UUID) (_ *schema.Status, err error) {
	_, endSpan := otel.StartSpan(m.tracer, ctx, "GetTask",
		attribute.String("uuid", id.String()),
	)
	defer func() { endSpan(err) }()

	e, err := m.entry(id)
	if err != nil {
		return nil, err
	}

	e.Lock()
	defer e.Unlock()
	status := e.status

	return &status, nil
}
