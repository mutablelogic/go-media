package manager_test

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	// Packages
	uuid "github.com/google/uuid"
	manager "github.com/mutablelogic/go-media/task/manager"
	schema "github.com/mutablelogic/go-media/task/schema"
	test "github.com/mutablelogic/go-media/task/test"
	require "github.com/stretchr/testify/require"
)

////////////////////////////////////////////////////////////////////////////////
// FAKE TASK

// fakeTask reports progress and a result once, then blocks until either its
// context is cancelled or the test signals it to finish (via done), and
// returns runErr. If ignoreCancel is set, it only unblocks on done, to
// simulate a task that doesn't respect cancellation.
type fakeTask struct {
	done         chan struct{}
	runErr       error
	ignoreCancel bool
}

func (t *fakeTask) Task() string {
	return "fake"
}

func (t *fakeTask) Validate() error {
	return nil
}

func (t *fakeTask) Run(ctx schema.Context) error {
	if ctx.Progress != nil {
		ctx.Progress(1, 2)
	}
	if ctx.Result != nil {
		ctx.Result("fake result")
	}
	if t.ignoreCancel {
		<-t.done
		return t.runErr
	}
	select {
	case <-t.done:
		return t.runErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

////////////////////////////////////////////////////////////////////////////////
// TESTS

func TestManager_AddReturnsUUID(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	id, err := mgr.Add(ctx, "probe", &fakeTask{done: make(chan struct{})})
	require.NoError(err)
	require.NotEqual(uuid.UUID{}, id)
}

func TestManager_AddNilTask(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	_, err := mgr.Add(ctx, "probe", nil)
	require.Error(err)
}

func TestManager_NotRunning(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	// A freshly constructed manager whose Run has never been called - every
	// method that touches task bookkeeping should refuse to work.
	mgr, err := manager.New(ctx)
	require.NoError(err)

	select {
	case <-mgr.Ready():
		t.Fatal("expected Ready to still be open before Run is called")
	default:
	}

	_, err = mgr.Add(ctx, "probe", &fakeTask{done: make(chan struct{})})
	require.Error(err)
	require.Error(mgr.Start(ctx, uuid.New()))
	require.Error(mgr.Cancel(ctx, uuid.New()))
	require.Error(mgr.Remove(ctx, uuid.New()))
	_, err = mgr.Wait(ctx, uuid.New())
	require.Error(err)
}

func TestManager_RunTwiceFails(t *testing.T) {
	require := require.New(t)

	mgr, err := manager.New(context.Background())
	require.NoError(err)

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runDone := make(chan error, 1)
	go func() { runDone <- mgr.Run(runCtx, slog.Default()) }()
	<-mgr.Ready()

	// Calling Run again while it's still running must fail rather than
	// panicking on a second close(m.ready).
	require.Error(mgr.Run(context.Background(), slog.Default()))

	cancel()
	require.NoError(<-runDone)

	// And after it's finished, it still refuses to run again.
	require.Error(mgr.Run(context.Background(), slog.Default()))
}

func TestManager_StartCompletes(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	ft := &fakeTask{done: make(chan struct{})}

	id, err := mgr.Add(ctx, "probe", ft)
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))

	close(ft.done)

	status, err := mgr.Wait(ctx, id)
	require.NoError(err)
	require.Equal(schema.StateDone, status.State())
	require.Equal(schema.Progress{Current: 1, Total: 2}, status.Progress)
	require.Equal("fake result", status.Result)
	require.Greater(status.Duration(), time.Duration(0))
}

// TestManager_StartOutlivesCallerContext is a regression test: a task's own
// execution must be scoped to Run's ctx, not whatever ctx Start happened to
// be called with - otherwise a caller whose own ctx ends shortly after
// Start returns (e.g. an HTTP request's context, once its handler returns
// without waiting for the task) would cut the task short, even though it's
// meant to keep running independently.
func TestManager_StartOutlivesCallerContext(t *testing.T) {
	require := require.New(t)
	mgr, _ := test.Begin(t)
	defer test.End(t)

	ft := &fakeTask{done: make(chan struct{})}

	id, err := mgr.Add(context.Background(), "probe", ft)
	require.NoError(err)

	// An already-cancelled ctx, scoping only this Start call - if Start
	// wrongly used it as the task's own execution context too, the task
	// would immediately observe ctx.Done() and return ctx.Err(), rather
	// than actually running until ft.done closes.
	startCtx, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(mgr.Start(startCtx, id))

	// Wait with a short deadline of its own - if the bug were present, the
	// task would already be finished (with context.Canceled as its error)
	// well before this expires, and Wait would return that immediately
	// instead of timing out.
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer waitCancel()
	_, err = mgr.Wait(waitCtx, id)
	require.ErrorIs(err, context.DeadlineExceeded)

	close(ft.done)
	status, err := mgr.Wait(context.Background(), id)
	require.NoError(err)
	require.Equal(schema.StateDone, status.State())
}

// TestManager_StartConcurrentWithGetTask is a regression test for a
// suspected lock-order inversion between Start (which briefly held a
// task's own lock while acquiring the Manager's) and GetTask/Cancel/
// Remove/Wait (which acquire the two locks sequentially, never nested) -
// see the comment in Start for the details. It hammers Start and GetTask
// concurrently across many tasks and requires the whole thing to finish
// well within a generous deadline; a reintroduced lock-order inversion
// would hang this test rather than fail an assertion, so the real
// assertion here is "this returns at all."
func TestManager_StartConcurrentWithGetTask(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	const n = 50
	ids := make([]uuid.UUID, n)
	tasks := make([]*fakeTask, n)
	for i := range ids {
		tasks[i] = &fakeTask{done: make(chan struct{})}
		id, err := mgr.Add(ctx, "probe", tasks[i])
		require.NoError(err)
		ids[i] = id
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		var wg sync.WaitGroup
		wg.Add(2 * n)
		for _, id := range ids {
			go func(id uuid.UUID) {
				defer wg.Done()
				_ = mgr.Start(ctx, id)
			}(id)
			go func(id uuid.UUID) {
				defer wg.Done()
				for j := 0; j < 20; j++ {
					_, _ = mgr.GetTask(ctx, id)
				}
			}(id)
		}
		wg.Wait()
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Start/GetTask deadlocked under concurrent load")
	}

	for _, task := range tasks {
		close(task.done)
	}
	for _, id := range ids {
		_, err := mgr.Wait(ctx, id)
		require.NoError(err)
	}
}

func TestManager_StartTwiceFails(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	ft := &fakeTask{done: make(chan struct{})}
	defer close(ft.done)

	id, err := mgr.Add(ctx, "probe", ft)
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))
	require.Error(mgr.Start(ctx, id))
}

func TestManager_Cancel(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	ft := &fakeTask{done: make(chan struct{})}
	defer close(ft.done)

	id, err := mgr.Add(ctx, "probe", ft)
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))
	require.NoError(mgr.Cancel(ctx, id))

	status, err := mgr.Wait(ctx, id)
	require.True(errors.Is(err, context.Canceled))
	require.Equal(schema.StateCancelled, status.State())
}

func TestManager_RemoveNotStartedSucceeds(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	id, err := mgr.Add(ctx, "probe", &fakeTask{done: make(chan struct{})})
	require.NoError(err)
	require.NoError(mgr.Remove(ctx, id))

	// It's gone - Cancel (or any other by-id method) now reports not found.
	require.Error(mgr.Cancel(ctx, id))
}

func TestManager_RemoveRunningTaskFails(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	ft := &fakeTask{done: make(chan struct{})}
	defer close(ft.done)

	id, err := mgr.Add(ctx, "probe", ft)
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))

	// Start sets Started synchronously before the task's goroutine even
	// begins running, so the task is already "running" as far as Remove is
	// concerned the moment Start returns - no need to wait for anything.
	require.Error(mgr.Remove(ctx, id))
}

func TestManager_RemoveAfterFinishSucceeds(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	ft := &fakeTask{done: make(chan struct{})}
	id, err := mgr.Add(ctx, "probe", ft)
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))
	close(ft.done)

	_, err = mgr.Wait(ctx, id)
	require.NoError(err)
	require.NoError(mgr.Remove(ctx, id))

	// It's gone - Cancel (or any other by-id method) now reports not found.
	require.Error(mgr.Cancel(ctx, id))
}

func TestManager_NotFound(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	id := uuid.New()

	require.Error(mgr.Start(ctx, id))
	require.Error(mgr.Cancel(ctx, id))
	require.Error(mgr.Remove(ctx, id))
	_, err := mgr.Wait(ctx, id)
	require.Error(err)
}

func TestManager_WaitNotStartedFails(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	ft := &fakeTask{done: make(chan struct{})}
	defer close(ft.done)

	id, err := mgr.Add(ctx, "probe", ft)
	require.NoError(err)

	_, err = mgr.Wait(ctx, id)
	require.Error(err)
}

func TestManager_WaitBlocksUntilFinished(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	ft := &fakeTask{done: make(chan struct{})}
	id, err := mgr.Add(ctx, "probe", ft)
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))

	go func() {
		time.Sleep(10 * time.Millisecond)
		close(ft.done)
	}()

	status, err := mgr.Wait(ctx, id)
	require.NoError(err)
	require.Equal(schema.StateDone, status.State())
	require.Equal("fake result", status.Result)

	// Waiting again on an already-finished task should return immediately
	// with the same status.
	status, err = mgr.Wait(ctx, id)
	require.NoError(err)
	require.Equal(schema.StateDone, status.State())
}

func TestManager_WaitReturnsTaskError(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	wantErr := errors.New("boom")
	ft := &fakeTask{done: make(chan struct{}), runErr: wantErr}
	id, err := mgr.Add(ctx, "probe", ft)
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))
	close(ft.done)

	status, err := mgr.Wait(ctx, id)
	require.True(errors.Is(err, wantErr))
	require.Equal(schema.StateError, status.State())
}

func TestManager_WaitTimesOut(t *testing.T) {
	require := require.New(t)
	mgr, ctx := test.Begin(t)
	defer test.End(t)

	ft := &fakeTask{done: make(chan struct{}), ignoreCancel: true}
	defer close(ft.done)

	id, err := mgr.Add(ctx, "stubborn", ft)
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))

	waitCtx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()

	_, err = mgr.Wait(waitCtx, id)
	require.True(errors.Is(err, context.DeadlineExceeded))
}
