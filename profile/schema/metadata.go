package schema

import (
	"maps"
	"slices"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Metadata is a JSON-friendly key/value view of a gomedia.Metadata entry -
// binary payloads (e.g. artwork) aren't represented here.
type Metadata struct {
	Key   string `json:"key" help:"Metadata tag name, e.g. \"title\", \"artist\", \"language\"." example:"title"`
	Value string `json:"value,omitempty" help:"Metadata tag value." example:"Sample Track"`
	Any   any    `json:"any,omitempty" help:"Metadata tag value, as a JSON-friendly type (string, number, bool, array, or object)."`
}

// Artwork is a JSON-friendly view of a gomedia.Metadata entry holding
// embedded artwork (cover art, thumbnail, ...) - the binary payload Metadata
// omits, decoded far enough to report its mimetype and pixel dimensions.
type Artwork struct {
	Type   string `json:"type,omitempty" help:"Artwork mimetype, e.g. \"image/jpeg\"." example:"image/jpeg"`
	Width  int    `json:"width,omitempty" help:"Artwork width, in pixels." example:"600"`
	Height int    `json:"height,omitempty" help:"Artwork height, in pixels." example:"600"`
	Data   []byte `json:"data,omitempty" help:"Raw artwork image data."`
}

// Chapter is a JSON-friendly view of a gomedia.Chapter - a single chapter
// marker's position in the container plus its own metadata tags (usually
// just "title").
type Chapter struct {
	Start    Duration   `json:"start" help:"Chapter start time, as a duration string (e.g. \"1h2m3.5s\")." example:"1m30s"`
	End      Duration   `json:"end,omitempty" help:"Chapter end time, as a duration string." example:"3m45s"`
	Metadata []Metadata `json:"metadata,omitempty" help:"Chapter tags, e.g. \"title\"."`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewMetadataList converts gomedia.Metadata entries (an interface - which
// encoding/json can marshal but never unmarshal, having no concrete type to
// instantiate) to the JSON-friendly, round-trippable Metadata shape. Used
// both for StreamProfile's per-stream metadata and for container-level
// metadata elsewhere (e.g. task.ProbeResponse).
func NewMetadataList(entries []gomedia.Metadata) []Metadata {
	if len(entries) == 0 {
		return nil
	}
	result := make([]Metadata, 0, len(entries))
	for _, e := range entries {
		if e == nil {
			continue
		}
		result = append(result, Metadata{Key: e.Key(), Value: e.Value(), Any: e.Any()})
	}
	return result
}

// NewArtwork converts a gomedia.Metadata entry holding artwork (see
// gomedia.MetaArtwork) into the JSON-friendly Artwork shape. Value() is used
// for Type since, per gomedia.Metadata's contract, it returns the mimetype
// when the underlying value is a byte slice. Width/Height are read from
// Image(), which is nil (leaving both zero) if the data can't be decoded as
// an image.
func NewArtwork(entry gomedia.Metadata) *Artwork {
	if entry == nil {
		return nil
	}
	artwork := &Artwork{
		Type: entry.Value(),
		Data: entry.Bytes(),
	}
	if img := entry.Image(); img != nil {
		bounds := img.Bounds()
		artwork.Width = bounds.Dx()
		artwork.Height = bounds.Dy()
	}
	return artwork
}

// NewArtworkList converts gomedia.Metadata entries holding artwork (see
// gomedia.MetaArtwork) to the JSON-friendly Artwork shape, one per entry -
// e.g. an audio file with multiple attached-pic streams (cover, cover-2, ...).
func NewArtworkList(entries []gomedia.Metadata) []Artwork {
	if len(entries) == 0 {
		return nil
	}
	result := make([]Artwork, 0, len(entries))
	for _, e := range entries {
		if e == nil {
			continue
		}
		result = append(result, *NewArtwork(e))
	}
	return result
}

// NewChapter converts a gomedia.Chapter into the JSON-friendly Chapter
// shape. Tags are sorted by key for deterministic output, since a
// gomedia.Chapter's Metadata is a map and iteration order isn't stable.
func NewChapter(chapter gomedia.Chapter) Chapter {
	tags := make([]Metadata, 0, len(chapter.Metadata))
	for _, key := range slices.Sorted(maps.Keys(chapter.Metadata)) {
		tags = append(tags, Metadata{Key: key, Value: chapter.Metadata[key]})
	}
	return Chapter{
		Start:    Duration(chapter.Start),
		End:      Duration(chapter.End),
		Metadata: tags,
	}
}

// NewChapterList converts gomedia.Metadata entries holding chapter markers
// (see gomedia.MetaChapter) to the JSON-friendly Chapter shape, one per
// entry, in the same order the chapters appear in the container.
func NewChapterList(entries []gomedia.Metadata) []Chapter {
	if len(entries) == 0 {
		return nil
	}
	result := make([]Chapter, 0, len(entries))
	for _, e := range entries {
		if e == nil {
			continue
		}
		chapter, ok := e.Any().(gomedia.Chapter)
		if !ok {
			continue
		}
		result = append(result, NewChapter(chapter))
	}
	return result
}
