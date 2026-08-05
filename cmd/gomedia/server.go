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
	task "github.com/mutablelogic/go-media/task/cmd"
	taskhttphandler "github.com/mutablelogic/go-media/task/httphandler"
	taskmanager "github.com/mutablelogic/go-media/task/manager"
	tmdbcmd "github.com/mutablelogic/go-media/tmdb/cmd"
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
	Task      task.ClientCommands    `embed:""`
	RunServer RunServer              `cmd:"" name:"run" help:"Run the gomedia server." group:"SERVER"`
	servercmd.OpenAPICommands
	TMDB tmdbcmd.ClientCommands `embed:""`
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

	// Create the managers, run the server, and return any error
	return runner.WithProfileManager(ctx, conn, func(profiles *profilemanager.Profile) error {
		return runner.WithTaskManager(ctx, func(tasks *taskmanager.Manager) error {
			return runner.WithMediaManager(ctx, profiles, tasks, func(media *mediamanager.Media) error {
				// Create an error context - which will cancel any other goroutine on exit
				errgroup, errctx := errgroup.WithContext(ctx.Context())

				// Register http handlers for the manager
				runner.Register(func(router *httprouter.Router) error {
					ctx.Logger().DebugContext(ctx.Context(), "registering http handlers")
					return errors.Join(
						profilehttphandler.RegisterCapabilityHandlers(profiles, router),
						profilehttphandler.RegisterAudioProfileHandlers(profiles, router),
						taskhttphandler.RegisterTaskHandlers(tasks, router),
						mediahttphandler.RegisterMetadataHandlers(media, router),
						mediahttphandler.RegisterEncoderHandlers(media, tasks, router),
					)
				})

				// Run the profile manager
				//errgroup.Go(func() error {
				//	return profiles.Run(errctx, ctx.Logger())
				//})

				// Run the media manager
				errgroup.Go(func() error {
					return media.Run(errctx, ctx.Logger())
				})

				// Run the task manager
				errgroup.Go(func() error {
					return tasks.Run(errctx, ctx.Logger())
				})

				// Run the server - if any co-routine in the error group returns an error, the server will be shutdown
				errgroup.Go(func() error {
					return runner.RunServer.Run(ctx.WithContext(errctx))
				})

				// Wait for the server and manager to exit, and return any error
				return errgroup.Wait()
			})
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

func (runner *RunServer) WithTaskManager(ctx server.Cmd, fn func(*taskmanager.Manager) error) error {
	// Create a manager and then call the function with the manager, returning any error
	opts := []taskmanager.Opt{taskmanager.WithTracer(ctx.Tracer())}
	if manager, err := taskmanager.New(ctx.Context(), opts...); err != nil {
		return err
	} else {
		return fn(manager)
	}
}

func (runner *RunServer) WithMediaManager(ctx server.Cmd, profiles *profilemanager.Profile, tasks *taskmanager.Manager, fn func(*mediamanager.Media) error) error {
	// Create a manager and then call the function with the manager, returning any error
	opts := []mediamanager.Opt{
		mediamanager.WithTracer(ctx.Tracer()),
		mediamanager.WithProfileManager(profiles),
		mediamanager.WithTaskManager(tasks),
	}
	if media, err := mediamanager.New(ctx.Context(), opts...); err != nil {
		return err
	} else {
		return fn(media)
	}
}
