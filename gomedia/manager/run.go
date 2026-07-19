package manager

import (
	"context"
	"log/slog"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Run blocks until the context is canceled and returns the context error.
func (m *Media) Run(ctx context.Context, _ *slog.Logger) (err error) {
	// If the context is cancelled while starting up (before the runloop's own
	// graceful shutdown handling takes over), don't report that as a failure.
	defer func() {
		if err != nil && ctx.Err() != nil {
			err = nil
		}
	}()

	// Wait for the context to be canceled
	<-ctx.Done()

	// Return success
	return nil
}
