package test

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"

	// Packages
	manager "github.com/mutablelogic/go-media/profile/manager"
	pg "github.com/mutablelogic/go-pg"
	test "github.com/mutablelogic/go-pg/pkg/test"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// Set to connect to an already-running Postgres instead of starting a
	// Docker testcontainer - for CI on platforms without a fast Docker
	// daemon (macOS, Windows), mirroring how other mutablelogic projects
	// provision a native Postgres for tests on those runners.
	envTestDatabaseURL = "GOMEDIA_TEST_DATABASE_URL"
)

var (
	shared  *manager.Profile
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

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Main is the test main function for tests. It starts up a container (or
// connects to an existing database, if GOMEDIA_TEST_DATABASE_URL is set)
// and runs the tests, providing a manager instance to each test.
func Main(m *testing.M, setup func(*manager.Profile) (func(), error), opts ...manager.Opt) {
	run := func(pool pg.PoolConn) (func(), error) {
		profile, err := manager.New(context.Background(), pool, opts...)
		if err != nil {
			return nil, err
		}
		shared = profile

		teardown := func() {}
		if setup != nil {
			if teardown_, err := setup(profile); err != nil {
				return nil, err
			} else if teardown_ != nil {
				teardown = teardown_
			}
		}

		runCtx, cancel := context.WithCancel(context.Background())
		runDone := make(chan error, 1)
		go func() {
			runDone <- profile.Run(runCtx, slog.Default())
		}()
		return func() {
			cancel()
			if err := <-runDone; err != nil {
				panic(err)
			}
			shared = nil
			teardown()
		}, nil
	}

	if url := os.Getenv(envTestDatabaseURL); url != "" {
		os.Exit(mainWithURL(url, m, run))
		return
	}

	test.Main(m, run)
}

// mainWithURL connects to an already-running Postgres instead of starting a
// testcontainer, and otherwise follows the same setup/run/teardown sequence
// as go-pg/pkg/test.Main.
func mainWithURL(url string, m *testing.M, setup func(pg.PoolConn) (func(), error)) int {
	ctx := context.Background()
	pool, err := pg.NewPool(ctx, pg.WithURL(url))
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	cleanup := func() {}
	if setup != nil {
		cleanup_, err := setup(pool)
		if err != nil {
			panic(err)
		} else if cleanup_ != nil {
			cleanup = cleanup_
		}
	}
	defer cleanup()

	return m.Run()
}

// Begin returns the shared test manager and a per-test context.
func Begin(t *testing.T) (*manager.Profile, context.Context) {
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
