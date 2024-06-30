package ffmpeg

import (
	"encoding/json"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Metadata struct {
	Key   string `json:"key" writer:",width:30"`
	Value any    `json:"value,omitempty" writer:",wrap,width:50"`
}

const (
	MetaArtwork = "artwork" // Metadata key for artwork, set the value as []byte
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewMetadata(key string, value any) *Metadata {
	return &Metadata{
		Key:   key,
		Value: value,
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINIGY

func (m *Metadata) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}
