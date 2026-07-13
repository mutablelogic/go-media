package manager

import (
	"context"
	"log/slog"
)

// Run blocks until the context is canceled and returns the context error.
func Run(ctx context.Context, _ *slog.Logger) error {
	<-ctx.Done()
	return ctx.Err()
}
