package writer_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/url"
	"os"
	"testing"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	profile "github.com/mutablelogic/go-media/profile/schema"
	writer "github.com/mutablelogic/go-media/writer"
)

//////////////////////////////////////////////////////////////////////////////
// HELPERS

// artworkMeta is a minimal gomedia.Metadata implementation carrying artwork
// (image) data under the gomedia.MetaArtwork key.
type artworkMeta struct {
	data []byte
}

func (a artworkMeta) Key() string        { return gomedia.MetaArtwork }
func (a artworkMeta) Value() string      { return "" }
func (a artworkMeta) Bytes() []byte      { return a.data }
func (a artworkMeta) Image() image.Image { return nil }
func (a artworkMeta) Any() any           { return a.data }

// textMeta is a minimal gomedia.Metadata implementation for a plain text tag.
type textMeta struct {
	key, value string
}

func (m textMeta) Key() string        { return m.key }
func (m textMeta) Value() string      { return m.value }
func (m textMeta) Bytes() []byte      { return []byte(m.value) }
func (m textMeta) Image() image.Image { return nil }
func (m textMeta) Any() any           { return m.value }

func testJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("jpeg.Encode: %v", err)
	}
	return buf.Bytes()
}

func testPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{B: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
	return buf.Bytes()
}

//////////////////////////////////////////////////////////////////////////////
// TESTS

func TestCreate_WithArtwork(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	w, err := writer.Create(tempFileURL(t, "out.mp4"), output,
		writer.WithProfile(0, audioStream(t)),
		writer.WithMetadata(artworkMeta{data: testJPEG(t)}),
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestCreate_WithArtwork_Multiple(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	w, err := writer.Create(tempFileURL(t, "out.mp4"), output,
		writer.WithProfile(0, audioStream(t)),
		writer.WithMetadata(
			artworkMeta{data: testJPEG(t)},
			artworkMeta{data: testPNG(t)},
		),
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestCreate_WithArtwork_InvalidImage(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	_, err := writer.Create(tempFileURL(t, "out.mp4"), output,
		writer.WithProfile(0, audioStream(t)),
		writer.WithMetadata(artworkMeta{data: []byte("not an image")}),
	)
	if err == nil {
		t.Fatal("Create: expected error for invalid artwork image data")
	}
}

func TestCreate_WithMetadata_TextEntry(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	w, err := writer.Create(tempFileURL(t, "out.mp4"), output,
		writer.WithProfile(0, audioStream(t)),
		writer.WithMetadata(textMeta{key: "title", value: "Test Title"}),
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

// TestCreate_WithArtwork_EncodeFrame confirms artwork is actually muxed into
// the output (not just accepted at Create time) by comparing file size
// against a baseline encoded without artwork, and that encoding real frames
// afterward still works correctly (exercising the lazy write-artwork-on-
// first-packet path in writePacket).
func TestCreate_WithArtwork_EncodeFrame(t *testing.T) {
	output := profile.OutputWithName("mp4")
	if output == nil {
		t.Fatal("OutputWithName(mp4): nil output")
	}

	encodeOne := func(path string, opts ...writer.Opt) int64 {
		t.Helper()
		w, err := writer.Create(&url.URL{Path: path}, output, opts...)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		numSamples := w.FrameSize(0)
		if numSamples == 0 {
			numSamples = 1024
		}
		frame := silentFrame(t, 0, numSamples)
		frame.SetPts(0)
		if err := w.Encode(frame); err != nil {
			frame.Close()
			t.Fatalf("Encode: %v", err)
		}
		frame.Close()

		if err := w.Flush(0); err != nil {
			t.Fatalf("Flush: %v", err)
		}
		if err := w.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat: %v", err)
		}
		return info.Size()
	}

	baselinePath := tempFileURL(t, "baseline.mp4").Path
	baselineSize := encodeOne(baselinePath, writer.WithProfile(0, audioStream(t)))

	artworkPath := tempFileURL(t, "artwork.mp4").Path
	artworkSize := encodeOne(artworkPath,
		writer.WithProfile(0, audioStream(t)),
		writer.WithMetadata(artworkMeta{data: testJPEG(t)}),
	)

	if artworkSize <= baselineSize {
		t.Fatalf("expected artwork output (%d bytes) to be larger than baseline (%d bytes)", artworkSize, baselineSize)
	}
}
