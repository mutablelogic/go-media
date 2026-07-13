package heif

import (
	"encoding/json"
	"fmt"
	"image"
	"strings"

	// Packages
	media "github.com/mutablelogic/go-media"
	"github.com/mutablelogic/go-media/pkg/exif"
	"github.com/mutablelogic/go-media/pkg/xmp"
	libheif "github.com/mutablelogic/go-media/sys/libheif"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Meta struct {
	key  string
	val  string
	data []byte
	any  any
}

var _ media.Metadata = (*Meta)(nil)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (m *Meta) Key() string        { return m.key }
func (m *Meta) Value() string      { return m.val }
func (m *Meta) Bytes() []byte      { return append([]byte(nil), m.data...) }
func (m *Meta) Image() image.Image { return nil }
func (m *Meta) Any() any           { return m.any }

func (m *Meta) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{Key: m.key, Value: m.val})
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Metadata returns metadata from the primary image.
//
// EXIF and XMP payloads are parsed and expanded into their structured metadata
// entries. Any remaining blocks are returned as generic HEIF metadata records.
func (h *HEIF) Metadata() []media.Metadata {
	if h == nil || h.ctx == nil {
		return nil
	}

	handle, err := libheif.Libheif_context_get_primary_image_handle(h.ctx)
	if err != nil || handle == nil {
		return nil
	}
	defer libheif.Libheif_image_handle_release(handle)

	count := libheif.Libheif_image_handle_get_number_of_metadata_blocks(handle, "")
	if count <= 0 {
		return nil
	}

	ids := libheif.Libheif_image_handle_get_list_of_metadata_block_IDs(handle, "", count)
	if len(ids) == 0 {
		return nil
	}

	var result []media.Metadata
	for _, id := range ids {
		result = append(result, h.metadataForBlock(handle, id)...)
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (h *HEIF) metadataForBlock(handle *libheif.ImageHandle, id libheif.ItemID) []media.Metadata {
	data, err := libheif.Libheif_image_handle_get_metadata(handle, id)
	if err != nil || len(data) == 0 {
		return nil
	}

	typeName := strings.TrimSpace(libheif.Libheif_image_handle_get_metadata_type(handle, id))
	contentType := strings.TrimSpace(libheif.Libheif_image_handle_get_metadata_content_type(handle, id))

	if parsed := parseMetadataBlock(typeName, contentType, data); len(parsed) > 0 {
		return parsed
	}

	key := typeName
	if key == "" {
		key = contentType
	}
	if key == "" {
		key = "metadata"
	}
	return []media.Metadata{&Meta{
		key:  "heif:" + key,
		val:  valueForMetadataBlock(typeName, contentType, data),
		data: data,
		any:  append([]byte(nil), data...),
	}}
}

func parseMetadataBlock(typeName, contentType string, data []byte) []media.Metadata {
	lowerType := strings.ToLower(typeName)
	lowerContent := strings.ToLower(contentType)

	if strings.Contains(lowerType, "exif") || strings.Contains(lowerContent, "exif") {
		if items := parseEXIFMetadata(data); len(items) > 0 {
			return items
		}
	}

	if strings.Contains(lowerType, "xmp") || strings.Contains(lowerContent, "xmp") || strings.Contains(lowerContent, "rdf") || strings.Contains(lowerContent, "xml") {
		if items := parseXMPMetadata(data); len(items) > 0 {
			return items
		}
	}

	return nil
}

func parseEXIFMetadata(data []byte) []media.Metadata {
	if doc, err := exif.Parse(data); err == nil {
		return metadataItemsFromEXIF(doc)
	}
	if stripped := unwrapHEIFExif(data); len(stripped) > 0 {
		if doc, err := exif.Parse(stripped); err == nil {
			return metadataItemsFromEXIF(doc)
		}
	}
	return nil
}

func parseXMPMetadata(data []byte) []media.Metadata {
	if doc := parseXMPDocument(data); doc != nil {
		return metadataItemsFromXMP(doc)
	}
	return nil
}

func metadataItemsFromEXIF(doc *exif.EXIF) []media.Metadata {
	items := make([]media.Metadata, 0, len(doc.Tags()))
	for _, tag := range doc.Tags() {
		items = append(items, tag)
	}
	return items
}

func metadataItemsFromXMP(doc *xmp.XMP) []media.Metadata {
	items := make([]media.Metadata, 0, len(doc.Items()))
	for _, item := range doc.Items() {
		items = append(items, item)
	}
	return items
}

func (h *HEIF) xmpMetadataBlocks(handle *libheif.ImageHandle) [][]*xmp.Item {
	count := libheif.Libheif_image_handle_get_number_of_metadata_blocks(handle, "XMP")
	if count <= 0 {
		return nil
	}

	ids := libheif.Libheif_image_handle_get_list_of_metadata_block_IDs(handle, "XMP", count)
	if len(ids) == 0 {
		return nil
	}

	blocks := make([][]*xmp.Item, 0, len(ids))
	for _, id := range ids {
		data, err := libheif.Libheif_image_handle_get_metadata(handle, id)
		if err != nil || len(data) == 0 {
			continue
		}
		if doc := parseXMPDocument(data); doc != nil {
			blocks = append(blocks, doc.Items())
		}
	}
	return blocks
}

func parseXMPDocument(data []byte) *xmp.XMP {
	if doc, err := xmp.Parse(data); err == nil {
		return doc
	}
	if stripped := unwrapHEIFMetadata(data); len(stripped) > 0 {
		if doc, err := xmp.Parse(stripped); err == nil {
			return doc
		}
	}
	return nil
}

func unwrapHEIFExif(data []byte) []byte {
	if len(data) >= 4 {
		return data[4:]
	}
	return nil
}

func unwrapHEIFMetadata(data []byte) []byte {
	if len(data) <= 4 {
		return nil
	}
	if data[4] == 'E' || data[4] == 'x' || data[4] == '<' {
		return data[4:]
	}
	return nil
}

func valueForMetadataBlock(typeName, contentType string, data []byte) string {
	switch {
	case contentType != "" && typeName != "":
		return fmt.Sprintf("%s (%s)", contentType, typeName)
	case contentType != "":
		return contentType
	case typeName != "":
		return typeName
	default:
		return fmt.Sprintf("%d bytes", len(data))
	}
}
