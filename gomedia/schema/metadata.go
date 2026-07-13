package schema

import (
	"encoding/json"

	gomedia "github.com/mutablelogic/go-media"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Meta struct {
	Name        string     `json:"name,omitempty"`
	ContentType string     `json:"content_type,omitempty"`
	Meta        []MetaItem `json:"meta,omitempty"`
}

type MetaItem struct {
	gomedia.Metadata
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m MetaItem) MarshalJSON() ([]byte, error) {
	type kv struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
	}
	return json.Marshal(kv{Key: m.Key(), Value: m.Value()})
}

////////////////////////////////////////////////////////////////////////////////
// TABLE WRITER

func (MetaItem) Header() []string {
	return []string{"Key", "Value"}
}

func (r MetaItem) Cell(col int) string {
	switch col {
	case 0:
		return r.Key()
	case 1:
		return types.Stringify(r.Value())
	default:
		return ""
	}
}

func (MetaItem) Width(col int) int {
	return 0
}
