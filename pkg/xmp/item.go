package xmp

import (
	"encoding/json"
	"image"
	"strings"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Kind describes the shape of an XMP property value.
type Kind uint8

const (
	Simple Kind = iota // a single string (optionally language-tagged)
	Bag                // unordered set of strings
	Seq                // ordered sequence of strings
	Alt                // localized alternatives keyed by xml:lang
	Struct             // nested record with named fields
)

// Item represents a single XMP metadata property and implements media.Metadata.
type Item struct {
	ns     string // namespace URI
	prefix string // preferred namespace prefix
	name   string // local property name
	kind   Kind
	lang   string  // xml:lang qualifier (Simple leaf values)
	value  string  // Simple leaf value
	items  []*Item // Bag/Seq/Alt members or Struct fields
}

var _ media.Metadata = (*Item)(nil)

////////////////////////////////////////////////////////////////////////////////
// CONSTRUCTORS

// NewItem creates a Simple property.
func NewItem(ns, prefix, name, value string) *Item {
	return &Item{ns: ns, prefix: prefix, name: name, kind: Simple, value: value}
}

// NewItemLang creates a Simple property with an xml:lang qualifier.
func NewItemLang(ns, prefix, name, lang, value string) *Item {
	return &Item{ns: ns, prefix: prefix, name: name, kind: Simple, lang: lang, value: value}
}

// NewBag creates an unordered-set property.
func NewBag(ns, prefix, name string, values ...string) *Item {
	return newListItem(ns, prefix, name, Bag, values)
}

// NewSeq creates an ordered-sequence property.
func NewSeq(ns, prefix, name string, values ...string) *Item {
	return newListItem(ns, prefix, name, Seq, values)
}

// NewAlt creates an alternatives property. langValues is a sequence of [lang, value]
// pairs; use "x-default" as the lang for the canonical value.
func NewAlt(ns, prefix, name string, langValues ...[2]string) *Item {
	children := make([]*Item, len(langValues))
	for i, lv := range langValues {
		children[i] = &Item{ns: nsRDF, prefix: "rdf", name: "li", kind: Simple, lang: lv[0], value: lv[1]}
	}
	return &Item{ns: ns, prefix: prefix, name: name, kind: Alt, items: children}
}

// NewStruct creates a structured property with named fields.
func NewStruct(ns, prefix, name string, fields ...*Item) *Item {
	return &Item{ns: ns, prefix: prefix, name: name, kind: Struct, items: fields}
}

func newListItem(ns, prefix, name string, kind Kind, values []string) *Item {
	children := make([]*Item, len(values))
	for i, v := range values {
		children[i] = &Item{ns: nsRDF, prefix: "rdf", name: "li", kind: Simple, value: v}
	}
	return &Item{ns: ns, prefix: prefix, name: name, kind: kind, items: children}
}

////////////////////////////////////////////////////////////////////////////////
// ACCESSORS

// NS returns the namespace URI.
func (it *Item) NS() string { return it.ns }

// Prefix returns the preferred namespace prefix.
func (it *Item) Prefix() string { return it.prefix }

// LocalName returns the local property name.
func (it *Item) LocalName() string { return it.name }

// ItemKind returns the value kind.
func (it *Item) ItemKind() Kind { return it.kind }

// Lang returns the xml:lang qualifier (non-empty for Simple items with a language tag
// and for Alt child items).
func (it *Item) Lang() string { return it.lang }

// Items returns child items: members of Bag/Seq/Alt, or fields of a Struct.
func (it *Item) Items() []*Item { return it.items }

// ValueType returns the registered scalar type for this item's key.
func (it *Item) ValueType() ValueType {
	return ValueTypeForKey(it.Key())
}

// AsTime parses this Simple item as a time value.
func (it *Item) AsTime() (time.Time, bool) {
	if it.kind != Simple {
		return time.Time{}, false
	}
	return parseTimeValue(it.value)
}

// AsDuration parses this Simple item as a duration.
func (it *Item) AsDuration() (time.Duration, bool) {
	if it.kind != Simple {
		return 0, false
	}
	return parseDurationValue(it.value)
}

// AsBool parses this Simple item as a boolean.
func (it *Item) AsBool() (bool, bool) {
	if it.kind != Simple {
		return false, false
	}
	return parseBoolValue(it.value)
}

// AsRational parses this Simple item as a rational value.
func (it *Item) AsRational() (Rational, bool) {
	if it.kind != Simple {
		return Rational{}, false
	}
	return parseRationalValue(it.value)
}

// AsGPSCoord parses this Simple item as a decimal degree coordinate.
func (it *Item) AsGPSCoord() (float64, bool) {
	if it.kind != Simple {
		return 0, false
	}
	return parseGPSCoordValue(it.value)
}

// TypedValue returns a typed scalar for Simple items where a registered type
// exists, and falls back to string (or the normal Any() for non-Simple kinds).
func (it *Item) TypedValue() any {
	if it.kind != Simple {
		return it.Any()
	}
	switch it.ValueType() {
	case ValueTypeTime:
		if v, ok := it.AsTime(); ok {
			return v
		}
	case ValueTypeDuration:
		if v, ok := it.AsDuration(); ok {
			return v
		}
	case ValueTypeBoolean:
		if v, ok := it.AsBool(); ok {
			return v
		}
	case ValueTypeRational:
		if v, ok := it.AsRational(); ok {
			return v
		}
	case ValueTypeGPSCoord:
		if v, ok := it.AsGPSCoord(); ok {
			return v
		}
	}
	return it.value
}

////////////////////////////////////////////////////////////////////////////////
// media.Metadata

// Key returns "prefix:name" (implements media.Metadata).
func (it *Item) Key() string {
	if it.prefix == "" {
		return it.name
	}
	return it.prefix + ":" + it.name
}

// Value returns the string value (implements media.Metadata).
// For Bag/Seq, values are joined with "; ".
// For Alt, the "x-default" value is returned, falling back to the first entry.
// For Struct, an empty string is returned.
func (it *Item) Value() string {
	switch it.kind {
	case Simple:
		return it.value
	case Bag, Seq:
		parts := make([]string, 0, len(it.items))
		for _, child := range it.items {
			parts = append(parts, child.value)
		}
		return strings.Join(parts, "; ")
	case Alt:
		for _, child := range it.items {
			if child.lang == "x-default" {
				return child.value
			}
		}
		if len(it.items) > 0 {
			return it.items[0].value
		}
	}
	return ""
}

// Bytes returns nil — XMP properties are text-only (implements media.Metadata).
func (it *Item) Bytes() []byte { return nil }

// Image returns nil (implements media.Metadata).
func (it *Item) Image() image.Image { return nil }

// Any returns the property value as an appropriate Go type (implements media.Metadata):
//   - Simple  → string
//   - Bag/Seq → []string
//   - Alt     → map[string]string  (lang → value)
//   - Struct  → []*Item
func (it *Item) Any() any {
	switch it.kind {
	case Simple:
		return it.value
	case Bag, Seq:
		out := make([]string, len(it.items))
		for i, child := range it.items {
			out[i] = child.value
		}
		return out
	case Alt:
		out := make(map[string]string, len(it.items))
		for _, child := range it.items {
			out[child.lang] = child.value
		}
		return out
	case Struct:
		return it.items
	}
	return nil
}

// String returns "key=value".
func (it *Item) String() string {
	return it.Key() + "=" + it.Value()
}

// MarshalJSON implements json.Marshaler.
func (it *Item) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Key  string `json:"key"`
		NS   string `json:"ns,omitempty"`
		Kind string `json:"kind"`
		Lang string `json:"lang,omitempty"`
		Val  any    `json:"value"`
	}{
		Key:  it.Key(),
		NS:   it.ns,
		Kind: it.kind.String(),
		Lang: it.lang,
		Val:  it.Any(),
	})
}

////////////////////////////////////////////////////////////////////////////////
// Kind stringer

func (k Kind) String() string {
	switch k {
	case Simple:
		return "simple"
	case Bag:
		return "bag"
	case Seq:
		return "seq"
	case Alt:
		return "alt"
	case Struct:
		return "struct"
	}
	return "unknown"
}
