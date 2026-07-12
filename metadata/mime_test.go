package metadata_test

import (
	"os"
	"path/filepath"
	"testing"

	// Packages
	. "github.com/mutablelogic/go-media/metadata"
)

const (
	TEST_DIR = "../etc/test"
)

func Test_mime_000(t *testing.T) {
	entries, err := os.ReadDir(TEST_DIR)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			f, err := os.Open(filepath.Join(TEST_DIR, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			contentType, meta, err := ContentType(f)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("%s: %s %v", entry.Name(), contentType, meta)
		})
	}
}
