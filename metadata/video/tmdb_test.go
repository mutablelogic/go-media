package video

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	// Packages
	metadata "github.com/mutablelogic/go-media/metadata"
)

// namedReader wraps a Reader with a Name(), the way an uploaded or on-disk
// file is presented to metadata handlers via gomedia.NamedReader. The TMDB
// handler doesn't need real video bytes: it derives its search query from
// the filename alone.
type namedReader struct {
	io.Reader
	name string
}

func (r namedReader) Name() string { return r.name }

// Test_tmdb_000 checks that the TMDB handler looks up a movie by the title
// derived from the input's filename and returns its metadata, via the
// public GetMetadata entry point. Requires a live TMDB_TOKEN; skipped if it
// isn't set.
func Test_tmdb_000(t *testing.T) {
	token := os.Getenv("TMDB_TOKEN")
	if token == "" {
		t.Skip("TMDB_TOKEN not set")
	}

	r := namedReader{Reader: strings.NewReader(""), name: "Hunt for Red October.mp4"}

	// Requesting the "dc" namespace also selects the generic ffmpeg-backed
	// video handler (registered under "dc", "video"), which fails on this
	// fake reader's empty, non-video content - a warning alongside the
	// TMDB handler's own metadata, not a reason to fail the test.
	meta, err := metadata.GetMetadata(context.Background(), r, "video/mp4", metadata.WithTMDB(token), metadata.WithNamespace("tmdb", "dc"))
	if err != nil {
		t.Logf("warning: %v", err)
	}

	got := make(map[string]string, len(meta))
	for _, m := range meta {
		got[m.Key()] = m.Value()
	}

	if want := "The Hunt for Red October"; got["tmdb:title"] != want {
		t.Errorf("tmdb:title = %q, want %q", got["tmdb:title"], want)
	}
	if want := "The Hunt for Red October"; got["dc:title"] != want {
		t.Errorf("dc:title = %q, want %q", got["dc:title"], want)
	}
}
