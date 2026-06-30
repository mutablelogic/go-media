package raw_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/mutablelogic/go-media/pkg/raw"
)

const testRAW = "../../etc/test/RAW_OLYMPUS_E3.ORF"

////////////////////////////////////////////////////////////////////////////////
// OPEN / READ / PARSE

func Test_raw_000(t *testing.T) {
	r, err := raw.Open(testRAW)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
}

func Test_raw_001(t *testing.T) {
	_, err := raw.Open("/does/not/exist.orf")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	t.Log(err)
}

func Test_raw_002(t *testing.T) {
	buf, err := os.ReadFile(testRAW)
	if err != nil {
		t.Fatal(err)
	}
	r, err := raw.Parse(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
}

func Test_raw_003(t *testing.T) {
	f, err := os.Open(testRAW)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	r, err := raw.Read(f)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
}

////////////////////////////////////////////////////////////////////////////////
// ACCESSORS

func Test_raw_010(t *testing.T) {
	r, err := raw.Open(testRAW)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	t.Logf("make=%q model=%q software=%q", r.Make(), r.Model(), r.Software())
	if r.Make() == "" {
		t.Error("expected non-empty make")
	}
	if r.Model() == "" {
		t.Error("expected non-empty model")
	}
}

func Test_raw_011(t *testing.T) {
	r, err := raw.Open(testRAW)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	t.Logf("iso=%.0f shutter=%v aperture=%.1f focal_len=%.0fmm",
		r.ISOSpeed(), r.Shutter(), r.Aperture(), r.FocalLength())
	t.Logf("timestamp=%v", r.Timestamp())
}

func Test_raw_012(t *testing.T) {
	r, err := raw.Open(testRAW)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	t.Logf("raw=%dx%d image=%dx%d", r.RawWidth(), r.RawHeight(), r.Width(), r.Height())
	if r.RawWidth() == 0 || r.RawHeight() == 0 {
		t.Error("expected non-zero raw dimensions")
	}
}

////////////////////////////////////////////////////////////////////////////////
// METADATA

func Test_raw_020(t *testing.T) {
	r, err := raw.Open(testRAW)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	meta := r.Metadata()
	if len(meta) == 0 {
		t.Fatal("expected at least one metadata entry")
	}
	for _, m := range meta {
		t.Logf("key=%q value=%q typed=%T(%v)", m.Key(), m.Value(), m.Any(), m.Any())
	}
}

func Test_raw_021(t *testing.T) {
	r, err := raw.Open(testRAW)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	for _, m := range r.Metadata() {
		data, err := json.Marshal(m)
		if err != nil {
			t.Fatalf("key=%q: %v", m.Key(), err)
		}
		t.Log(string(data))
	}
}

func Test_raw_022(t *testing.T) {
	r, err := raw.Open(testRAW)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	xmp := r.XMP()
	if xmp != nil {
		t.Logf("xmp length=%d bytes", len(xmp))
	} else {
		t.Log("no embedded XMP")
	}
}

////////////////////////////////////////////////////////////////////////////////
// THUMBNAIL

func Test_raw_030(t *testing.T) {
	r, err := raw.Open(testRAW)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	img, err := r.Thumbnail()
	if err != nil {
		t.Fatal(err)
	}
	bounds := img.Bounds()
	t.Logf("thumbnail size=%dx%d", bounds.Dx(), bounds.Dy())
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		t.Error("expected non-zero thumbnail dimensions")
	}
}

func Test_raw_031(t *testing.T) {
	r, err := raw.Open(testRAW)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	data, err := r.ThumbnailBytes()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty thumbnail bytes")
	}
	t.Logf("thumbnail bytes=%d", len(data))
}

////////////////////////////////////////////////////////////////////////////////
// FULL IMAGE

func Test_raw_040(t *testing.T) {
	r, err := raw.Open(testRAW)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	img, err := r.Image()
	if err != nil {
		t.Fatal(err)
	}
	bounds := img.Bounds()
	t.Logf("image size=%dx%d", bounds.Dx(), bounds.Dy())
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		t.Error("expected non-zero image dimensions")
	}
}
