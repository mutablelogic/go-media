package schema

import (
	"encoding/json"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ListFilterRequest struct {
	Name string `json:"name"`
}

type ListFilterResponse []Filter

type Filter struct {
	*ff.AVFilter
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFilter(filter *ff.AVFilter) *Filter {
	if filter == nil {
		return nil
	}
	return &Filter{AVFilter: filter}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r Filter) MarshalJSON() ([]byte, error) {
	if r.AVFilter == nil {
		return json.Marshal(nil)
	}
	return r.AVFilter.MarshalJSON()
}

func (r Filter) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
