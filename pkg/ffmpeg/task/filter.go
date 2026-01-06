package task

import (
	"context"
	"strings"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return all supported filters
func (manager *Manager) ListFilters(_ context.Context, req *schema.ListFilterRequest) (schema.ListFilterResponse, error) {
	var opaque uintptr
	result := make(schema.ListFilterResponse, 0, 512)

	// Filter function
	matches := func(req *schema.ListFilterRequest, f *schema.Filter) bool {
		if req == nil {
			return true
		}
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
		if f := schema.NewFilter(filter); f != nil {
			if matches(req, f) {
				result = append(result, *f)
			}
		}
	}
	return result, nil
}
