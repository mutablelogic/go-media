package schema

import (
	"net/url"

	// Packages
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// EventName identifies what changed about a task when an Event is emitted.
type EventName string

const (
	EventAdded     EventName = "added"     // a task was registered with Manager.Add
	EventStarted   EventName = "started"   // a task's Run goroutine was launched
	EventProgress  EventName = "progress"  // a task reported progress
	EventResult    EventName = "result"    // a task set its result
	EventCancelled EventName = "cancelled" // a task's cancellation was requested
	EventFinished  EventName = "finished"  // a task's Run goroutine returned
	EventRemoved   EventName = "removed"   // a task was unregistered with Manager.Remove
)

// Event reports a change to a task tracked by a Manager - see
// Manager.Subscribe.
type Event struct {
	Name   EventName `json:"name" help:"What changed about the task." enum:"added,started,progress,result,cancelled,finished,removed" example:"started"`
	Status Status    `json:"status" help:"Snapshot of the task at the time of the event."`
}

// TaskEventRequest filters the events streamed by GET /task/event. A nil
// UUID or empty Event means "don't filter on this".
type TaskEventRequest struct {
	UUID  *string     `json:"uuid,omitempty" help:"Filter by task uuid; only events for this task are streamed." example:"123e4567-e89b-12d3-a456-426614174000"`
	Event []EventName `json:"event,omitempty" help:"Filter by event name; only these event types are streamed." enum:"added,started,progress,result,cancelled,finished,removed" example:"[\"started\",\"finished\"]"`
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

func (r TaskEventRequest) Query() url.Values {
	query := url.Values{}
	if r.UUID != nil {
		query.Set("uuid", *r.UUID)
	}
	for _, name := range r.Event {
		query.Add("event", string(name))
	}
	return query
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (e Event) String() string {
	return types.Stringify(e)
}
