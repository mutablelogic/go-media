package reader

import (
	"bytes"
	"fmt"
	"image"

	// Image imports for decoding
	_ "image/gif"  // Register GIF decoder for artwork.DecodeConfig
	_ "image/jpeg" // Register JPEG decoder for artwork.DecodeConfig
	_ "image/png"  // Register PNG decoder for artwork.DecodeConfig

	_ "golang.org/x/image/bmp"  // Register BMP decoder for artwork.DecodeConfig
	_ "golang.org/x/image/webp" // Register WebP decoder for artwork.DecodeConfig

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type meta struct {
	key   string
	value any
}

var _ gomedia.Metadata = (*meta)(nil)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (m *meta) Key() string {
	if m == nil {
		return ""
	}
	return m.key
}

// Value returns the value as a string. If the value is a byte slice, it will
// return the mimetype of the byte slice.
func (m *meta) Value() string {
	if m == nil || m.value == nil {
		return ""
	}
	switch v := m.value.(type) {
	case string:
		return v
	case []byte:
		mimetype, _, err := metadata.ContentType(bytes.NewReader(v))
		if err == nil {
			return mimetype
		} else {
			return ""
		}
	default:
		return fmt.Sprint(v)
	}
}

// Returns the value as a byte slice
func (m *meta) Bytes() []byte {
	if m == nil || m.value == nil {
		return nil
	}
	switch v := m.value.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	}
	return nil
}

// Returns the value as an image
func (m *meta) Image() image.Image {
	if m == nil || m.value == nil {
		return nil
	}
	switch v := m.value.(type) {
	case []byte:
		if img, _, err := image.Decode(bytes.NewReader(v)); err == nil {
			return img
		}
	}
	return nil
}

// Returns the value as an interface
func (m *meta) Any() any {
	if m == nil {
		return nil
	}
	return m.value
}
