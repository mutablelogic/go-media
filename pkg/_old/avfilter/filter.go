package avfilter

import (
	media "github.com/mutablelogic/go-media"
	"github.com/mutablelogic/go-media/pkg/ffmpeg"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Filter struct {
	ctx *ff.AVFilter
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Return all filters
func Filters() []media.Metadata {
	result := make([]media.Metadata, 0, 100)
	var opaque uintptr
	for {
		filter := ff.AVFilter_iterate(&opaque)
		if filter == nil {
			break
		}
		result = append(result, ffmpeg.NewMetadata(filter.Name(), &Filter{filter}))
	}
	return result
}

// Create a new filter by name
func NewFilter(name string) *Filter {
	if filter := ff.AVFilter_get_by_name(name); filter != nil {
		return &Filter{ctx: filter}
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (f *Filter) Name() string {
	return f.ctx.Name()
}

func (f *Filter) Description() string {
	return f.ctx.Description()
}
