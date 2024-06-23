package media

import "encoding/json"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type meta struct {
	Key   string `json:"key" writer:",width:30"`
	Value any    `json:"value,omitempty" writer:",wrap,width:50"`
}

type metadata struct {
	meta
}

const (
	MetaArtwork  = "artwork"  // Metadata key for artwork, sets the value as []byte
	MetaDuration = "duration" // Metadata key for duration, sets the value as float64 (seconds)
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Return the metadata for the media stream
func newMetadata(key string, value any) Metadata {
	return &metadata{meta{key, value}}
}

////////////////////////////////////////////////////////////////////////////////
// STRINIGY

func (m *metadata) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (m *metadata) Key() string {
	return m.meta.Key
}

func (m *metadata) Value() any {
	return m.meta.Value
}
