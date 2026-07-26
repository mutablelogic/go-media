package httpclient_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	// Packages
	task "github.com/mutablelogic/go-media/gomedia/task"
	test "github.com/mutablelogic/go-media/gomedia/test"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// HELPERS

func sampleFilePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "etc", "test", name)
}

////////////////////////////////////////////////////////////////////////////////
// TESTS

func TestProbe_FormData(t *testing.T) {
	_, ctx := test.Begin(t)
	defer test.End(t)
	c := test.Client(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	// onRead is invoked from whatever goroutine is writing the request body
	// (net/http streams a chunked upload on its own goroutine, decoupled
	// from the one that returns the response), so lastRead needs atomic
	// access rather than a plain variable.
	var lastRead atomic.Int64
	resp, err := c.Probe(ctx, task.ProbeRequest{Reader: f}, types.ContentTypeFormData, func(n int64) { lastRead.Store(n) })
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected a non-nil response")
	}
	// Multipart uploads only carry the filename's basename, not the full
	// path (standard multipart convention, not something this client/server
	// pair controls).
	if want := filepath.Base(f.Name()); resp.Name != want {
		t.Fatalf("Name = %q, want %q", resp.Name, want)
	}
	if resp.Format == nil {
		t.Fatal("expected a non-nil Format")
	}
	if len(resp.Streams) == 0 {
		t.Fatal("expected at least one stream")
	}
	// Regression check: decoding a probe response over HTTP used to
	// silently zero out every stream (StreamProfile had no UnmarshalJSON),
	// which re-marshaled as a bogus zero-dimension video stream since a
	// zero AVCodecParameters.codec_type collides with AVMEDIA_TYPE_VIDEO.
	if s := resp.Streams[0]; s.Par() == nil || s.Par().SampleRate() != 44100 {
		t.Fatalf("Streams[0].Par().SampleRate() = %v, want 44100", s.Par())
	}
	// FFmpeg's probing only reads as much of the stream as it needs to
	// identify the format/streams (governed by probesize), which for a
	// format like mp3 can be less than the whole file - so onRead's final
	// count isn't expected to reach info.Size(), just to have reported some
	// progress.
	if got := lastRead.Load(); got <= 0 || got > info.Size() {
		t.Fatalf("onRead final count = %d, want in (0, %d]", got, info.Size())
	}
}

func TestProbe_RawBody(t *testing.T) {
	_, ctx := test.Begin(t)
	defer test.End(t)
	c := test.Client(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	var lastRead atomic.Int64
	resp, err := c.Probe(ctx, task.ProbeRequest{Reader: f}, "audio/mpeg", func(n int64) { lastRead.Store(n) })
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected a non-nil response")
	}
	if resp.Format == nil {
		t.Fatal("expected a non-nil Format")
	}
	if len(resp.Streams) == 0 {
		t.Fatal("expected at least one stream")
	}
	// Regression check: decoding a probe response over HTTP used to
	// silently zero out every stream (StreamProfile had no UnmarshalJSON),
	// which re-marshaled as a bogus zero-dimension video stream since a
	// zero AVCodecParameters.codec_type collides with AVMEDIA_TYPE_VIDEO.
	if s := resp.Streams[0]; s.Par() == nil || s.Par().SampleRate() != 44100 {
		t.Fatalf("Streams[0].Par().SampleRate() = %v, want 44100", s.Par())
	}
	// FFmpeg's probing only reads as much of the stream as it needs to
	// identify the format/streams (governed by probesize), which for a
	// format like mp3 can be less than the whole file - so onRead's final
	// count isn't expected to reach info.Size(), just to have reported some
	// progress.
	if got := lastRead.Load(); got <= 0 || got > info.Size() {
		t.Fatalf("onRead final count = %d, want in (0, %d]", got, info.Size())
	}
}

func TestProbe_InvalidData(t *testing.T) {
	_, ctx := test.Begin(t)
	defer test.End(t)
	c := test.Client(t)

	req := strings.NewReader("not a real media file")
	if _, err := c.Probe(ctx, task.ProbeRequest{Reader: req}, "audio/mpeg", nil); err == nil {
		t.Fatal("expected an error for invalid data")
	}
}
