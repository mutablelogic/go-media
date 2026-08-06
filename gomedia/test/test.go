package test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	// Packages
	httpclient "github.com/mutablelogic/go-media/gomedia/httpclient"
	httphandler "github.com/mutablelogic/go-media/gomedia/httphandler"
	manager "github.com/mutablelogic/go-media/gomedia/manager"
	taskmanager "github.com/mutablelogic/go-media/task/manager"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	shared            *manager.Media
	sharedTaskManager *taskmanager.Manager
	sharedClient      *httpclient.Client
	sharedServer      *httptest.Server
	cancels           cancelRegistry
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
// providing a manager instance, and an HTTP client wired to a test server with
// the standard handlers registered, to each test.
func Main(m *testing.M, setup func(*manager.Media) (func(), error), opts ...manager.Opt) {
	taskMgr, err := taskmanager.New(context.Background())
	if err != nil {
		panic(err)
	}

	taskRunCtx, taskRunCancel := context.WithCancel(context.Background())
	taskRunDone := make(chan error, 1)
	go func() {
		taskRunDone <- taskMgr.Run(taskRunCtx, slog.Default())
	}()
	<-taskMgr.Ready()

	// A caller-supplied WithTaskManager (if any) takes precedence over this
	// default, since it's listed first and apply() applies options in order.
	opts = append([]manager.Opt{manager.WithTaskManager(taskMgr)}, opts...)
	sharedTaskManager = taskMgr

	media, err := manager.New(context.Background(), opts...)
	if err != nil {
		panic(err)
	}
	shared = media

	router, err := httprouter.NewRouter(context.Background(), http.NewServeMux(), "/", "", "Test API", "1.0.0")
	if err != nil {
		panic(err)
	}
	if err := httphandler.RegisterMetadataHandlers(media, router); err != nil {
		panic(err)
	}
	if err := httphandler.RegisterEncoderHandlers(media, taskMgr, router); err != nil {
		panic(err)
	}
	sharedServer = httptest.NewServer(router)
	sharedClient, err = httpclient.New(sharedServer.URL)
	if err != nil {
		panic(err)
	}

	runCtx, runCancel := context.WithCancel(context.Background())
	runDone := make(chan error, 1)
	go func() {
		runDone <- media.Run(runCtx, slog.Default())
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
	taskRunCancel()
	if err := <-taskRunDone; err != nil && !errors.Is(err, context.Canceled) {
		panic(err)
	}
	sharedServer.Close()
	shared = nil
	sharedTaskManager = nil
	sharedClient = nil
	sharedServer = nil
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

// TaskManager returns the shared task manager - useful for a test to wait on
// or inspect a task started by the Media under test (e.g. Media.Encode),
// which doesn't wait for completion itself.
func TaskManager(t *testing.T) *taskmanager.Manager {
	t.Helper()
	if sharedTaskManager == nil {
		t.Fatal("test task manager is not initialized; call test.Main from TestMain")
	}
	return sharedTaskManager
}

// Client returns the shared HTTP client, wired to a test server with the
// standard handlers registered.
func Client(t *testing.T) *httpclient.Client {
	t.Helper()
	if sharedClient == nil {
		t.Fatal("test client is not initialized; call test.Main from TestMain")
	}
	return sharedClient
}

// ServerURL returns the base URL of the shared test server - useful for a
// test that needs to make a raw HTTP request (e.g. a multipart upload with a
// custom Accept header) rather than going through Client.
func ServerURL(t *testing.T) string {
	t.Helper()
	if sharedServer == nil {
		t.Fatal("test server is not initialized; call test.Main from TestMain")
	}
	return sharedServer.URL
}

// End releases the per-test context created by Begin.
func End(t *testing.T) {
	t.Helper()
	if cancel, ok := cancels.LoadAndDelete(t); ok {
		cancel()
	}
}
