package task_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	// Packages
	uuid "github.com/google/uuid"
	task "github.com/mutablelogic/go-media/gomedia/task"
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

func (t *fakeTask) Run(ctx task.Context) error {
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
	m := task.NewManager(nil)

	id, err := m.Add(context.Background(), "probe", &fakeTask{done: make(chan struct{})})
	if err != nil {
		t.Fatal(err)
	}
	if id == (uuid.UUID{}) {
		t.Fatal("expected a non-zero uuid")
	}
}

func TestManager_AddNilTask(t *testing.T) {
	m := task.NewManager(nil)

	if _, err := m.Add(context.Background(), "probe", nil); err == nil {
		t.Fatal("expected an error for a nil task")
	}
}

func TestManager_RunCompletes(t *testing.T) {
	m := task.NewManager(nil)

	ft := &fakeTask{done: make(chan struct{})}
	id, err := m.Add(context.Background(), "probe", ft)
	if err != nil {
		t.Fatal(err)
	}

	status, err := m.Status(id)
	if err != nil {
		t.Fatal(err)
	}
	if status.State() != task.StateNotStarted {
		t.Fatalf("state = %v, want %v", status.State(), task.StateNotStarted)
	}

	if err := m.Run(context.Background(), id); err != nil {
		t.Fatal(err)
	}

	// Wait for the progress report.
	waitFor(t, func() bool {
		status, err := m.Status(id)
		return err == nil && status.Progress.Total == 2
	})

	status, err = m.Status(id)
	if err != nil {
		t.Fatal(err)
	}
	if status.State() != task.StateRunning {
		t.Fatalf("state = %v, want %v", status.State(), task.StateRunning)
	}
	if status.Progress != (task.Progress{Current: 1, Total: 2}) {
		t.Fatalf("progress = %+v, want {1 2}", status.Progress)
	}
	if status.Result != "fake result" {
		t.Fatalf("result = %v, want %q", status.Result, "fake result")
	}

	close(ft.done)

	waitFor(t, func() bool {
		status, err := m.Status(id)
		return err == nil && status.State() != task.StateRunning
	})

	status, err = m.Status(id)
	if err != nil {
		t.Fatal(err)
	}
	if status.State() != task.StateDone {
		t.Fatalf("state = %v, want %v", status.State(), task.StateDone)
	}
	if status.Err != nil {
		t.Fatalf("err = %v, want nil", status.Err)
	}
	if status.Duration() <= 0 {
		t.Fatalf("duration = %v, want > 0", status.Duration())
	}
}

func TestManager_RunTwiceFails(t *testing.T) {
	m := task.NewManager(nil)

	ft := &fakeTask{done: make(chan struct{})}
	defer close(ft.done)

	id, err := m.Add(context.Background(), "probe", ft)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Run(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if err := m.Run(context.Background(), id); err == nil {
		t.Fatal("expected an error running an already-started task")
	}
}

func TestManager_Cancel(t *testing.T) {
	m := task.NewManager(nil)

	ft := &fakeTask{done: make(chan struct{})}
	id, err := m.Add(context.Background(), "probe", ft)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Run(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if err := m.Cancel(context.Background(), id); err != nil {
		t.Fatal(err)
	}

	waitFor(t, func() bool {
		status, err := m.Status(id)
		return err == nil && status.State() != task.StateRunning
	})

	status, err := m.Status(id)
	if err != nil {
		t.Fatal(err)
	}
	if status.State() != task.StateCancelled {
		t.Fatalf("state = %v, want %v", status.State(), task.StateCancelled)
	}
	if !errors.Is(status.Err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", status.Err)
	}
}

func TestStatus_MarshalJSON(t *testing.T) {
	id := uuid.New()

	t.Run("not started", func(t *testing.T) {
		status := task.Status{UUID: id, Name: "probe"}
		got := marshal(t, status)

		if got["uuid"] != id.String() {
			t.Fatalf("uuid = %v, want %v", got["uuid"], id.String())
		}
		if got["name"] != "probe" {
			t.Fatalf("name = %v, want probe", got["name"])
		}
		if got["status"] != "not_started" {
			t.Fatalf("status = %v, want not_started", got["status"])
		}
		for _, key := range []string{"started", "duration", "percent", "result", "error"} {
			if _, exists := got[key]; exists {
				t.Fatalf("expected %q to be omitted, got %v", key, got[key])
			}
		}
	})

	t.Run("running with valid progress", func(t *testing.T) {
		status := task.Status{
			UUID:     id,
			Name:     "probe",
			Started:  time.Now().Add(-time.Second),
			Progress: task.Progress{Current: 1, Total: 4},
		}
		got := marshal(t, status)

		if got["status"] != "running" {
			t.Fatalf("status = %v, want running", got["status"])
		}
		if got["started"] == nil {
			t.Fatal("expected started to be present")
		}
		if got["duration"] == nil {
			t.Fatal("expected duration to be present")
		}
		if percent, ok := got["percent"].(float64); !ok || percent != 25 {
			t.Fatalf("percent = %v, want 25", got["percent"])
		}
	})

	t.Run("running with unknown total omits percent", func(t *testing.T) {
		status := task.Status{
			UUID:     id,
			Name:     "probe",
			Started:  time.Now(),
			Progress: task.Progress{Current: 1},
		}
		got := marshal(t, status)

		if _, exists := got["percent"]; exists {
			t.Fatalf("expected percent to be omitted, got %v", got["percent"])
		}
	})

	t.Run("done with result", func(t *testing.T) {
		now := time.Now()
		status := task.Status{
			UUID:     id,
			Name:     "probe",
			Started:  now.Add(-time.Second),
			Finished: now,
			Result:   "ok",
		}
		got := marshal(t, status)

		if got["status"] != "done" {
			t.Fatalf("status = %v, want done", got["status"])
		}
		if got["result"] != "ok" {
			t.Fatalf("result = %v, want ok", got["result"])
		}
		if _, exists := got["error"]; exists {
			t.Fatalf("expected error to be omitted, got %v", got["error"])
		}
	})

	t.Run("error", func(t *testing.T) {
		now := time.Now()
		status := task.Status{
			UUID:     id,
			Name:     "probe",
			Started:  now.Add(-time.Second),
			Finished: now,
			Err:      errors.New("boom"),
		}
		got := marshal(t, status)

		if got["status"] != "error" {
			t.Fatalf("status = %v, want error", got["status"])
		}
		if got["error"] != "boom" {
			t.Fatalf("error = %v, want boom", got["error"])
		}
	})

	t.Run("cancelled", func(t *testing.T) {
		now := time.Now()
		status := task.Status{
			UUID:      id,
			Name:      "probe",
			Started:   now.Add(-time.Second),
			Finished:  now,
			Cancelled: true,
			Err:       context.Canceled,
		}
		got := marshal(t, status)

		if got["status"] != "cancelled" {
			t.Fatalf("status = %v, want cancelled", got["status"])
		}
	})
}

func TestManager_NotFound(t *testing.T) {
	m := task.NewManager(nil)
	id := uuid.New()

	if _, err := m.Status(id); err == nil {
		t.Fatal("expected an error for an unknown uuid")
	}
	if err := m.Run(context.Background(), id); err == nil {
		t.Fatal("expected an error for an unknown uuid")
	}
	if err := m.Cancel(context.Background(), id); err == nil {
		t.Fatal("expected an error for an unknown uuid")
	}
}

func TestManager_List(t *testing.T) {
	m := task.NewManager(nil)

	probeFt := &fakeTask{done: make(chan struct{})}
	defer close(probeFt.done)
	probeID, err := m.Add(context.Background(), "probe", probeFt)
	if err != nil {
		t.Fatal(err)
	}

	segmentFt := &fakeTask{done: make(chan struct{})}
	defer close(segmentFt.done)
	segmentID, err := m.Add(context.Background(), "segment", segmentFt)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Run(context.Background(), segmentID); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		status, err := m.Status(segmentID)
		return err == nil && status.State() == task.StateRunning
	})

	all, err := m.List(context.Background(), task.TaskListRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("len(all) = %d, want 2", len(all))
	}
	if all[0].UUID != probeID || all[1].UUID != segmentID {
		t.Fatalf("order = [%v %v], want [%v %v]", all[0].UUID, all[1].UUID, probeID, segmentID)
	}

	byName, err := m.List(context.Background(), task.TaskListRequest{Name: strPtr("segment")})
	if err != nil {
		t.Fatal(err)
	}
	if len(byName) != 1 || byName[0].UUID != segmentID {
		t.Fatalf("byName = %+v, want just %v", byName, segmentID)
	}

	running := task.StateRunning
	byState, err := m.List(context.Background(), task.TaskListRequest{State: &running})
	if err != nil {
		t.Fatal(err)
	}
	if len(byState) != 1 || byState[0].UUID != segmentID {
		t.Fatalf("byState = %+v, want just %v", byState, segmentID)
	}
}

func TestManager_CloseCancelsRunningTasks(t *testing.T) {
	m := task.NewManager(nil)

	// Not started - Close shouldn't touch it.
	idleFt := &fakeTask{done: make(chan struct{})}
	defer close(idleFt.done)
	idleID, err := m.Add(context.Background(), "idle", idleFt)
	if err != nil {
		t.Fatal(err)
	}

	// Running and respects cancellation - Close should stop it and wait for
	// it to actually finish.
	runningFt := &fakeTask{done: make(chan struct{})}
	defer close(runningFt.done)
	runningID, err := m.Add(context.Background(), "running", runningFt)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Run(context.Background(), runningID); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		status, err := m.Status(runningID)
		return err == nil && status.State() == task.StateRunning
	})

	if err := m.Close(context.Background()); err != nil {
		t.Fatal(err)
	}

	idleStatus, err := m.Status(idleID)
	if err != nil {
		t.Fatal(err)
	}
	if idleStatus.State() != task.StateNotStarted {
		t.Fatalf("idle state = %v, want %v", idleStatus.State(), task.StateNotStarted)
	}

	runningStatus, err := m.Status(runningID)
	if err != nil {
		t.Fatal(err)
	}
	if runningStatus.State() != task.StateCancelled {
		t.Fatalf("running state = %v, want %v", runningStatus.State(), task.StateCancelled)
	}
}

func TestManager_CloseTimesOutOnStubbornTask(t *testing.T) {
	m := task.NewManager(nil)

	ft := &fakeTask{done: make(chan struct{}), ignoreCancel: true}
	defer close(ft.done)

	id, err := m.Add(context.Background(), "stubborn", ft)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Run(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		status, err := m.Status(id)
		return err == nil && status.State() == task.StateRunning
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if err := m.Close(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
}

func TestManager_Wait(t *testing.T) {
	m := task.NewManager(nil)

	ft := &fakeTask{done: make(chan struct{})}
	id, err := m.Add(context.Background(), "probe", ft)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := m.Wait(context.Background(), id); err == nil {
		t.Fatal("expected an error waiting on a task that hasn't started")
	}

	if err := m.Run(context.Background(), id); err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		close(ft.done)
	}()

	status, err := m.Wait(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if status.State() != task.StateDone {
		t.Fatalf("state = %v, want %v", status.State(), task.StateDone)
	}
	if status.Result != "fake result" {
		t.Fatalf("result = %v, want %q", status.Result, "fake result")
	}

	// Waiting again on an already-finished task should return immediately
	// with the same status.
	status, err = m.Wait(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if status.State() != task.StateDone {
		t.Fatalf("state = %v, want %v", status.State(), task.StateDone)
	}
}

func TestManager_WaitReturnsTaskError(t *testing.T) {
	m := task.NewManager(nil)

	wantErr := errors.New("boom")
	ft := &fakeTask{done: make(chan struct{}), runErr: wantErr}
	id, err := m.Add(context.Background(), "probe", ft)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Run(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	close(ft.done)

	status, err := m.Wait(context.Background(), id)
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want it to wrap %v", err, wantErr)
	}
	if status.State() != task.StateError {
		t.Fatalf("state = %v, want %v", status.State(), task.StateError)
	}
}

func TestManager_WaitTimesOut(t *testing.T) {
	m := task.NewManager(nil)

	ft := &fakeTask{done: make(chan struct{}), ignoreCancel: true}
	defer close(ft.done)

	id, err := m.Add(context.Background(), "stubborn", ft)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Run(context.Background(), id); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if _, err := m.Wait(ctx, id); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
}

////////////////////////////////////////////////////////////////////////////////
// HELPERS

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("timed out waiting for condition")
}

// marshal round-trips status through JSON and back into a generic map, so
// tests can check which keys are present/absent without depending on the
// unexported jsonStatus shape.
func marshal(t *testing.T, status task.Status) map[string]any {
	t.Helper()
	data, err := json.Marshal(status)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	return got
}

func strPtr(s string) *string {
	return &s
}
