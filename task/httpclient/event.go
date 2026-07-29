package httpclient

import (
	"context"

	// Packages
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-media/task/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// TaskEventFunc is called for every task event received from SubscribeEvents.
// Return io.EOF to stop the subscription cleanly, or any other error to stop
// it and have SubscribeEvents return that error.
type TaskEventFunc func(*schema.Event) error

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// SubscribeEvents streams task events matching req - see
// schema.TaskEventRequest for how uuid and event filter the stream. It
// blocks, calling fn for every event, until ctx is done, the server stops,
// or fn itself stops it (io.EOF to stop cleanly, any other error to stop
// and have SubscribeEvents return that error).
func (c *Client) SubscribeEvents(ctx context.Context, req schema.TaskEventRequest, fn TaskEventFunc) error {
	return c.DoWithContext(ctx, nil, nil,
		client.OptPath("task", "event"),
		client.OptQuery(req.Query()),
		client.OptTextStreamCallback(func(evt client.TextStreamEvent) error {
			// Keep-alive pings carry no data - nothing to decode or deliver.
			if evt.Data == "" {
				return nil
			}
			var status schema.Status
			if err := evt.Json(&status); err != nil {
				return err
			}
			return fn(&schema.Event{Name: schema.EventName(evt.Event), Status: status})
		}),
	)
}
