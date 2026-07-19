package manager

import (
	"context"
	"log/slog"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (profile *Profile) Run(ctx context.Context, logger *slog.Logger) error {
	<-ctx.Done()
	return nil
}
