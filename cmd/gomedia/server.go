//go:build !cli && !client

package main

import (
	"errors"
	"fmt"

	// Packages
	gomedia "github.com/mutablelogic/go-media/gomedia/cmd"
	mediahttphandler "github.com/mutablelogic/go-media/gomedia/httphandler"
	mediamanager "github.com/mutablelogic/go-media/gomedia/manager"
	profile "github.com/mutablelogic/go-media/profile/cmd"
	profilehttphandler "github.com/mutablelogic/go-media/profile/httphandler"
	profilemanager "github.com/mutablelogic/go-media/profile/manager"
	pg "github.com/mutablelogic/go-pg"
	pgcmd "github.com/mutablelogic/go-pg/pkg/cmd"
	server "github.com/mutablelogic/go-server"
	servercmd "github.com/mutablelogic/go-server/pkg/cmd"
	httprouter "github.com/mutablelogic/go-server/pkg/httprouter"
	errgroup "golang.org/x/sync/errgroup"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLI struct {
	Media     gomedia.ClientCommands `embed:""`
	Profile   profile.ClientCommands `embed:""`
	RunServer RunServer              `cmd:"" name:"run" help:"Run the gomedia server." group:"SERVER"`
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
	return runner.WithProfileManager(ctx, conn, func(manager *profilemanager.Profile) error {
		return runner.WithMediaManager(ctx, func(media *mediamanager.Media) error {
			// Create an error context - which will cancel any other goroutine on exit
			errgroup, errctx := errgroup.WithContext(ctx.Context())

			// Register http handlers for the manager
			runner.Register(func(router *httprouter.Router) error {
				ctx.Logger().DebugContext(ctx.Context(), "registering http handlers")
				return errors.Join(
					profilehttphandler.RegisterCapabilityHandlers(manager, router),
					profilehttphandler.RegisterAudioProfileHandlers(manager, router),
					mediahttphandler.RegisterMetadataHandlers(media, router),
				)
			})

			// Run the profile manager
			//errgroup.Go(func() error {
			//	return profilemanager.Run(errctx, ctx.Logger())
			//})

			// Run the media manager
			errgroup.Go(func() error {
				return media.Run(errctx, ctx.Logger())
			})

			// Run the server - if any co-routine in the error group returns an error, the server will be shutdown
			errgroup.Go(func() error {
				return runner.RunServer.Run(ctx.WithContext(errctx))
			})

			// Wait for the server and manager to exit, and return any error
			return errgroup.Wait()
		})
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (runner *RunServer) WithProfileManager(ctx server.Cmd, conn pg.PoolConn, fn func(*profilemanager.Profile) error) error {
	// Create a manager and then call the function with the manager, returning any error
	opts := []profilemanager.Opt{profilemanager.WithTracer(ctx.Tracer())}
	if manager, err := profilemanager.New(ctx.Context(), conn, opts...); err != nil {
		return err
	} else {
		return fn(manager)
	}
}

func (runner *RunServer) WithMediaManager(ctx server.Cmd, fn func(*mediamanager.Media) error) error {
	// Create a manager and then call the function with the manager, returning any error
	opts := []mediamanager.Opt{mediamanager.WithTracer(ctx.Tracer())}
	if media, err := mediamanager.New(ctx.Context(), opts...); err != nil {
		return err
	} else {
		return fn(media)
	}
}
