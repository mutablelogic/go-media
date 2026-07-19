package schema

import (
	"encoding/json"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Meta struct {
	Name        string     `json:"name,omitempty" yaml:"name,omitempty"`
	ContentType string     `json:"content_type,omitempty" yaml:"content_type,omitempty"`
	Meta        []MetaItem `json:"meta,omitempty" yaml:"meta,omitempty"`
}

type MetaItem struct {
	Name             string `json:"name,omitempty" yaml:"name,omitempty"`
	MetaKey          string `json:"key,omitempty" yaml:"key,omitempty"`
	MetaValue        any    `json:"value,omitempty" yaml:"value,omitempty"`
	gomedia.Metadata `json:"-" yaml:"-"`
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m MetaItem) Key() string {
	if m.Metadata != nil {
		return m.Metadata.Key()
	}
	return m.MetaKey
}

func (m MetaItem) Any() any {
	if m.Metadata != nil {
		return m.Metadata.Any()
	}
	return m.MetaValue
}

func (m MetaItem) MarshalJSON() ([]byte, error) {
	type kv struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
	}
	return json.Marshal(kv{Key: m.Key(), Value: m.Any()})
}

func (m MetaItem) MarshalYAML() (any, error) {
	type kv struct {
		Key   string `yaml:"key"`
		Value any    `yaml:"value"`
	}
	return kv{Key: m.Key(), Value: m.Any()}, nil
}

////////////////////////////////////////////////////////////////////////////////
// TABLE WRITER

func (MetaItem) Header() []string {
	return []string{"Path", "Key", "Value"}
}

func (r MetaItem) Cell(col int) string {
	switch col {
	case 0:
		return r.Name
	case 1:
		return r.Key()
	case 2:
		return types.Stringify(r.Any())
	default:
		return ""
	}
}

func (MetaItem) Width(col int) int {
	return 0
}
