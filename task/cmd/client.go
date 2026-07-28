package cmd

import (
	"context"

	// Packages
	otel "github.com/mutablelogic/go-client/pkg/otel"
	httpclient "github.com/mutablelogic/go-media/task/httpclient"
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ClientCommands struct {
	ListTasks  ListTasksCmd  `cmd:"" name:"tasks" help:"List tasks." group:"TASK"`
	GetTask    GetTaskCmd    `cmd:"" name:"task" help:"Get the details of a task." group:"TASK"`
	CancelTask CancelTaskCmd `cmd:"" name:"task-cancel" help:"Cancel a task." group:"TASK"`
	WatchTasks WatchTasksCmd `cmd:"" name:"task-events" help:"Stream task events until interrupted (Ctrl+C)." group:"TASK"`
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func withClient(cmd server.Cmd, name string, fn func(context.Context, *httpclient.Client) error) (err error) {
	endpoint, opts, err := cmd.ClientEndpoint()
	if err != nil {
		return err
	}
	client, err := httpclient.New(endpoint, opts...)
	if err != nil {
		return err
	}
	ctx, endSpan := otel.StartSpan(cmd.Tracer(), cmd.Context(), name)
	defer func() { endSpan(err) }()
	return fn(ctx, client)
}
