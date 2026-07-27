package manager_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	// Packages
	task "github.com/mutablelogic/go-media/gomedia/task"
	test "github.com/mutablelogic/go-media/gomedia/test"
)

////////////////////////////////////////////////////////////////////////////////
// HELPERS

func sampleFilePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "etc", "test", name)
}

////////////////////////////////////////////////////////////////////////////////
// TESTS

func TestProbeMedia(t *testing.T) {
	m, ctx := test.Begin(t)
	defer test.End(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	resp, err := m.ProbeMedia(ctx, task.ProbeMediaRequest{Reader: f})
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected a non-nil response")
	}
	if resp.Name != f.Name() {
		t.Fatalf("Name = %q, want %q", resp.Name, f.Name())
	}
	if resp.Format == nil {
		t.Fatal("expected a non-nil Format")
	}
	if len(resp.Streams) == 0 {
		t.Fatal("expected at least one stream")
	}
}

// Cover art is demuxed as a video stream with the attached-pic disposition,
// which reader.Streams() (and so ProbeResponse.Streams) excludes - see
// reader_test.go's TestReader_Streams_ExcludesArtwork. This just checks
// probing a file with embedded artwork works at all, since it exercises a
// different demux path (an extra attached-pic stream) than a plain sample.
func TestProbeMedia_WithArtwork(t *testing.T) {
	m, ctx := test.Begin(t)
	defer test.End(t)

	f, err := os.Open(sampleFilePath(t, "sample_with_artwork.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	resp, err := m.ProbeMedia(ctx, task.ProbeMediaRequest{Reader: f})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Streams) == 0 {
		t.Fatal("expected at least one stream")
	}
}

func TestProbeMedia_NilReader(t *testing.T) {
	m, ctx := test.Begin(t)
	defer test.End(t)

	if _, err := m.ProbeMedia(ctx, task.ProbeMediaRequest{Reader: nil}); err == nil {
		t.Fatal("expected an error for a nil reader")
	}
}

func TestProbeMedia_InvalidData(t *testing.T) {
	m, ctx := test.Begin(t)
	defer test.End(t)

	req := task.ProbeMediaRequest{Reader: strings.NewReader("not a real media file")}
	if _, err := m.ProbeMedia(ctx, req); err == nil {
		t.Fatal("expected an error for invalid data")
	}
}
