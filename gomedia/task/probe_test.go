package task_test

import (
	"context"
	"testing"

	// Packages
	task "github.com/mutablelogic/go-media/gomedia/task"
)

func TestProbeMediaRequest_Query(t *testing.T) {
	req := task.ProbeMediaRequest{ProbeRequestOpts: task.ProbeRequestOpts{Format: "mpegts", Opts: []string{"a=1", "b=2"}}}
	query := req.Query()

	if got := query.Get("format"); got != "mpegts" {
		t.Fatalf("format = %q, want %q", got, "mpegts")
	}
	if got := query["opts"]; len(got) != 2 || got[0] != "a=1" || got[1] != "b=2" {
		t.Fatalf("opts = %v, want [a=1 b=2]", got)
	}
}

func TestProbeMediaRequest_Query_Empty(t *testing.T) {
	query := task.ProbeMediaRequest{}.Query()
	if len(query) != 0 {
		t.Fatalf("expected empty query, got %v", query)
	}
}

// A scheme FFmpeg has no registered protocol for must be rejected before
// ever reaching AVFormat_open_url, so the error is a clear "unsupported
// scheme" rather than a confusing low-level one.
func TestProbeSourceTask_UnsupportedScheme(t *testing.T) {
	tk, err := task.NewProbeSourceTask(task.ProbeSourceRequest{
		Url: "bogus-scheme://example.com",
	})
	if err != nil {
		t.Fatalf("NewProbeSourceTask: %v", err)
	}
	if err := tk.Run(task.Context{Context: context.Background()}); err == nil {
		t.Fatal("Run: expected an error for an unsupported URL scheme")
	}
}

// "device://<format>/<address>" must carry both a format (the host) and an
// address (the path) - missing either is a caller error, not something
// that should reach reader.Open.
func TestProbeSourceTask_DeviceURL_MissingFormat(t *testing.T) {
	tk, err := task.NewProbeSourceTask(task.ProbeSourceRequest{
		Url: "device:///0:0",
	})
	if err != nil {
		t.Fatalf("NewProbeSourceTask: %v", err)
	}
	if err := tk.Run(task.Context{Context: context.Background()}); err == nil {
		t.Fatal("Run: expected an error for a device URL with no format")
	}
}

func TestProbeSourceTask_DeviceURL_MissingAddress(t *testing.T) {
	tk, err := task.NewProbeSourceTask(task.ProbeSourceRequest{
		Url: "device://avfoundation",
	})
	if err != nil {
		t.Fatalf("NewProbeSourceTask: %v", err)
	}
	if err := tk.Run(task.Context{Context: context.Background()}); err == nil {
		t.Fatal("Run: expected an error for a device URL with no address")
	}
}
