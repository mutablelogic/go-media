package cmd

import (
	"fmt"

	// Packages
	httphandler "github.com/mutablelogic/go-media/profile/httphandler"
	manager "github.com/mutablelogic/go-media/profile/manager"
	pg "github.com/mutablelogic/go-pg"
	pgcmd "github.com/mutablelogic/go-pg/pkg/cmd"
	server "github.com/mutablelogic/go-server"
	servercmd "github.com/mutablelogic/go-server/pkg/cmd"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	errgroup "golang.org/x/sync/errgroup"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ServerCommands struct {
	RunServer RunServer `cmd:"" name:"run" help:"Run the profile server." group:"SERVER"`
	servercmd.OpenAPICommands
}

type RunServer struct {
	pgcmd.PostgresFlags
	servercmd.RunServer
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (runner *RunServer) Run(ctx server.Cmd) error {
	// Connect to the database, if configured
	conn, err := runner.PostgresFlags.Connect(ctx)
	if err != nil {
		return err
	} else if conn == nil {
		return fmt.Errorf("database connection is required")
	}

	// Log the server configuration
	ctx.Logger().InfoContext(ctx.Context(), "starting profile server", "name", ctx.Name(), "version", ctx.Version())

	// Create the manager, run the server, and return any error
	return runner.WithProfileManager(ctx, conn, func(manager *manager.Profile) error {
		// Create an error context - which will cancel any other goroutine on exit
		errgroup, errctx := errgroup.WithContext(ctx.Context())

		// Register http handlers for the manager
		runner.Register(func(router *httprouter.Router) error {
			ctx.Logger().DebugContext(ctx.Context(), "registering http handlers")
			return httphandler.RegisterHandlers(manager, router)
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

func (runner *RunServer) WithProfileManager(ctx server.Cmd, conn pg.PoolConn, fn func(*manager.Profile) error) error {
	// Create a manager and then call the function with the manager, returning any error
	opts := []manager.Opt{manager.WithTracer(ctx.Tracer())}
	if manager, err := manager.New(ctx.Context(), conn, opts...); err != nil {
		return err
	} else {
		return fn(manager)
	}
}
