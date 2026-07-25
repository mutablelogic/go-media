package task

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"sync"
	"time"

	// Packages
	uuid "github.com/google/uuid"
	otel "github.com/mutablelogic/go-client/pkg/otel"
	gomedia "github.com/mutablelogic/go-media"
	attribute "go.opentelemetry.io/otel/attribute"
	trace "go.opentelemetry.io/otel/trace"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// State is the lifecycle state of a task tracked by a Manager.
type State string

const (
	StateNotStarted State = "not_started"
	StateRunning    State = "running"
	StateCancelled  State = "cancelled"
	StateError      State = "error"
	StateDone       State = "done"
)

// Progress reports how far a running task has got, in task-defined units
// (e.g. bytes, frames, streams). Total is 0 if not known in advance, in
// which case a percentage can't be computed.
type Progress struct {
	Current int64
	Total   int64
}

func (p Progress) valid() bool {
	return p.Total > 0
}

func (p Progress) percent() float64 {
	return float64(p.Current) / float64(p.Total) * 100
}

// Status is a snapshot of a task tracked by a Manager.
type Status struct {
	UUID      uuid.UUID
	Name      string
	Progress  Progress
	Result    any
	Started   time.Time // zero until the task has been run
	Finished  time.Time // zero until the task has finished
	Cancelled bool      // set by Cancel, regardless of the error the task returns
	Err       error
}

// State reports the task's current lifecycle state.
func (s Status) State() State {
	switch {
	case s.Started.IsZero():
		return StateNotStarted
	case s.Finished.IsZero():
		return StateRunning
	case s.Cancelled:
		return StateCancelled
	case s.Err != nil:
		return StateError
	default:
		return StateDone
	}
}

// Duration is how long the task has been running, or ran for if it has
// finished. Zero if the task hasn't started yet.
func (s Status) Duration() time.Duration {
	switch {
	case s.Started.IsZero():
		return 0
	case s.Finished.IsZero():
		return time.Since(s.Started)
	default:
		return s.Finished.Sub(s.Started)
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	type jsonStatus struct {
		UUID     uuid.UUID  `json:"uuid"`
		Name     string     `json:"name"`
		Status   State      `json:"status"`
		Started  *time.Time `json:"started,omitempty"`
		Duration string     `json:"duration,omitempty"`
		Percent  *float64   `json:"percent,omitempty"`
		Result   any        `json:"result,omitempty"`
		Error    string     `json:"error,omitempty"`
	}

	out := jsonStatus{
		UUID:   s.UUID,
		Name:   s.Name,
		Status: s.State(),
		Result: s.Result,
	}
	if !s.Started.IsZero() {
		started := s.Started
		out.Started = &started
		out.Duration = s.Duration().String()
	}
	if s.Progress.valid() {
		percent := s.Progress.percent()
		out.Percent = &percent
	}
	if s.Err != nil {
		out.Error = s.Err.Error()
	}

	return json.Marshal(out)
}

// TaskListRequest filters the tasks returned by Manager.List. A nil field
// means "don't filter on this".
type TaskListRequest struct {
	Name  *string `json:"name,omitempty"`
	State *State  `json:"state,omitempty"`
}

// TaskList is a snapshot of the tasks tracked by a Manager, in the order
// they were added.
type TaskList []Status

// entry is a task tracked by a Manager, plus the state Manager needs to run,
// cancel, and report on it.
type entry struct {
	sync.Mutex
	task   Task
	status Status
	cancel context.CancelFunc
	done   chan struct{} // closed once the task's Run goroutine returns
}

// cancelLocked cancels e if it's currently running, and reports whether it
// did so. The caller must hold e's lock.
func (e *entry) cancelLocked() bool {
	if e.cancel == nil || !e.status.Finished.IsZero() {
		return false
	}
	e.status.Cancelled = true
	e.cancel()
	return true
}

// Manager is an in-memory registry of tasks: Add one to get back a UUID,
// then Run or Cancel it by that UUID.
type Manager struct {
	tracer trace.Tracer

	sync.Mutex
	tasks map[uuid.UUID]*entry
	order []uuid.UUID // insertion order, so List is deterministic
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewManager creates a new, empty task manager. tracer can be nil.
func NewManager(tracer trace.Tracer) *Manager {
	return &Manager{
		tracer: tracer,
		tasks:  make(map[uuid.UUID]*entry),
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Add registers task under name and returns its UUID. The task isn't run
// until Run is called with that UUID.
func (m *Manager) Add(ctx context.Context, name string, task Task) (_ uuid.UUID, err error) {
	_, endSpan := otel.StartSpan(m.tracer, ctx, "Add",
		attribute.String("name", name),
	)
	defer func() { endSpan(err) }()

	if task == nil {
		return uuid.UUID{}, gomedia.ErrBadParameter.With("nil task")
	}

	id := uuid.New()

	m.Lock()
	defer m.Unlock()
	m.tasks[id] = &entry{
		task:   task,
		status: Status{UUID: id, Name: name},
		done:   make(chan struct{}),
	}
	m.order = append(m.order, id)

	return id, nil
}

// List returns a snapshot of the tasks matching req, in the order they were
// added.
func (m *Manager) List(ctx context.Context, req TaskListRequest) (_ TaskList, err error) {
	_, endSpan := otel.StartSpan(m.tracer, ctx, "List")
	defer func() { endSpan(err) }()

	m.Lock()
	order := slices.Clone(m.order)
	m.Unlock()

	result := make(TaskList, 0, len(order))
	for _, id := range order {
		m.Lock()
		e, exists := m.tasks[id]
		m.Unlock()
		if !exists {
			continue
		}

		e.Lock()
		status := e.status
		e.Unlock()

		if req.Name != nil && status.Name != *req.Name {
			continue
		}
		if req.State != nil && status.State() != *req.State {
			continue
		}

		result = append(result, status)
	}

	return result, nil
}

// Run starts the task registered under id in a new goroutine, using ctx as
// the parent for cancellation and tracing, and returns immediately - use
// Cancel to stop the task, and Status to poll its progress or result.
func (m *Manager) Run(ctx context.Context, id uuid.UUID) (err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "Run",
		attribute.String("uuid", id.String()),
	)
	defer func() { endSpan(err) }()

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
		runErr := e.task.Run(Context{
			Context: runCtx,
			Tracer:  m.tracer,
			Progress: func(current, total int64) {
				e.Lock()
				defer e.Unlock()
				e.status.Progress = Progress{Current: current, Total: total}
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

	e, err := m.entry(id)
	if err != nil {
		return err
	}

	e.Lock()
	defer e.Unlock()
	e.cancelLocked()

	return nil
}

// Close cancels every task that's still running, then waits for each to
// finish (i.e. for its Run goroutine to return) or for ctx to be done,
// whichever comes first - so a shutdown sequence can bound how long it
// waits by passing a context with a deadline. It's safe to call even if
// some or all tasks have already finished or were never started.
func (m *Manager) Close(ctx context.Context) (err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "Close")
	defer func() { endSpan(err) }()

	m.Lock()
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
		if e.cancelLocked() {
			pending = append(pending, e.done)
		}
		e.Unlock()
	}

	for _, done := range pending {
		select {
		case <-done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// Wait blocks until the task registered under id finishes, or until ctx is
// done, whichever comes first, then returns its final status. It returns an
// error if the task hasn't been started (there's nothing to wait for).
func (m *Manager) Wait(ctx context.Context, id uuid.UUID) (_ *Status, err error) {
	ctx, endSpan := otel.StartSpan(m.tracer, ctx, "Wait",
		attribute.String("uuid", id.String()),
	)
	defer func() { endSpan(err) }()

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

	err = errors.Join(err, status.Err)
	return &status, err
}

// Status returns a snapshot of the task registered under id.
func (m *Manager) Status(id uuid.UUID) (*Status, error) {
	e, err := m.entry(id)
	if err != nil {
		return nil, err
	}

	e.Lock()
	defer e.Unlock()
	status := e.status

	return &status, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (m *Manager) entry(id uuid.UUID) (*entry, error) {
	m.Lock()
	defer m.Unlock()

	e, exists := m.tasks[id]
	if !exists {
		return nil, gomedia.ErrNotFound.Withf("task %q not found", id)
	}

	return e, nil
}
