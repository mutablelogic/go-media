package manager

import (
	"context"
	"log/slog"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

// shutdownTimeout bounds how long Run waits for in-flight tasks to stop
// gracefully after its context is cancelled, so a stuck task can't hang
// shutdown forever.
const shutdownTimeout = 30 * time.Second

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Run blocks until the context is canceled, then cancels any tasks still
// running on the Media's task manager and waits (up to shutdownTimeout) for
// them to stop before returning.
func (m *Media) Run(ctx context.Context, _ *slog.Logger) (err error) {
	// Wait for the context to be canceled
	<-ctx.Done()

	// Cancel any running tasks and wait for them to stop gracefully
	closeCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return m.tasks.Close(closeCtx)
}
