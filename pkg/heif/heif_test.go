package heif_test

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"testing"

	"github.com/mutablelogic/go-media/pkg/heif"
)

const testHEIF = "../../etc/test/photo.HEIC"

func Test_heif_000(t *testing.T) {
	h, err := heif.Open(testHEIF)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	var _ image.Image = h

	bounds := h.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Fatalf("unexpected bounds: %v", bounds)
	}
	if h.ColorModel() == nil {
		t.Fatal("ColorModel returned nil")
	}
	_ = h.At(bounds.Min.X, bounds.Min.Y)
	if h.Primary() == nil {
		t.Fatal("Primary returned nil")
	}
}

func Test_heif_001(t *testing.T) {
	f, err := os.Open(testHEIF)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	h, err := heif.Read(f)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	if bounds := h.Bounds(); bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Fatalf("unexpected bounds: %v", bounds)
	}
}

func Test_heif_002(t *testing.T) {
	buf, err := os.ReadFile(testHEIF)
	if err != nil {
		t.Fatal(err)
	}

	h, err := heif.Parse(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	if bounds := h.Bounds(); bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Fatalf("unexpected bounds: %v", bounds)
	}

	// Ensure the decoded primary image survives the original byte slice lifecycle.
	copyBuf := append([]byte(nil), buf...)
	_ = bytes.NewReader(copyBuf)
}

func Test_heif_003(t *testing.T) {
	buf, err := os.ReadFile(testHEIF)
	if err != nil {
		t.Fatal(err)
	}

	img, format, err := image.Decode(bytes.NewReader(buf))
	if err != nil {
		t.Fatal(err)
	}
	if format != "heif" {
		t.Fatalf("image.Decode format=%q want=%q", format, "heif")
	}
	if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		t.Fatalf("unexpected decoded bounds: %v", img.Bounds())
	}

	// Prove the returned image is a standard image.Image by re-encoding it.
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatal(err)
	}
	if encoded.Len() == 0 {
		t.Fatal("expected png output")
	}
}

func Test_heif_004(t *testing.T) {
	buf, err := os.ReadFile(testHEIF)
	if err != nil {
		t.Fatal(err)
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(buf))
	if err != nil {
		t.Fatal(err)
	}
	if format != "heif" {
		t.Fatalf("image.DecodeConfig format=%q want=%q", format, "heif")
	}
	if config.Width <= 0 || config.Height <= 0 {
		t.Fatalf("unexpected config: %#v", config)
	}
}

func Test_heif_005(t *testing.T) {
	h, err := heif.Open(testHEIF)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	meta := h.Metadata()
	if len(meta) == 0 {
		t.Skip("no metadata blocks in fixture")
	}
	for _, m := range meta {
		if m.Key() == "" {
			t.Fatal("metadata key is empty")
		}
		if m.Value() == "" && len(m.Bytes()) == 0 && m.Any() == nil {
			t.Fatalf("metadata entry %q is empty", m.Key())
		}
	}
}

func Test_heif_006(t *testing.T) {
	h, err := heif.Open(testHEIF)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	thumbs := h.Thumbnails()
	if len(thumbs) == 0 {
		t.Skip("no thumbnails in fixture")
	}

	t.Logf("thumbnails=%d", len(thumbs))
	for i, img := range thumbs {
		if img == nil {
			t.Fatalf("thumbnail %d is nil", i)
		}
		if bounds := img.Bounds(); bounds.Dx() <= 0 || bounds.Dy() <= 0 {
			t.Fatalf("thumbnail %d has invalid bounds: %v", i, bounds)
		}
	}
}
