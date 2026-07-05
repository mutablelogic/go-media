package xmp

import (
	"bytes"
	"encoding/json"
	"io"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// XMP holds a parsed XMP metadata document.
type XMP struct {
	about string  // rdf:about value (usually empty string)
	items []*Item // top-level properties
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New returns an empty XMP document.
func New() *XMP {
	return &XMP{}
}

// Parse parses an XMP document from a byte slice.
func Parse(data []byte) (*XMP, error) {
	return Read(bytes.NewReader(data))
}

// Read parses an XMP document from r.
func Read(r io.Reader) (*XMP, error) {
	return decode(r)
}

// Write encodes the XMP document as an XMP/RDF/XML packet to w.
func (x *XMP) Write(w io.Writer) error {
	return encode(w, x)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Items returns all top-level metadata properties.
func (x *XMP) Items() []*Item {
	return x.items
}

// Get returns all items whose key equals key. The key may be "prefix:name" or
// just the local name.
func (x *XMP) Get(key string) []*Item {
	var out []*Item
	for _, it := range x.items {
		if it.Key() == key || it.name == key {
			out = append(out, it)
		}
	}
	return out
}

// Add appends items to the document.
func (x *XMP) Add(items ...*Item) {
	x.items = append(x.items, items...)
}

// First returns the first item with a non-empty value from the first matching
// key in the ordered list. This implements the priority-fallback pattern used
// when the same semantic field can appear under multiple namespace prefixes,
// e.g. x.First("photoshop:DateCreated", "exif:DateTimeOriginal", "xmp:CreateDate").
func (x *XMP) First(keys ...string) *Item {
	for _, key := range keys {
		for _, it := range x.items {
			if (it.Key() == key || it.name == key) && it.Value() != "" {
				return it
			}
		}
	}
	return nil
}

// Delete removes all items matching key and returns the count removed.
func (x *XMP) Delete(key string) int {
	kept := x.items[:0]
	n := 0
	for _, it := range x.items {
		if it.Key() == key || it.name == key {
			n++
		} else {
			kept = append(kept, it)
		}
	}
	x.items = kept
	return n
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// String returns the document encoded as an XMP/XML string.
func (x *XMP) String() string {
	var buf bytes.Buffer
	_ = x.Write(&buf)
	return buf.String()
}

// MarshalJSON implements json.Marshaler.
func (x *XMP) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		About string  `json:"about,omitempty"`
		Items []*Item `json:"items"`
	}{
		About: x.about,
		Items: x.items,
	})
}
