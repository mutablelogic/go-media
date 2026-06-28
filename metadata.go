package media

import "image"

// Metadata is a key/value pair which can be used to describe a media object
// or other metadata. The value can be retrieved as a string value,
// data, or other type. If the value is a byte slice, then it can also
// be retrieved as an image (for artwork)
type Metadata interface {
	// Return the metadata key
	Key() string

	// Return the value as a string. Returns the mimetype
	// if the value is a byte slice, and the mimetype can be
	// detected.
	Value() string

	// Returns the value as a byte slice
	Bytes() []byte

	// Returns the value as an image
	Image() image.Image

	// Returns the value as an interface
	Any() any
}
