package schema

import (
	"encoding/json"
	"fmt"
	"image"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

type testMetadata struct {
	key   string
	value any
}

func (m testMetadata) Key() string        { return m.key }
func (m testMetadata) Value() string      { return fmt.Sprint(m.value) }
func (m testMetadata) Bytes() []byte      { return nil }
func (m testMetadata) Image() image.Image { return nil }
func (m testMetadata) Any() any           { return m.value }

func TestMetaMarshalJSON(t *testing.T) {
	m := Meta{
		ContentType: "audio/mp4",
		Meta:        []MetaItem{},
	}

	// Build with concrete metadata values wrapped as schema.MetaItem.
	m.Meta = append(m.Meta, MetaItem{Metadata: testMetadata{key: "dc:title", value: "Jenny Ondioline"}})
	m.Meta = append(m.Meta, MetaItem{Metadata: testMetadata{key: "audio:Duration", value: 1088.179955}})
	m.Meta = append(m.Meta, MetaItem{Metadata: testMetadata{key: "audio:IsLive", value: true}})
	m.Meta = append(m.Meta, MetaItem{Metadata: testMetadata{key: "audio:Blob", value: []byte{0x01, 0x02}}})

	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}

	if got := out["content_type"]; got != "audio/mp4" {
		t.Fatalf("content_type = %v, want %q", got, "audio/mp4")
	}
	arr, ok := out["meta"].([]any)
	if !ok {
		t.Fatalf("meta has unexpected type %T", out["meta"])
	}
	if len(arr) != 4 {
		t.Fatalf("meta length = %d, want 4", len(arr))
	}
	first, ok := arr[0].(map[string]any)
	if !ok {
		t.Fatalf("meta[0] has unexpected type %T", arr[0])
	}
	if first["key"] != "dc:title" || first["value"] != "Jenny Ondioline" {
		t.Fatalf("meta[0] = %v, want key/value pair", first)
	}
	second, ok := arr[1].(map[string]any)
	if !ok {
		t.Fatalf("meta[1] has unexpected type %T", arr[1])
	}
	if second["key"] != "audio:Duration" {
		t.Fatalf("meta[1] key = %v, want %q", second["key"], "audio:Duration")
	}
	if v, ok := second["value"].(float64); !ok || v != 1088.179955 {
		t.Fatalf("meta[1] value = %#v, want float64 1088.179955", second["value"])
	}
	third, ok := arr[2].(map[string]any)
	if !ok {
		t.Fatalf("meta[2] has unexpected type %T", arr[2])
	}
	if third["key"] != "audio:IsLive" || third["value"] != true {
		t.Fatalf("meta[2] = %v, want boolean value", third)
	}
	fourth, ok := arr[3].(map[string]any)
	if !ok {
		t.Fatalf("meta[3] has unexpected type %T", arr[3])
	}
	if fourth["key"] != "audio:Blob" {
		t.Fatalf("meta[3] key = %v, want %q", fourth["key"], "audio:Blob")
	}
	if fourth["value"] != "AQI=" {
		t.Fatalf("meta[3] value = %v, want base64 %q", fourth["value"], "AQI=")
	}
}

func TestMetaItemYAMLUnmarshal(t *testing.T) {
	input := `name: sample
content_type: audio/mp4
meta:
  - key: dc:title
    value: Jenny Ondioline
  - key: audio:Duration
    value: 1088.179955
  - key: audio:IsLive
    value: true
`

	var m Meta
	if err := yaml.Unmarshal([]byte(input), &m); err != nil {
		t.Fatal(err)
	}

	if m.ContentType != "audio/mp4" {
		t.Fatalf("content_type = %q, want %q", m.ContentType, "audio/mp4")
	}
	if len(m.Meta) != 3 {
		t.Fatalf("meta length = %d, want 3", len(m.Meta))
	}
	if m.Meta[0].Key() != "dc:title" || m.Meta[0].Any() != "Jenny Ondioline" {
		t.Fatalf("meta[0] = (%q,%v), want (%q,%q)", m.Meta[0].Key(), m.Meta[0].Any(), "dc:title", "Jenny Ondioline")
	}
	if m.Meta[1].Key() != "audio:Duration" {
		t.Fatalf("meta[1] key = %q, want %q", m.Meta[1].Key(), "audio:Duration")
	}
	if _, ok := m.Meta[1].Any().(float64); !ok {
		t.Fatalf("meta[1] value type = %T, want float64", m.Meta[1].Any())
	}
	if m.Meta[2].Key() != "audio:IsLive" || m.Meta[2].Any() != true {
		t.Fatalf("meta[2] = (%q,%v), want (%q,%v)", m.Meta[2].Key(), m.Meta[2].Any(), "audio:IsLive", true)
	}
}
