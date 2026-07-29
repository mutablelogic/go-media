package httpclient

import (
	"context"

	// Packages
	uuid "github.com/google/uuid"
	client "github.com/mutablelogic/go-client"
	schema "github.com/mutablelogic/go-media/task/schema"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListTasks returns the tasks matching req, in the order they were added.
func (c *Client) ListTasks(ctx context.Context, req schema.TaskListRequest) (*schema.TaskList, error) {
	// Perform request
	var response schema.TaskList
	if err := c.DoWithContext(ctx, nil, &response, client.OptPath("task"), client.OptQuery(req.Query())); err != nil {
		return nil, err
	}

	// Return the response
	return types.Ptr(response), nil
}

// GetTask returns the current status of the task with the given uuid.
func (c *Client) GetTask(ctx context.Context, id uuid.UUID) (*schema.Status, error) {
	// Perform request
	var response schema.Status
	if err := c.DoWithContext(ctx, nil, &response, client.OptPath("task", id)); err != nil {
		return nil, err
	}

	// Return the response
	return types.Ptr(response), nil
}

// CancelTask cancels the task with the given uuid. It's a no-op if the task
// has already finished or was never started. Returns the task's status
// after the cancellation request.
func (c *Client) CancelTask(ctx context.Context, id uuid.UUID) (*schema.Status, error) {
	// Perform request
	var response schema.Status
	if err := c.DoWithContext(ctx, client.MethodDelete, &response, client.OptPath("task", id)); err != nil {
		return nil, err
	}

	// Return the response
	return types.Ptr(response), nil
}
