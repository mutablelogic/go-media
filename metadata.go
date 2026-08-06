package media

import (
	"image"
	"io"
	"time"
)

const (
	// MetaArtwork is the metadata key for artwork (album art, cover art, etc.)
	MetaArtwork = "artwork"

	// MetaChapter is the metadata key for a chapter marker. Unlike other
	// metadata, which is a single container-level entry per key, a Metadata
	// slice can hold multiple MetaChapter entries - one per chapter, in
	// order - the same way multiple attached-pic streams yield multiple
	// MetaArtwork entries. Each entry's Any() returns a Chapter.
	MetaChapter = "chapter"
)

// Chapter describes a single chapter marker: its position in the container
// and its own metadata tags (an AVChapter carries its own AVDictionary,
// distinct from the container's - usually just "title", but not limited to
// it). It's the value returned by a MetaChapter entry's Any() method.
type Chapter struct {
	Start    time.Duration
	End      time.Duration
	Metadata map[string]string
}

// Metadata is a key/value pair which can be used to describe a media object
// or other metadata. The value can be retrieved as a string value,
// data, or other type. If the value is a byte slice, then it can also
// be retrieved as an image (for artwork)
type Metadata interface {
	// Return the metadata key (may include ns:name, e.g., "exif:DateTimeOriginal")
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

// NamedReader is an interface that extends io.Reader with a Name() method
type NamedReader interface {
	io.Reader
	Name() string
}

// NamedWriter is an interface that extends io.Writer with a Name() method
type NamedWriter interface {
	io.Writer
	Name() string
}
I 