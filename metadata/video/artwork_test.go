package video_test

import (
	"bytes"
	"context"
	"image"
	"os"
	"testing"

	// Packages
	metadata "github.com/mutablelogic/go-media/metadata"
	_ "github.com/mutablelogic/go-media/metadata/video"
)

const (
	testDir              = "../../etc/test"
	testFileArtwork      = "sample_with_artwork.mp4"
	testFileMultiArtwork = "sample_with_multi_artwork.mp4"
	testFileNoArtwork    = "sample.mp4"
)

// Test_artwork_000 checks that filter="artwork:" and "artwork:cover" both
// extract the embedded cover art as a valid, decodable image.
func Test_artwork_000(t *testing.T) {
	for _, filter := range []string{"artwork:", "artwork:cover"} {
		t.Run(filter, func(t *testing.T) {
			f, err := os.Open(testDir + "/" + testFileArtwork)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			contentType, _, err := metadata.ContentType(f)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.Seek(0, 0); err != nil {
				t.Fatal(err)
			}

			meta, err := metadata.GetMetadata(context.Background(), f, contentType, filter)
			if err != nil {
				t.Fatal(err)
			}
			if len(meta) != 1 {
				t.Fatalf("expected 1 artwork entry, got %d", len(meta))
			}

			m := meta[0]
			if m.Key() != "artwork:cover" {
				t.Errorf("Key() = %q, want %q", m.Key(), "artwork:cover")
			}

			data := m.Bytes()
			if len(data) == 0 {
				t.Fatal("Bytes() is empty")
			}
			decoded, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("Bytes() did not decode: %v", err)
			}

			img := m.Image()
			if img == nil {
				t.Fatal("Image() returned nil")
			}
			if img.Bounds() != decoded.Bounds() {
				t.Errorf("Image().Bounds() = %v, decoded Bytes() bounds = %v", img.Bounds(), decoded.Bounds())
			}
		})
	}
}

// Test_artwork_001 checks that a file with no embedded artwork returns no
// artwork entries and no error.
func Test_artwork_001(t *testing.T) {
	f, err := os.Open(testDir + "/" + testFileNoArtwork)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	contentType, _, err := metadata.ContentType(f)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	meta, err := metadata.GetMetadata(context.Background(), f, contentType, "artwork:cover")
	if err != nil {
		t.Fatal(err)
	}
	if len(meta) != 0 {
		t.Fatalf("expected no artwork entries, got %d", len(meta))
	}
}

// Test_artwork_002 checks that filters unrelated to artwork don't trigger
// cover art extraction, even on a file that has embedded artwork.
func Test_artwork_002(t *testing.T) {
	for _, filter := range []string{"", "dc:", "video:"} {
		t.Run(filter, func(t *testing.T) {
			f, err := os.Open(testDir + "/" + testFileArtwork)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			contentType, _, err := metadata.ContentType(f)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := f.Seek(0, 0); err != nil {
				t.Fatal(err)
			}

			meta, err := metadata.GetMetadata(context.Background(), f, contentType, filter)
			if err != nil {
				t.Fatal(err)
			}
			for _, m := range meta {
				if m.Key() == "artwork:cover" {
					t.Fatalf("did not expect artwork:cover for filter %q", filter)
				}
			}
		})
	}
}

// Test_artwork_003 checks that a file with more than one embedded picture
// (e.g. front and back cover) extracts all of them under distinct keys,
// and that a specific "artwork:cover-N" filter narrows to just that one.
func Test_artwork_003(t *testing.T) {
	open := func(t *testing.T) (*os.File, string) {
		t.Helper()
		f, err := os.Open(testDir + "/" + testFileMultiArtwork)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close() })
		contentType, _, err := metadata.ContentType(f)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Seek(0, 0); err != nil {
			t.Fatal(err)
		}
		return f, contentType
	}

	t.Run("artwork: returns both", func(t *testing.T) {
		f, contentType := open(t)
		meta, err := metadata.GetMetadata(context.Background(), f, contentType, "artwork:")
		if err != nil {
			t.Fatal(err)
		}
		if len(meta) != 2 {
			t.Fatalf("expected 2 artwork entries, got %d", len(meta))
		}
		keys := map[string]bool{}
		for _, m := range meta {
			keys[m.Key()] = true
			if len(m.Bytes()) == 0 {
				t.Errorf("%s: Bytes() is empty", m.Key())
			}
		}
		if !keys["artwork:cover"] || !keys["artwork:cover-2"] {
			t.Fatalf("expected keys artwork:cover and artwork:cover-2, got %v", meta)
		}
	})

	t.Run("artwork:cover-2 returns only the second", func(t *testing.T) {
		f, contentType := open(t)
		meta, err := metadata.GetMetadata(context.Background(), f, contentType, "artwork:cover-2")
		if err != nil {
			t.Fatal(err)
		}
		if len(meta) != 1 || meta[0].Key() != "artwork:cover-2" {
			t.Fatalf("expected only artwork:cover-2, got %v", meta)
		}
	})
}
