package image_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	// Packages
	metadata "github.com/mutablelogic/go-media/metadata"
	_ "github.com/mutablelogic/go-media/metadata/image"
)

const heifTestFile = "photo.HEIC"

func Test_heif_metadata_000(t *testing.T) {
	path := filepath.Join(TEST_DIR, heifTestFile)
	contentType := contentTypeForFile(t, path)
	if contentType != "image/heic" {
		t.Fatalf("content type = %q, want %q", contentType, "image/heic")
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	meta, err := metadata.GetMetadata(context.Background(), f, contentType, "tiff:")
	if err != nil {
		t.Fatal(err)
	}

	got := make(map[string]string, len(meta))
	for _, m := range meta {
		got[m.Key()] = m.Value()
	}

	if got["tiff:Make"] != "Apple" {
		t.Fatalf("tiff:Make = %q, want %q", got["tiff:Make"], "Apple")
	}
	if got["tiff:Model"] != "iPhone 11 Pro" {
		t.Fatalf("tiff:Model = %q, want %q", got["tiff:Model"], "iPhone 11 Pro")
	}
	if got["tiff:Software"] == "" {
		t.Fatal("expected tiff:Software to be present")
	}
}
