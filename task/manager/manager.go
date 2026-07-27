package manager

import (
	"context"
	"log/slog"
	"slices"
	"sync"
	"time"

	// Packages
	uuid "github.com/google/uuid"
	otel "github.com/mutablelogic/go-client/pkg/otel"
	gomedia "github.com/mutablelogic/go-media"
	schema "github.com/mutablelogic/go-media/task/schema"
	attribute "go.opentelemetry.io/otel/attribute"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Manager is an in-memory registry of tasks: Add one to get back a UUID,
// then Start, Cancel or Remove it by that UUID.
//
// Add, Start, Cancel, Wait and Remove all require Run to be actively running
// - the entry bookkeeping they touch is only meaningful while Run's
// shutdown-cancellation loop is watching it - and return an error otherwise.
type Manager struct {
	opt
	sync.Mutex
	tasks   map[uuid.UUID]*entry
	order   []uuid.UUID   // insertion order, so List is deterministic
	running bool          // set by Run once it's watching tasks, cleared once it stops
	ready   chan struct{} // closed once Run sets running - lets callers (tests, startup code) wait for it
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new media object
func New(ctx context.Context, opts ...Opt) (_ *Manager, err error) {
	self := new(Manager)
	if err := self.apply(opts); err != nil {
		return nil, err
	} else {
		self.tasks = make(map[uuid.UUID]*entry)
		self.order = make([]uuid.UUID, 0)
		self.ready = make(chan struct{})
	}

	// Return the task manager
	return self, nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Ready returns a channel that's closed once Run has started - Add, Start,
// Cancel, Wait and Remove aren't usable until then.
func (m *Manager) Ready() <-chan struct{} {
	return m.ready
}

// Run the task manager until ctx is cancelled, at which point it cancels
// every task still running and waits for each to finish (i.e. for its Run
// goroutine to return) or for ctx to be done a second time (e.g. a shutdown
// deadline), whichever comes first. Must be called exactly once.
func (m *Manager) Run(ctx context.Context, log *slog.Logger) error {
	log.DebugContext(ctx, "Starting task manager")

	m.Lock()
	m.running = true
	m.Unlock()
	close(m.ready)

	<-ctx.Done()
	log.DebugContext(ctx, "Stopping task manager")

	m.Lock()
	m.running = false
	order := slices.Clone(m.order)
	m.Unlock()

	pending := make([]<-chan struct{}, 0, len(order))
	for _, id := range order {
		m.Lock()
		e, exists := m.tasks[id]
		m.Unlock()
		if !exists {
			continue
		}

		e.Lock()
		if e.Cancel() {
			pending = append(pending, e.done)
		}
		e.Unlock()
	}

	// Wait for pending tasks to finish
	for _, done := range pending {
		select {
		case <-done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Return success
	return nil
}

// Add registers task under name and returns its UUID. The task isn't run
// until Start is called with that UUID.
func (m *Manager) Add(ctx context.Context, name string, task schema.Task) (_ uuid.UUID, err error) {
	_, endSpan := otel.StartSpan(m.tracer, ctx, "Add",
		attribute.String("name", name),
	)
	defer func() { endSpan(err) }()

	if err := m.checkRunning(); err != nil {
		return uuid.UUID{}, err
	}
	if task == nil {
		return uuid.UUID{}, gomedia.ErrBadParameter.With("nil task")
	}

	id := uuid.New()

	m.Lock()
	defer m.Unlock()
	m.tasks[id] = &entry{
		task:   task,
		status: schema.Status{UUID: id, Name: name},
		done:   make(chan struct{}),
	}
	m.order = append(m.order, id)

	return id, nil
}

// Start runs the task registered under id in a new goroutine, using ctx as
// the parent for cancellation and tracing, and returns immediately - use
// Cancel to stop the task, and Wait to block until it finishes.
func (m *Manager) Start(ctx context.Context, id uuid.UUID) (err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "Start",
		attribute.String("uuid", id.String()),
	)
	defer func() { endSpan(err) }()

	if err := m.checkRunning(); err != nil {
		return err
	}

	e, err := m.entry(id)
	if err != nil {
		return err
	}

	e.Lock()
	defer e.Unlock()
	if !e.status.Started.IsZero() {
		return gomedia.ErrBadParameter.Withf("task %q already started", id)
	}

	runCtx, cancel := context.WithCancel(ctx)
	e.status.Started = time.Now()
	e.cancel = cancel

	go func() {
		runErr := e.task.Run(schema.Context{
			Context: runCtx,
			Progress: func(current, total uint64) {
				e.Lock()
				defer e.Unlock()
				e.status.Progress = schema.Progress{Current: current, Total: total}
			},
			Result: func(result any) {
				e.Lock()
				defer e.Unlock()
				e.status.Result = result
			},
		})

		e.Lock()
		e.status.Finished = time.Now()
		e.status.Err = runErr
		e.Unlock()

		close(e.done)
	}()

	return nil
}

// Cancel stops the running task registered under id. It's a no-op if the
// task has already finished or was never started.
func (m *Manager) Cancel(ctx context.Context, id uuid.UUID) (err error) {
	_, endSpan := otel.StartSpan(m.tracer, ctx, "Cancel",
		attribute.String("uuid", id.String()),
	)
	defer func() { endSpan(err) }()

	if err := m.checkRunning(); err != nil {
		return err
	}

	e, err := m.entry(id)
	if err != nil {
		return err
	}

	e.Lock()
	defer e.Unlock()
	e.Cancel()

	return nil
}

// Remove unregisters the task registered under id, so it's no longer
// tracked by the manager. It fails if the task is still running - Cancel it
// and Wait for it to finish before removing it.
func (m *Manager) Remove(ctx context.Context, id uuid.UUID) (err error) {
	_, endSpan := otel.StartSpan(m.tracer, ctx, "Remove",
		attribute.String("uuid", id.String()),
	)
	defer func() { endSpan(err) }()

	if err := m.checkRunning(); err != nil {
		return err
	}

	e, err := m.entry(id)
	if err != nil {
		return err
	}

	e.Lock()
	running := !e.status.Started.IsZero() && e.status.Finished.IsZero()
	e.Unlock()
	if running {
		return gomedia.ErrBadParameter.Withf("task %q is still running", id)
	}

	m.Lock()
	defer m.Unlock()
	delete(m.tasks, id)
	m.order = slices.DeleteFunc(m.order, func(other uuid.UUID) bool {
		return other == id
	})

	return nil
}

// Wait blocks until the task registered under id finishes, or until ctx is
// done, whichever comes first, then returns its final status. It returns an
// error if the task hasn't been started (there's nothing to wait for), or if
// the task itself returned an error (joined into err, alongside a context
// error if ctx is what ended the wait).
func (m *Manager) Wait(ctx context.Context, id uuid.UUID) (_ *schema.Status, err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "Wait",
		attribute.String("uuid", id.String()),
	)
	defer func() { endSpan(err) }()

	if err := m.checkRunning(); err != nil {
		return nil, err
	}

	e, err := m.entry(id)
	if err != nil {
		return nil, err
	}

	e.Lock()
	started := !e.status.Started.IsZero()
	finished := !e.status.Finished.IsZero()
	done := e.done
	e.Unlock()

	if !started {
		return nil, gomedia.ErrBadParameter.Withf("task %q has not been started", id)
	}
	if !finished {
		select {
		case <-done:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	e.Lock()
	status := e.status
	e.Unlock()

	return &status, status.Err
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// checkRunning reports an error unless Run is actively watching tasks.
func (m *Manager) checkRunning() error {
	m.Lock()
	defer m.Unlock()

	if !m.running {
		return gomedia.ErrBadParameter.With("task manager is not running")
	}

	return nil
}

func (m *Manager) entry(id uuid.UUID) (*entry, error) {
	m.Lock()
	defer m.Unlock()

	e, exists := m.tasks[id]
	if !exists {
		return nil, gomedia.ErrNotFound.Withf("task %q not found", id)
	}

	return e, nil
}
