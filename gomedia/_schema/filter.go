package schema

import (
	"encoding/json"
	"strconv"

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

	// Get options directly from filter priv_class when available.
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

	filterJSON, err := r.AVFilter.MarshalJSON()
	if err != nil {
		return nil, err
	}

	if len(r.Opts) == 0 {
		return filterJSON, nil
	}

	var result map[string]any
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

////////////////////////////////////////////////////////////////////////////////
// TABLE WRITER

func (Filter) Header() []string {
	return []string{"Name", "Inputs", "Outputs", "Description"}
}

func (f Filter) Cell(col int) string {
	switch col {
	case 0:
		if f.AVFilter == nil {
			return ""
		}
		return f.Name()
	case 1:
		if f.AVFilter == nil {
			return ""
		}
		return strconv.FormatUint(uint64(f.NumInputs()), 10)
	case 2:
		if f.AVFilter == nil {
			return ""
		}
		return strconv.FormatUint(uint64(f.NumOutputs()), 10)
	case 3:
		if f.AVFilter == nil {
			return ""
		}
		return f.Description()
	default:
		return ""
	}
}

func (Filter) Width(col int) int {
	switch col {
	case 0:
		return 24
	case 1, 2:
		return 12
	default:
		return 0
	}
}
