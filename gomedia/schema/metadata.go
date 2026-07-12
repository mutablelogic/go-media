package schema

import (
	"encoding/json"

	gomedia "github.com/mutablelogic/go-media"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Meta struct {
	ContentType string     `json:"content_type,omitempty"`
	Meta        []MetaItem `json:"meta,omitempty"`
}

type MetaItem struct {
	gomedia.Metadata
}

type metaJSONItem struct {
	Key   string `json:"key"`
	Value any    `json:"value,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// MarshalJSON emits metadata as explicit key/value pairs rather than
// marshaling embedded interface-backed structs directly.
func (m Meta) MarshalJSON() ([]byte, error) {
	out := struct {
		ContentType string         `json:"content_type,omitempty"`
		Meta        []metaJSONItem `json:"meta,omitempty"`
	}{
		ContentType: m.ContentType,
	}

	if len(m.Meta) > 0 {
		out.Meta = make([]metaJSONItem, 0, len(m.Meta))
		for _, item := range m.Meta {
			if item.Metadata == nil {
				continue
			}
			out.Meta = append(out.Meta, metaJSONItem{Key: item.Key(), Value: jsonValue(item.Metadata)})
		}
	}

	return json.Marshal(out)
}

func jsonValue(item gomedia.Metadata) any {
	switch v := item.Any().(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		bool,
		[]byte:
		return v
	default:
		return item.Value()
	}
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
