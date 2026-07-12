package manager

import (
	"context"
	"strings"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListFilters returns all supported filters with optional name filtering.
func (m *Media) ListFilters(_ context.Context, req schema.ListFilterRequest) (schema.ListFilterResponse, error) {
	var opaque uintptr
	result := make(schema.ListFilterResponse, 0, 512)

	matches := func(f *schema.Filter) bool {
		if req.Name != "" && !strings.Contains(f.Name(), req.Name) {
			return false
		}
		return true
	}

	for {
		filter := ff.AVFilter_iterate(&opaque)
		if filter == nil {
			break
		}
		if f := schema.NewFilter(filter); f != nil && matches(f) {
			result = append(result, *f)
		}
	}

	return result, nil
}
