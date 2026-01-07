package schema

import (
	"encoding/json"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ListFilterRequest struct {
	Name string `json:"name" help:"Filter by filter name (partial match)"`
}

type ListFilterResponse []Filter

type Filter struct {
	*ff.AVFilter
	Opts []*ff.AVOption `json:"options,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFilter(filter *ff.AVFilter) *Filter {
	if filter == nil {
		return nil
	}
	f := &Filter{AVFilter: filter}

	// Get options directly from filter's priv_class if available
	// Uses the FAKE_OBJ trick like ffmpeg's cmdutils
	if class := filter.PrivClass(); class != nil {
		f.Opts = ff.AVUtil_opt_list_from_class(class)
	}

	return f
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r Filter) MarshalJSON() ([]byte, error) {
	if r.AVFilter == nil {
		return json.Marshal(nil)
	}

	// Get base filter JSON
	filterJSON, err := r.AVFilter.MarshalJSON()
	if err != nil {
		return nil, err
	}

	// If no options, return base JSON
	if len(r.Opts) == 0 {
		return filterJSON, nil
	}

	// Unmarshal to map and add options
	var result map[string]interface{}
	if err := json.Unmarshal(filterJSON, &result); err != nil {
		return nil, err
	}
	result["options"] = r.Opts

	return json.Marshal(result)
}

func (r Filter) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r ListFilterResponse) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
