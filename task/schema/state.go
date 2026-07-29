package schema

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
