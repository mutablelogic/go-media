package manager_test

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	test "github.com/mutablelogic/go-media/gomedia/test"
)

func TestSegmentAudio_Input(t *testing.T) {
	m, ctx := test.Begin(t)

	testFile := filepath.Join("..", "..", "etc", "test", "sample.mp3")
	f, err := os.Open(testFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	out := t.TempDir()

	if err := m.SegmentAudio(ctx, schema.SegmentAudioRequest{Reader: f, OutputDir: out}); err != nil {
		t.Fatal(err)
	}

	files, err := filepath.Glob(path.Join(out, "*.m4a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("expected at least one m4a segment output")
	}
}
