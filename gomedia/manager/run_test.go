package manager

import (
	"context"
	"log/slog"
	"testing"
	"time"

	// Packages
	task "github.com/mutablelogic/go-media/gomedia/task"
)

// fakeTask blocks until its context is cancelled, then returns ctx.Err().
type fakeTask struct{}

func (fakeTask) Run(ctx task.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

// TestRun_CancelsTasksOnShutdown checks that Run, once its context is
// cancelled, cancels any task still running on the Media's task manager and
// waits for it to actually stop before returning.
func TestRun_CancelsTasksOnShutdown(t *testing.T) {
	m, err := New(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	id, err := m.tasks.Add(context.Background(), "fake", fakeTask{})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.tasks.Run(context.Background(), id); err != nil {
		t.Fatal(err)
	}

	runCtx, cancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() {
		runDone <- m.Run(runCtx, slog.Default())
	}()

	cancel()

	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("Run returned %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after its context was cancelled")
	}

	status, err := m.tasks.Status(id)
	if err != nil {
		t.Fatal(err)
	}
	if status.State() != task.StateCancelled {
		t.Fatalf("task state = %v, want %v", status.State(), task.StateCancelled)
	}
}
