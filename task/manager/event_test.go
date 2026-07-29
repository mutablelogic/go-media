package manager_test

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	// Packages
	manager "github.com/mutablelogic/go-media/task/manager"
	schema "github.com/mutablelogic/go-media/task/schema"
	require "github.com/stretchr/testify/require"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// eventRecorder collects events emitted via Subscribe - Start's goroutine
// (and its Progress/Result/Finished callbacks) run on a different goroutine
// than the test, so access must be synchronized.
type eventRecorder struct {
	mu     sync.Mutex
	events []*schema.Event
}

func (r *eventRecorder) fn(e *schema.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
}

func (r *eventRecorder) names() []schema.EventName {
	r.mu.Lock()
	defer r.mu.Unlock()
	names := make([]schema.EventName, len(r.events))
	for i, e := range r.events {
		names[i] = e.Name
	}
	return names
}

func (r *eventRecorder) len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.events)
}

////////////////////////////////////////////////////////////////////////////////
// HELPERS

// newRunningManager creates a Manager and starts Run in the background,
// returning once it's ready. Callers get back the cancel func for Run's own
// context, to shut it down at the end of the test.
func newRunningManager(t *testing.T) (*manager.Manager, context.Context, context.CancelFunc) {
	t.Helper()
	require := require.New(t)

	mgr, err := manager.New(context.Background())
	require.NoError(err)

	runCtx, cancel := context.WithCancel(context.Background())
	go func() { _ = mgr.Run(runCtx, slog.Default()) }()
	<-mgr.Ready()

	return mgr, context.Background(), cancel
}

// subscribe runs mgr.Subscribe(fn) in the background under its own
// cancellable context, and confirms - via throwaway warm-up tasks, purely as
// a synchronization signal - that the subscription is actually registered
// before returning. Without that handshake, a caller's very next Add/Start
// could race the subscriber goroutine and silently miss its event, since
// launching the goroutine only means Subscribe has been called, not that
// it's finished registering fn yet.
//
// The handshake itself would have the same race if it only tried once:
// adding a single warm-up task and waiting for it can just as easily run
// before the goroutine finishes registering. So it retries on a short
// timeout until wrapped is actually invoked (proving registration
// happened), which real wall-clock time between attempts guarantees
// eventually occurs. Warm-up events are filtered out by name unconditionally
// (not just before the handshake completes), so a later subscribe call's
// retries can never leak into an earlier, already-active subscriber sharing
// the same Manager.
//
// Returns a func to unsubscribe (cancel that context and wait for Subscribe
// to actually return).
func subscribe(t *testing.T, mgr *manager.Manager, ctx context.Context, fn func(*schema.Event)) (unsubscribe func() error) {
	t.Helper()
	require := require.New(t)

	var active atomic.Bool
	probe := make(chan struct{}, 1)

	wrapped := func(e *schema.Event) {
		if !active.Load() {
			select {
			case probe <- struct{}{}:
			default:
			}
			return
		}
		if e.Status.Name == "warmup" {
			return
		}
		fn(e)
	}

	subCtx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- mgr.Subscribe(subCtx, wrapped) }()

	registered := false
	deadline := time.Now().Add(time.Second)
	for !registered && time.Now().Before(deadline) {
		_, err := mgr.Add(ctx, "warmup", &fakeTask{done: make(chan struct{})})
		require.NoError(err)

		select {
		case <-probe:
			registered = true
		case <-time.After(10 * time.Millisecond):
		}
	}
	require.True(registered, "subscription was not registered in time")
	active.Store(true)

	return func() error {
		cancel()
		return <-done
	}
}

////////////////////////////////////////////////////////////////////////////////
// TESTS

func TestManager_EventFn(t *testing.T) {
	require := require.New(t)
	mgr, ctx, cancel := newRunningManager(t)
	defer cancel()

	recorder := &eventRecorder{}
	unsubscribe := subscribe(t, mgr, ctx, recorder.fn)
	defer unsubscribe()

	ft := &fakeTask{done: make(chan struct{})}

	id, err := mgr.Add(ctx, "probe", ft)
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))

	close(ft.done)
	_, err = mgr.Wait(ctx, id)
	require.NoError(err)

	// Finished is emitted just after the done channel closes, so Wait can
	// return a beat before the event actually lands - give it a moment.
	require.Eventually(func() bool { return recorder.len() >= 5 }, time.Second, time.Millisecond)

	require.NoError(mgr.Remove(ctx, id))
	require.Eventually(func() bool { return recorder.len() >= 6 }, time.Second, time.Millisecond)

	require.Equal([]schema.EventName{
		schema.EventAdded,
		schema.EventStarted,
		schema.EventProgress,
		schema.EventResult,
		schema.EventFinished,
		schema.EventRemoved,
	}, recorder.names())
}

func TestManager_EventFnCancel(t *testing.T) {
	require := require.New(t)
	mgr, ctx, cancel := newRunningManager(t)
	defer cancel()

	recorder := &eventRecorder{}
	unsubscribe := subscribe(t, mgr, ctx, recorder.fn)
	defer unsubscribe()

	ft := &fakeTask{done: make(chan struct{})}
	defer close(ft.done)

	id, err := mgr.Add(ctx, "probe", ft)
	require.NoError(err)
	require.NoError(mgr.Start(ctx, id))
	require.NoError(mgr.Cancel(ctx, id))

	// Cancelling runCtx already unblocks fakeTask via its ctx.Done() case, so
	// ft.done is never closed here - only by the deferred cleanup above.
	_, err = mgr.Wait(ctx, id)
	require.Error(err)

	require.Eventually(func() bool { return recorder.len() >= 4 }, time.Second, time.Millisecond)

	names := recorder.names()
	require.Contains(names, schema.EventCancelled)
	require.Contains(names, schema.EventFinished)
	// Cancelled (synchronous, from Manager.Cancel) must be observed before
	// Finished (from the task's Run goroutine actually returning).
	require.Less(indexOf(names, schema.EventCancelled), indexOf(names, schema.EventFinished))
}

func TestManager_SubscribeNilFn(t *testing.T) {
	require := require.New(t)
	mgr, _, cancel := newRunningManager(t)
	defer cancel()

	require.Error(mgr.Subscribe(context.Background(), nil))
}

func TestManager_SubscribeMultiple(t *testing.T) {
	require := require.New(t)
	mgr, ctx, cancel := newRunningManager(t)
	defer cancel()

	a, b := &eventRecorder{}, &eventRecorder{}
	unsubA := subscribe(t, mgr, ctx, a.fn)
	defer unsubA()
	unsubB := subscribe(t, mgr, ctx, b.fn)
	defer unsubB()

	_, err := mgr.Add(ctx, "probe", &fakeTask{done: make(chan struct{})})
	require.NoError(err)

	require.Eventually(func() bool { return a.len() >= 1 && b.len() >= 1 }, time.Second, time.Millisecond)
	require.Equal([]schema.EventName{schema.EventAdded}, a.names())
	require.Equal([]schema.EventName{schema.EventAdded}, b.names())
}

func TestManager_SubscribeContextCancelUnsubscribes(t *testing.T) {
	require := require.New(t)
	mgr, ctx, cancel := newRunningManager(t)
	defer cancel()

	recorder := &eventRecorder{}
	unsubscribe := subscribe(t, mgr, ctx, recorder.fn)

	// Unsubscribing must both return Subscribe (with the cancellation error)
	// and stop delivering further events to fn.
	require.ErrorIs(unsubscribe(), context.Canceled)

	_, err := mgr.Add(ctx, "probe", &fakeTask{done: make(chan struct{})})
	require.NoError(err)

	// Give a would-be delivery a moment to land, then confirm it didn't.
	time.Sleep(20 * time.Millisecond)
	require.Equal(0, recorder.len())
}

func TestManager_SubscribeUnblocksWhenManagerStops(t *testing.T) {
	require := require.New(t)
	mgr, err := manager.New(context.Background())
	require.NoError(err)

	runCtx, runCancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() { runDone <- mgr.Run(runCtx, slog.Default()) }()
	<-mgr.Ready()

	// Subscribe with a context that's never cancelled directly - only
	// stopping the Manager itself should unblock it.
	subDone := make(chan error, 1)
	go func() { subDone <- mgr.Subscribe(context.Background(), func(*schema.Event) {}) }()

	runCancel()

	select {
	case err := <-subDone:
		require.NoError(err)
	case <-time.After(time.Second):
		t.Fatal("Subscribe did not unblock when the Manager stopped")
	}

	// No tasks were registered, so Run's own shutdown sweep has nothing to
	// wait for and returns cleanly once ctx is done.
	require.NoError(<-runDone)
}

////////////////////////////////////////////////////////////////////////////////
// HELPERS

func indexOf(names []schema.EventName, name schema.EventName) int {
	for i, n := range names {
		if n == name {
			return i
		}
	}
	return -1
}
