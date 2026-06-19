package cmd

import (
	// Packages
	manager "github.com/mutablelogic/go-media/media/manager"
	server "github.com/mutablelogic/go-server"
	servercmd "github.com/mutablelogic/go-server/pkg/cmd"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	errgroup "golang.org/x/sync/errgroup"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ServerCommands struct {
	RunServer RunServer `cmd:"" name:"run" help:"Run the media server." group:"SERVER"`
	servercmd.OpenAPICommands
}

type RunServer struct {
	servercmd.RunServer
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (runner *RunServer) Run(ctx server.Cmd) error {
	// Log the server configuration
	ctx.Logger().InfoContext(ctx.Context(), "starting media server", "name", ctx.Name(), "version", ctx.Version())

	// Create the manager, run the server, and return any error
	return runner.WithManager(ctx, func(manager *manager.Manager) error {
		// Create an error context - which will cancel any other goroutine on exit
		errgroup, errctx := errgroup.WithContext(ctx.Context())

		// Register http handlers for the manager
		runner.Register(func(router *httprouter.Router) error {
			ctx.Logger().DebugContext(ctx.Context(), "registering http handlers")
			return nil
		})

		// Run the manager
		errgroup.Go(func() error {
			return manager.Run(errctx, ctx.Logger())
		})

		// Run the server - if any co-routine in the error group returns an error, the server will be shutdown
		errgroup.Go(func() error {
			return runner.RunServer.Run(ctx.WithContext(errctx))
		})

		// Wait for the server and manager to exit, and return any error
		return errgroup.Wait()
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (runner *RunServer) WithManager(ctx server.Cmd, fn func(*manager.Manager) error) error {
	opts := []manager.Opt{
		manager.WithMeter(ctx.Meter()),
		manager.WithTracer(ctx.Tracer()),
	}
	if manager, err := manager.New(ctx.Context(), opts...); err != nil {
		return err
	} else {
		return fn(manager)
	}
}
