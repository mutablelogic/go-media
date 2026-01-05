package task

import (
	"context"
	"errors"
	"io"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Remux remuxes media streams from input to output without re-encoding.
// The writer must implement io.Writer and can optionally implement
// schema.Writer for enhanced feedback (progress and logging)
func (m *Manager) Remux(ctx context.Context, w io.Writer, req *schema.RemuxRequest) (*schema.RemuxResponse, error) {
	return nil, errors.New("not implemented")
}
