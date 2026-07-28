package httpclient_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	// Packages
	task "github.com/mutablelogic/go-media/gomedia/task"
	test "github.com/mutablelogic/go-media/gomedia/test"
	taskmetadata "github.com/mutablelogic/go-media/task/metadata"
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

func TestProbeMedia_FormData(t *testing.T) {
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
	resp, err := c.ProbeMedia(ctx, task.ProbeMediaRequest{Reader: f}, types.ContentTypeFormData, func(n int64) { lastRead.Store(n) })
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

func TestProbeMedia_RawBody(t *testing.T) {
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
	resp, err := c.ProbeMedia(ctx, task.ProbeMediaRequest{Reader: f}, "audio/mpeg", func(n int64) { lastRead.Store(n) })
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

func TestProbeMedia_InvalidData(t *testing.T) {
	_, ctx := test.Begin(t)
	defer test.End(t)
	c := test.Client(t)

	req := strings.NewReader("not a real media file")
	if _, err := c.ProbeMedia(ctx, task.ProbeMediaRequest{Reader: req}, "audio/mpeg", nil); err == nil {
		t.Fatal("expected an error for invalid data")
	}
}

// ProbeSource's "file" scheme is a registered FFmpeg protocol (unlike
// "rtsp", which is a container format, not a protocol - see reader.Protocols'
// doc comment), so this exercises the full POST /probe/source round trip
// without needing network access or a real capture device.
func TestProbeSource_File(t *testing.T) {
	_, ctx := test.Begin(t)
	defer test.End(t)
	c := test.Client(t)

	abs, err := filepath.Abs(sampleFilePath(t, "sample.mp3"))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := c.ProbeSource(ctx, task.ProbeSourceRequest{Url: fmt.Sprintf("file://%s", abs)})
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
	if s := resp.Streams[0]; s.Par() == nil || s.Par().SampleRate() != 44100 {
		t.Fatalf("Streams[0].Par().SampleRate() = %v, want 44100", s.Par())
	}
}

// An unregistered scheme must be rejected server-side before FFmpeg ever
// sees it (see ProbeSourceTask's doc comment).
func TestProbeSource_UnsupportedScheme(t *testing.T) {
	_, ctx := test.Begin(t)
	defer test.End(t)
	c := test.Client(t)

	if _, err := c.ProbeSource(ctx, task.ProbeSourceRequest{Url: "bogus-scheme://example.com"}); err == nil {
		t.Fatal("expected an error for an unsupported URL scheme")
	}
}

func TestMetadata_FormData(t *testing.T) {
	_, ctx := test.Begin(t)
	defer test.End(t)
	c := test.Client(t)

	f, err := os.Open(sampleFilePath(t, "sample.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var lastRead atomic.Int64
	resp, err := c.Metadata(ctx, taskmetadata.MetadataRequest{Reader: f}, types.ContentTypeFormData, func(n int64) { lastRead.Store(n) })
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected a non-nil response")
	}
	if want := filepath.Base(f.Name()); resp.Name != want {
		t.Fatalf("Name = %q, want %q", resp.Name, want)
	}
	if resp.Type == "" {
		t.Fatal("expected a non-empty content type")
	}
	if got := lastRead.Load(); got <= 0 {
		t.Fatalf("onRead final count = %d, want > 0", got)
	}
}

func TestMetadata_RawBody(t *testing.T) {
	_, ctx := test.Begin(t)
	defer test.End(t)
	c := test.Client(t)

	// The raw-body path has no filename to fall back on, so content type
	// detection relies entirely on Go's stdlib byte-sniffing (see
	// metadata.ContentType) - unlike sample.mp3, which has no leading ID3
	// tag and so isn't in net/http's sniff table, a JPEG's magic bytes are
	// always recognized without needing a name.
	f, err := os.Open(sampleFilePath(t, "sample.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var lastRead atomic.Int64
	resp, err := c.Metadata(ctx, taskmetadata.MetadataRequest{Reader: f}, "image/jpeg", func(n int64) { lastRead.Store(n) })
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected a non-nil response")
	}
	if resp.Type != "image/jpeg" {
		t.Fatalf("Type = %q, want %q", resp.Type, "image/jpeg")
	}
	if len(resp.Metadata) == 0 {
		t.Fatal("expected at least one metadata entry")
	}
	if got := lastRead.Load(); got <= 0 {
		t.Fatalf("onRead final count = %d, want > 0", got)
	}
}

// Plain text has no registered metadata handler, so the server-side task
// fails on its second pass - but its first pass (content type detection)
// already succeeded, so Metadata treats that failure as a warning and still
// returns the partial result rather than a hard error.
func TestMetadata_InvalidData(t *testing.T) {
	_, ctx := test.Begin(t)
	defer test.End(t)
	c := test.Client(t)

	req := strings.NewReader("not a real media file")
	resp, err := c.Metadata(ctx, taskmetadata.MetadataRequest{Reader: req}, "audio/mpeg", nil)
	if err != nil {
		t.Fatalf("expected no error (a warning, not a failure), got %v", err)
	}
	if resp == nil {
		t.Fatal("expected a non-nil response")
	}
	if resp.Type != "text/plain" {
		t.Fatalf("Type = %q, want %q", resp.Type, "text/plain")
	}
}

func TestProbeSource_MissingURL(t *testing.T) {
	_, ctx := test.Begin(t)
	defer test.End(t)
	c := test.Client(t)

	if _, err := c.ProbeSource(ctx, task.ProbeSourceRequest{}); err == nil {
		t.Fatal("expected an error for a missing URL")
	}
}
