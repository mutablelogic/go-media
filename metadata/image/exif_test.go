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

const (
	TEST_DIR = "../../etc/test"
)

// Test_image_000 walks etc/test, and for every file with at least one
// registered handler for its content type (JPEG, RAW, or any other
// image/* format), extracts metadata from it.
func Test_image_000(t *testing.T) {
	entries, err := os.ReadDir(TEST_DIR)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(TEST_DIR, entry.Name())
		contentType := contentTypeForFile(t, path)
		if len(metadata.GetHandlers(contentType)) == 0 {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			meta, err := metadata.GetMetadata(context.Background(), f, contentType, "")
			if err != nil {
				// A handler error is a warning, not fatal: other handlers
				// for the same content type may still have succeeded.
				t.Logf("warning: %v", err)
			}
			for _, m := range meta {
				t.Logf("%s = %v", m.Key(), m.Value())
			}
			if len(meta) == 0 {
				t.Errorf("expected metadata to be extracted from %s, got none", entry.Name())
			}
		})
	}
}

func contentTypeForFile(t *testing.T, path string) string {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	contentType, _, err := metadata.ContentType(f)
	if err != nil {
		t.Fatal(err)
	}
	return contentType
}
