package cmd

import (
	"context"
	"errors"
	"fmt"

	// Packages
	uuid "github.com/google/uuid"
	httpclient "github.com/mutablelogic/go-media/task/httpclient"
	schema "github.com/mutablelogic/go-media/task/schema"
	server "github.com/mutablelogic/go-server"
	types "github.com/mutablelogic/go-server/pkg/types"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ListTasksCmd struct {
	schema.TaskListRequest
}

type GetTaskCmd struct {
	UUID string `arg:"" name:"uuid" help:"UUID of the task."`
}

type CancelTaskCmd struct {
	UUID string `arg:"" name:"uuid" help:"UUID of the task."`
}

type WatchTasksCmd struct {
	schema.TaskEventRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ListTasksCmd) Run(ctx server.Cmd) error {
	return withClient(ctx, "ListTasks", func(ctx context.Context, client *httpclient.Client) error {
		// List the tasks
		tasks, err := client.ListTasks(ctx, cmd.TaskListRequest)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(tasks))

		// Return success
		return nil
	})
}

func (cmd *GetTaskCmd) Run(ctx server.Cmd) error {
	id, err := uuid.Parse(cmd.UUID)
	if err != nil {
		return err
	}
	return withClient(ctx, "GetTask", func(ctx context.Context, client *httpclient.Client) error {
		// Get the task
		task, err := client.GetTask(ctx, id)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(task))

		// Return success
		return nil
	})
}

func (cmd *CancelTaskCmd) Run(ctx server.Cmd) error {
	id, err := uuid.Parse(cmd.UUID)
	if err != nil {
		return err
	}
	return withClient(ctx, "CancelTask", func(ctx context.Context, client *httpclient.Client) error {
		// Cancel the task
		task, err := client.CancelTask(ctx, id)
		if err != nil {
			return err
		}

		fmt.Println(types.Stringify(task))

		// Return success
		return nil
	})
}

func (cmd *WatchTasksCmd) Run(ctx server.Cmd) error {
	return withClient(ctx, "WatchTasks", func(ctx context.Context, client *httpclient.Client) error {
		// Stream events, printing each as it arrives, until the context is
		// cancelled (Ctrl+C) or the server stops.
		err := client.SubscribeEvents(ctx, cmd.TaskEventRequest, func(e *schema.Event) error {
			fmt.Println(types.Stringify(e))
			return nil
		})
		if errors.Is(err, context.Canceled) {
			return nil
		}

		// Return any other error
		return err
	})
}
