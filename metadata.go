package media

import "encoding/json"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type metadata struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Return the metadata for the media stream
func newMetadata(key string, value any) Metadata {
	return &metadata{key, value}
}

////////////////////////////////////////////////////////////////////////////////
// STRINIGY

func (m *metadata) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}
