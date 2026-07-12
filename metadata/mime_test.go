package metadata_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	// Packages
	. "github.com/mutablelogic/go-media/metadata"
)

const (
	TEST_DIR = "../etc/test"
)

type namedReader struct {
	*bytes.Reader
	name string
}

func (r namedReader) Name() string {
	return r.name
}

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

func Test_mime_001_m4a_override(t *testing.T) {
	// MP4 family signature that usually sniffs as video/mp4.
	data := []byte{0x00, 0x00, 0x00, 0x20, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm'}
	r := namedReader{Reader: bytes.NewReader(data), name: "audio.m4a"}

	contentType, _, err := ContentType(r)
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "audio/mp4" {
		t.Fatalf("expected audio/mp4, got %q", contentType)
	}
}

func Test_mime_002_extensionByType(t *testing.T) {
	if got := ExtensionByType("image/jpeg"); got != ".jpg" {
		t.Fatalf("ExtensionByType(image/jpeg) = %q, want %q", got, ".jpg")
	}
	if got := ExtensionByType("application/x-not-a-real-type"); got != "" {
		t.Fatalf("ExtensionByType(unknown) = %q, want empty", got)
	}
}
