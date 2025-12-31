package ffmpeg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"

	// Packages
	media "github.com/mutablelogic/go-media"
	file "github.com/mutablelogic/go-media/pkg/file"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type meta struct {
	Key   string `json:"key" writer:",width:30"`
	Value any    `json:"value,omitempty" writer:",wrap,width:50"`
}

type Metadata struct {
	meta
}

var _ media.Metadata = (*Metadata)(nil)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	MetaArtwork = "artwork" // Metadata key for artwork, set the value as []byte
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Metadata implementation
func NewMetadata(key string, value any) *Metadata {
	return &Metadata{
		meta: meta{
			Key:   key,
			Value: value,
		},
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m *Metadata) String() string {
	data, _ := json.MarshalIndent(m, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (m *Metadata) Key() string {
	return m.meta.Key
}

// Value returns the value as a string. If the value is a byte slice, it will
// return the mimetype of the byte slice.
func (m *Metadata) Value() string {
	if m.meta.Value == nil {
		return ""
	}
	switch v := m.meta.Value.(type) {
	case string:
		return v
	case []byte:
		if mimetype, _, err := file.MimeType(v); err == nil {
			return mimetype
		} else {
			return ""
		}
	default:
		return fmt.Sprint(v)
	}
}

// Returns the value as a byte slice
func (m *Metadata) Bytes() []byte {
	if m.meta.Value == nil {
		return nil
	}
	switch v := m.meta.Value.(type) {
	case []byte:
		return v
	}
	return nil
}

// Returns the value as an image
func (m *Metadata) Image() image.Image {
	if m.meta.Value == nil {
		return nil
	}
	switch v := m.meta.Value.(type) {
	case []byte:
		if img, _, err := image.Decode(bytes.NewReader(v)); err == nil {
			return img
		}
	}
	return nil
}

// Returns the value as an interface
func (m *Metadata) Any() any {
	return m.meta.Value
}
