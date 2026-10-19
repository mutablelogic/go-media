package image_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	// Packages
	metadata "github.com/mutablelogic/go-media/metadata"
)

// svgMetadata returns the metadata of an SVG image, by key
func svgMetadata(t *testing.T, data string) map[string]any {
	t.Helper()
	meta, err := metadata.GetMetadata(context.Background(), strings.NewReader(data), "image/svg+xml")
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string]any, len(meta))
	for _, m := range meta {
		result[m.Key()] = m.Any()
	}
	return result
}

// Test_svg_000 checks the metadata of an image as Inkscape writes it, where
// the title element is preferred to the Dublin Core title, the creator and
// rights are the names in their agents, and the subject is a list
func Test_svg_000(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(TEST_DIR, "sample.svg"))
	if err != nil {
		t.Fatal(err)
	}
	meta := svgMetadata(t, string(data))

	for key, want := range map[string]any{
		"image:format":   "svg",
		"image:width":    794,
		"image:height":   559,
		"dc:title":       "Harbour at dusk",
		"dc:description": "Boats in the harbour at Bergen, drawn from a photograph",
		"dc:creator":     "June Kashio",
		"dc:rights":      "CC BY 4.0",
	} {
		if got := meta[key]; got != want {
			t.Errorf("%s = %#v, want %#v", key, got, want)
		}
	}
	if got, ok := meta["dc:subject"].([]string); !ok || strings.Join(got, ",") != "harbour,boats,Bergen" {
		t.Errorf("dc:subject = %#v", meta["dc:subject"])
	}
	if got, ok := meta["dc:date"].(time.Time); !ok || !got.Equal(time.Date(2021, 6, 14, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("dc:date = %#v", meta["dc:date"])
	}
}

// Test_svg_001 checks the size of an image, from its width and height when
// they're in pixels or have no unit, or else from its viewBox
func Test_svg_001(t *testing.T) {
	for _, tc := range []struct {
		attrs         string
		width, height int
	}{
		{`width="64" height="48"`, 64, 48},
		{`width="64px" height="48.4px"`, 64, 48},
		{`width="10cm" height="5cm" viewBox="0 0 400 200"`, 400, 200},
		{`width="100%" height="100%" viewBox="0,0,32,16"`, 32, 16},
		{`viewBox="-10 -10 20 30"`, 20, 30},
	} {
		meta := svgMetadata(t, `<svg xmlns="http://www.w3.org/2000/svg" `+tc.attrs+`/>`)
		if meta["image:width"] != tc.width || meta["image:height"] != tc.height {
			t.Errorf("%s: size = %v x %v, want %d x %d", tc.attrs, meta["image:width"], meta["image:height"], tc.width, tc.height)
		}
	}

	// Without a size, there's none
	meta := svgMetadata(t, `<svg xmlns="http://www.w3.org/2000/svg" width="10cm"/>`)
	if _, ok := meta["image:width"]; ok {
		t.Errorf("expected no width, got %v", meta["image:width"])
	}
}

// Test_svg_002 checks images which aren't as expected: without a namespace,
// with the Dublin Core title when there's no title element, in another
// encoding, which aren't well-formed, and which aren't SVG
func Test_svg_002(t *testing.T) {
	t.Run("NoNamespace", func(t *testing.T) {
		meta := svgMetadata(t, `<svg width="1" height="1"><title>Plain</title><desc>A desc</desc></svg>`)
		if meta["dc:title"] != "Plain" || meta["dc:description"] != "A desc" {
			t.Errorf("meta = %v", meta)
		}
	})

	t.Run("DublinCoreTitle", func(t *testing.T) {
		meta := svgMetadata(t, `<svg xmlns="http://www.w3.org/2000/svg" xmlns:dc="http://purl.org/dc/elements/1.1/"><metadata><dc:title>From RDF</dc:title></metadata></svg>`)
		if meta["dc:title"] != "From RDF" {
			t.Errorf("dc:title = %#v", meta["dc:title"])
		}
	})

	t.Run("Latin1", func(t *testing.T) {
		meta := svgMetadata(t, "<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?>\n<svg xmlns=\"http://www.w3.org/2000/svg\"><title>Z\xfcrich</title></svg>")
		if meta["dc:title"] != "Zürich" {
			t.Errorf("dc:title = %#v", meta["dc:title"])
		}
	})

	t.Run("Truncated", func(t *testing.T) {
		meta := svgMetadata(t, `<svg xmlns="http://www.w3.org/2000/svg" width="5" height="5"><title>Cut</title><g><rect`)
		if meta["dc:title"] != "Cut" || meta["image:width"] != 5 {
			t.Errorf("meta = %v", meta)
		}
	})

	t.Run("NotSVG", func(t *testing.T) {
		if meta := svgMetadata(t, `<html><title>Page</title></html>`); len(meta) != 0 {
			t.Errorf("expected no metadata, got %v", meta)
		}
		if meta := svgMetadata(t, `not xml at all`); len(meta) != 0 {
			t.Errorf("expected no metadata, got %v", meta)
		}
	})
}
