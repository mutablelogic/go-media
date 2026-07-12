package test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"

	// Packages
	manager "github.com/mutablelogic/go-media/gomedia/manager"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	shared  *manager.Media
	cancels cancelRegistry
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type cancelRegistry struct {
	mu      sync.Mutex
	cancels map[*testing.T]context.CancelFunc
}

func (r *cancelRegistry) Store(t *testing.T, cancel context.CancelFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancels == nil {
		r.cancels = make(map[*testing.T]context.CancelFunc)
	}
	r.cancels[t] = cancel
}

func (r *cancelRegistry) LoadAndDelete(t *testing.T) (context.CancelFunc, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancels == nil {
		return nil, false
	}
	cancel, ok := r.cancels[t]
	if ok {
		delete(r.cancels, t)
	}
	return cancel, ok
}

func (r *cancelRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, cancel := range r.cancels {
		cancel()
	}
	r.cancels = nil
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Main is the test main function for tests. It starts up a container and runs the tests,
// providing a manager instance to each test.
func Main(m *testing.M, setup func(*manager.Media) (func(), error), opts ...manager.Opt) {
	media, err := manager.New(context.Background(), opts...)
	if err != nil {
		panic(err)
	}
	shared = media
	runCtx, runCancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() {
		runDone <- manager.Run(runCtx, slog.Default())
	}()

	teardown := func() {}
	if setup != nil {
		if teardown_, err := setup(media); err != nil {
			panic(err)
		} else if teardown_ != nil {
			teardown = teardown_
		}
	}

	code := m.Run()

	// Best-effort cleanup for tests that forgot to call End.
	cancels.Clear()
	runCancel()
	if err := <-runDone; err != nil && !errors.Is(err, context.Canceled) {
		panic(err)
	}
	shared = nil
	teardown()

	os.Exit(code)
}

// Begin returns the shared test manager and a per-test context.
func Begin(t *testing.T) (*manager.Media, context.Context) {
	t.Helper()
	if shared == nil {
		t.Fatal("test manager is not initialized; call test.Main from TestMain")
	}
	base := context.Background()
	baseCancel := func() {}
	if deadline, ok := t.Deadline(); ok {
		base, baseCancel = context.WithDeadline(base, deadline)
	}
	ctx, cancel := context.WithCancel(base)
	stop := func() {
		cancel()
		baseCancel()
	}
	cancels.Store(t, context.CancelFunc(stop))
	t.Cleanup(func() {
		if cancel, ok := cancels.LoadAndDelete(t); ok {
			cancel()
		}
	})
	return shared, ctx
}

// End releases the per-test context created by Begin.
func End(t *testing.T) {
	t.Helper()
	if cancel, ok := cancels.LoadAndDelete(t); ok {
		cancel()
	}
}
