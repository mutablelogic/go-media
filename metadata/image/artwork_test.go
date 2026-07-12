package image_test

import (
	"bytes"
	"context"
	"image"
	"os"
	"path/filepath"
	"testing"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	metadata "github.com/mutablelogic/go-media/metadata"
	_ "github.com/mutablelogic/go-media/metadata/image"
)

// Test_artwork_000 checks that filter="artwork:" and "artwork:thumbnail"
// both return a single, valid artwork entry.
func Test_artwork_000(t *testing.T) {
	for _, filter := range []string{"artwork:", "artwork:thumbnail"} {
		t.Run(filter, func(t *testing.T) {
			path := filepath.Join(TEST_DIR, "sample.jpg")
			contentType := contentTypeForFile(t, path)

			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			meta, err := metadata.GetMetadata(context.Background(), f, contentType, filter)
			if err != nil {
				t.Fatal(err)
			}
			if len(meta) != 1 {
				t.Fatalf("expected 1 artwork entry, got %d", len(meta))
			}
			assertArtwork(t, meta[0])
		})
	}
}

// Test_artwork_001 checks that filters unrelated to artwork don't trigger
// artwork extraction.
func Test_artwork_001(t *testing.T) {
	for _, filter := range []string{"", "tiff:", "image:"} {
		t.Run(filter, func(t *testing.T) {
			path := filepath.Join(TEST_DIR, "sample.jpg")
			contentType := contentTypeForFile(t, path)

			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			meta, err := metadata.GetMetadata(context.Background(), f, contentType, filter)
			if err != nil {
				t.Fatal(err)
			}
			for _, m := range meta {
				if m.Key() == "artwork:thumbnail" {
					t.Fatalf("did not expect artwork:thumbnail for filter %q", filter)
				}
			}
		})
	}
}

// Test_artwork_002 checks that an already-jpeg image near the target width
// is returned unchanged (the fast path), byte-for-byte identical to the
// source file, rather than being needlessly re-encoded.
func Test_artwork_002(t *testing.T) {
	path := filepath.Join(TEST_DIR, "sample.jpg")
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	contentType := contentTypeForFile(t, path)

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	meta, err := metadata.GetMetadata(context.Background(), f, contentType, "artwork:")
	if err != nil {
		t.Fatal(err)
	}
	if len(meta) != 1 {
		t.Fatalf("expected 1 artwork entry, got %d", len(meta))
	}
	if meta[0].Value() != "image/jpeg" {
		t.Errorf("Value() = %q, want %q", meta[0].Value(), "image/jpeg")
	}
	if !bytes.Equal(meta[0].Bytes(), want) {
		t.Error("expected the fast path to return the original file bytes unchanged")
	}
}

// Test_artwork_003 checks that an oversized image is resized down to at
// most MaxWidth, preserving aspect ratio, and re-encoded as PNG.
func Test_artwork_003(t *testing.T) {
	path := filepath.Join(TEST_DIR, "sample.png")
	contentType := contentTypeForFile(t, path)

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	meta, err := metadata.GetMetadata(context.Background(), f, contentType, "artwork:")
	if err != nil {
		t.Fatal(err)
	}
	if len(meta) != 1 {
		t.Fatalf("expected 1 artwork entry, got %d", len(meta))
	}
	if meta[0].Value() != "image/png" {
		t.Errorf("Value() = %q, want %q", meta[0].Value(), "image/png")
	}

	img := assertArtwork(t, meta[0])
	if got := img.Bounds().Dx(); got > 640 {
		t.Errorf("width = %d, want <= 640", got)
	}
	// sample.png is 1280x720 (16:9); the resize should preserve that ratio
	if got, want := img.Bounds().Dy(), img.Bounds().Dx()*720/1280; got != want {
		t.Errorf("height = %d, want %d (aspect ratio not preserved)", got, want)
	}
}

// Test_artwork_004 checks that small non-jpeg/png images (gif, bmp, tiff)
// are still re-encoded to PNG, even though they're already under MaxWidth
// and don't need resizing.
func Test_artwork_004(t *testing.T) {
	for _, file := range []string{"sample.gif", "sample.bmp", "sample.tiff"} {
		t.Run(file, func(t *testing.T) {
			path := filepath.Join(TEST_DIR, file)
			contentType := contentTypeForFile(t, path)

			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			meta, err := metadata.GetMetadata(context.Background(), f, contentType, "artwork:")
			if err != nil {
				t.Fatal(err)
			}
			if len(meta) != 1 {
				t.Fatalf("expected 1 artwork entry, got %d", len(meta))
			}
			if meta[0].Value() != "image/png" {
				t.Errorf("Value() = %q, want %q", meta[0].Value(), "image/png")
			}
			assertArtwork(t, meta[0])
		})
	}
}

// assertArtwork checks that an artwork metadata entry has a decodable,
// non-empty Bytes() payload and a non-nil Image() with matching bounds,
// and returns the decoded image.
func assertArtwork(t *testing.T, m gomedia.Metadata) image.Image {
	t.Helper()

	if m.Key() != "artwork:thumbnail" {
		t.Errorf("Key() = %q, want %q", m.Key(), "artwork:thumbnail")
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
	return img
}
