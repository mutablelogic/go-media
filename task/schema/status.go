package schema

import (
	"encoding/json"
	"errors"
	"time"

	// Packages
	uuid "github.com/google/uuid"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Status is a snapshot of a task tracked by a Manager.
type Status struct {
	UUID uuid.UUID `json:"uuid" help:"Unique identifier for the task." example:"123e4567-e89b-12d3-a456-426614174000"`
	Name string    `json:"name" help:"Caller-chosen name for the task" example:"transcode"`
	Task string    `json:"task" help:"The task class." example:"metadata"`
	Progress
	Result    any       `json:"result,omitempty" help:"Task-defined result value, set once the task has finished successfully."`
	Started   time.Time `json:"started,omitzero" help:"When the task started running; zero until the task has been run." example:"2026-07-27T09:00:00Z"`
	Finished  time.Time `json:"finished,omitzero" help:"When the task finished; zero until the task has finished." example:"2026-07-27T09:05:00Z"`
	Cancelled bool      `json:"cancelled,omitempty" help:"Whether the task was cancelled, regardless of the error it returned." example:"false"`
	Err       error     `json:"error,omitempty" help:"Error the task returned, if any." example:"decode failed: unexpected EOF"`
}

////////////////////////////////////////////////////////////////////////////////
// MARSHALING

// MarshalJSON renders Err as its message string
func (s Status) MarshalJSON() ([]byte, error) {
	type jsonStatus Status
	out := struct {
		jsonStatus
		Err string `json:"error,omitempty"`
	}{
		jsonStatus: jsonStatus(s),
	}
	if s.Err != nil {
		out.Err = s.Err.Error()
	}
	return json.Marshal(out)
}

// UnmarshalJSON is the inverse of MarshalJSON: it reads Err back as a plain
// error carrying that message string
func (s *Status) UnmarshalJSON(data []byte) error {
	type jsonStatus Status
	aux := struct {
		*jsonStatus
		Err string `json:"error,omitempty"`
	}{
		jsonStatus: (*jsonStatus)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Err != "" {
		s.Err = errors.New(aux.Err)
	} else {
		s.Err = nil
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s Status) String() string {
	return types.Stringify(s)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

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
