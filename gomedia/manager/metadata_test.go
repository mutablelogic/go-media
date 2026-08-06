package manager_test

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"testing/iotest"

	// Packages
	manager "github.com/mutablelogic/go-media/gomedia/manager"
	taskmanager "github.com/mutablelogic/go-media/task/manager"
	taskmetadata "github.com/mutablelogic/go-media/task/metadata"
	taskschema "github.com/mutablelogic/go-media/task/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS

// TestMetadata_NoTaskManager guards against a nil m.opt.taskManager (e.g. a
// Media constructed without WithTaskManager) causing Metadata to panic
// instead of returning an error.
func TestMetadata_NoTaskManager(t *testing.T) {
	media, err := manager.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	req := taskmetadata.MetadataRequest{Reader: strings.NewReader("hello")}
	if _, err := media.Metadata(context.Background(), req); err == nil {
		t.Fatal("expected an error when no task manager is configured")
	}
}

// TestMetadata_RemovesFailedTaskFromManager is a regression test: Wait
// returns status.Err as its own error whenever the task itself failed, not
// just when waiting was interrupted, so an early "if err != nil { return }"
// right after Wait used to skip the cleanup Remove call entirely - a failed
// metadata task would linger in the task manager forever, visible via
// `gomedia tasks`. The request must fail during Run, not at Add (a nil
// reader is now rejected there, by MetadataRequest.Validate, before a task
// even exists to clean up) - a reader whose first Read errors gets past
// Validate but still fails Run's very first statement, schema.NewReadSeeker,
// so this is still a genuinely fatal, pre-result case, unlike
// TestMetadata_ReturnsPartialResultOnWarning.
func TestMetadata_RemovesFailedTaskFromManager(t *testing.T) {
	taskMgr, err := taskmanager.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = taskMgr.Run(runCtx, slog.Default()) }()
	<-taskMgr.Ready()

	media, err := manager.New(context.Background(), manager.WithTaskManager(taskMgr))
	if err != nil {
		t.Fatal(err)
	}

	before, err := taskMgr.ListTasks(context.Background(), taskschema.TaskListRequest{})
	if err != nil {
		t.Fatal(err)
	}

	req := taskmetadata.MetadataRequest{Reader: iotest.ErrReader(errors.New("boom"))}
	if _, err := media.Metadata(context.Background(), req); err == nil {
		t.Fatal("expected an error for a reader that fails on Read")
	}

	after, err := taskMgr.ListTasks(context.Background(), taskschema.TaskListRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if after.Count != before.Count {
		t.Fatalf("task count = %d, want %d (failed task should have been removed)", after.Count, before.Count)
	}
}

// TestMetadata_ReturnsPartialResultOnWarning: a task can set a result and
// still return an error (e.g. content type and basic tags were extracted
// fine, but no handler is registered to go any further) - Metadata should
// return that partial result rather than discarding it just because the
// task also errored. The failed-but-completed task must still be removed
// from the manager afterward, same as any other finished task.
func TestMetadata_ReturnsPartialResultOnWarning(t *testing.T) {
	taskMgr, err := taskmanager.New(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = taskMgr.Run(runCtx, slog.Default()) }()
	<-taskMgr.Ready()

	media, err := manager.New(context.Background(), manager.WithTaskManager(taskMgr))
	if err != nil {
		t.Fatal(err)
	}

	before, err := taskMgr.ListTasks(context.Background(), taskschema.TaskListRequest{})
	if err != nil {
		t.Fatal(err)
	}

	// Plain text has no registered metadata handler (only audio/video/image/
	// application do), so Run fails on its second pass with "not
	// implemented: no handler for content type text/plain" - but its first
	// pass (content type detection) already succeeded and set a result.
	req := taskmetadata.MetadataRequest{Reader: strings.NewReader("plain text, no media handler")}
	result, err := media.Metadata(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error (a warning, not a failure), got %v", err)
	}
	if result == nil {
		t.Fatal("expected a non-nil result despite the task's internal error")
	}
	if result.Type != "text/plain" {
		t.Fatalf("Type = %q, want %q", result.Type, "text/plain")
	}

	after, err := taskMgr.ListTasks(context.Background(), taskschema.TaskListRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if after.Count != before.Count {
		t.Fatalf("task count = %d, want %d (task should have been removed)", after.Count, before.Count)
	}
}
