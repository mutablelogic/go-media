package media

import "encoding/json"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type metadata struct {
	Key   string `json:"key" writer:",width:30"`
	Value any    `json:"value" writer:",wrap,width:50"`
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
