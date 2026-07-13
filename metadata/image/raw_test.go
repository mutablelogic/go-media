package image_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	// Packages
	metadata "github.com/mutablelogic/go-media/metadata"
	_ "github.com/mutablelogic/go-media/metadata/image"
)

const rawTestFile = "RAW_OLYMPUS_E3.ORF"

// Test_raw_000 checks that the expected metadata is extracted from a RAW
// file: libraw's own curated fields (tiff:/exif:/dc:) and the derived
// image:width/height/format, and that no artwork:thumbnail leaks in when
// it wasn't requested.
func Test_raw_000(t *testing.T) {
	path := filepath.Join(TEST_DIR, rawTestFile)
	contentType := contentTypeForFile(t, path)
	if contentType != "image/x-olympus-orf" {
		t.Fatalf("unexpected content type %q", contentType)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	meta, err := metadata.GetMetadata(context.Background(), f, contentType, "")
	if err != nil {
		// The generic image/* decoder handler is expected to fail on RAW
		// bytes (they're not a stdlib-decodable image container); that's a
		// non-fatal warning, not a reason to discard the metadata that did
		// come back from the RAW handler.
		t.Logf("warning: %v", err)
	}

	got := make(map[string]string, len(meta))
	for _, m := range meta {
		got[m.Key()] = m.Value()
	}

	want := map[string]string{
		"tiff:Make":             "Olympus",
		"tiff:Model":            "E-3",
		"tiff:Software":         "Version 1.0",
		"exif:ISOSpeedRatings":  "200",
		"exif:ExposureTime":     "1/320",
		"exif:FNumber":          "f/2.0",
		"exif:FocalLength":      "50mm",
		"exif:DateTimeOriginal": "2008-12-19T12:29:40Z",
		"dc:description":        "OLYMPUS DIGITAL CAMERA",
		"image:width":           "3720",
		"image:height":          "2800",
		"image:format":          "olympus",
	}
	for key, wantVal := range want {
		if got[key] != wantVal {
			t.Errorf("%s = %q, want %q", key, got[key], wantVal)
		}
	}

	if _, ok := got["artwork:thumbnail"]; ok {
		t.Error("did not expect artwork:thumbnail without an artwork: filter")
	}
}

// Test_raw_001 checks that namespace-scoped filters only return entries
// from that namespace.
func Test_raw_001(t *testing.T) {
	tests := []struct {
		filter    string
		namespace string
	}{
		{"tiff:", "tiff"},
		{"exif:", "exif"},
		{"dc:", "dc"},
		{"image:", "image"},
	}
	for _, test := range tests {
		t.Run(test.filter, func(t *testing.T) {
			path := filepath.Join(TEST_DIR, rawTestFile)
			contentType := contentTypeForFile(t, path)

			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			meta, err := metadata.GetMetadata(context.Background(), f, contentType, test.filter)
			if err != nil {
				// image: also matches the generic image/* decoder handler,
				// which is expected to fail on RAW bytes
				t.Logf("warning: %v", err)
			}
			if len(meta) == 0 {
				t.Fatalf("expected at least one entry for filter %q", test.filter)
			}
			for _, m := range meta {
				namespace, _, _ := strings.Cut(m.Key(), ":")
				if !strings.EqualFold(namespace, test.namespace) {
					t.Errorf("key %q does not belong to namespace %q", m.Key(), test.namespace)
				}
			}
		})
	}
}

// Test_raw_002 checks that filter="artwork:"/"artwork:thumbnail" extracts
// the embedded preview image as a valid, decodable thumbnail.
func Test_raw_002(t *testing.T) {
	for _, filter := range []string{"artwork:", "artwork:thumbnail"} {
		t.Run(filter, func(t *testing.T) {
			path := filepath.Join(TEST_DIR, rawTestFile)
			contentType := contentTypeForFile(t, path)

			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			meta, err := metadata.GetMetadata(context.Background(), f, contentType, filter)
			if err != nil {
				// The generic image/* artwork handler also matches this
				// content type and is expected to fail decoding RAW bytes
				t.Logf("warning: %v", err)
			}
			if len(meta) != 1 {
				t.Fatalf("expected 1 artwork entry, got %d", len(meta))
			}
			assertArtwork(t, meta[0])
		})
	}
}
