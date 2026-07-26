package task_test

import (
	"testing"

	// Packages
	task "github.com/mutablelogic/go-media/gomedia/task"
)

func TestProbeRequest_Query(t *testing.T) {
	req := task.ProbeRequest{Format: "mpegts", Opts: []string{"a=1", "b=2"}}
	query := req.Query()

	if got := query.Get("format"); got != "mpegts" {
		t.Fatalf("format = %q, want %q", got, "mpegts")
	}
	if got := query["opts"]; len(got) != 2 || got[0] != "a=1" || got[1] != "b=2" {
		t.Fatalf("opts = %v, want [a=1 b=2]", got)
	}
}

func TestProbeRequest_Query_Empty(t *testing.T) {
	query := task.ProbeRequest{}.Query()
	if len(query) != 0 {
		t.Fatalf("expected empty query, got %v", query)
	}
}
