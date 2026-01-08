package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"

	// Packages
	kong "github.com/alecthomas/kong"
	"github.com/mutablelogic/go-client"
	"github.com/mutablelogic/go-media/pkg/ffmpeg/httpclient"
	task "github.com/mutablelogic/go-media/pkg/ffmpeg/task"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Globals struct {
	Debug          bool             `name:"debug" help:"Enable debug logging"`
	Version        kong.VersionFlag `name:"version" help:"Print version and exit"`
	ChromaprintKey string           `name:"chromaprint-key" env:"CHROMAPRINT_KEY" help:"AcoustID API key for chromaprint lookups"`
	Endpoint       string           `name:"url" env:"GOMEDIA_ENDPOINT" help:"Server endpoint URL"`

	// Private fields
	ctx     context.Context
	cancel  context.CancelFunc
	manager *task.Manager
	client  *httpclient.Client
}

type CLI struct {
	Globals
	ListCommands
	FingerprintCommands
	ProbeCommands
	DecodeCommands
	RemuxCommands
	ServerCommands
	PlayCommands
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func main() {
	cli := new(CLI)
	ctx := kong.Parse(cli,
		kong.Name("gomedia"),
		kong.Description("go-media command line interface"),
		kong.Vars{
			"version": VersionJSON(),
		},
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	// Create the context and cancel function
	cli.Globals.ctx, cli.Globals.cancel = context.WithCancel(context.Background())
	defer cli.Globals.cancel()

	// Set up signal handling to force exit on CTRL+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nInterrupted")
		cli.Globals.cancel()
		// TODO: We should probably give some time for cleanup here
		os.Exit(130) // Standard exit code for CTRL+C
	}()

	// Set options
	opts := []task.Opt{}
	opts = append(opts, task.WithTraceFn(func(msg string) {
		fmt.Fprintln(os.Stderr, msg)
	}, cli.Globals.Debug))
	if cli.Globals.ChromaprintKey != "" {
		opts = append(opts, task.WithChromaprintKey(cli.Globals.ChromaprintKey))
	}

	// Create manager
	manager, err := task.NewManager(opts...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else {
		cli.Globals.manager = manager
	}

	// Call the Run() method of the selected parsed command.
	if err := ctx.Run(&cli.Globals); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (g *Globals) Client() (*httpclient.Client, error) {
	if g.Endpoint == "" {
		return nil, nil
	}

	url, err := url.Parse(g.Endpoint)
	if err != nil {
		return nil, err
	}

	// Client options
	opts := []client.ClientOpt{}
	if g.Debug {
		opts = append(opts, client.OptTrace(os.Stderr, true))
	}

	// Create a client with the calculated endpoint
	return httpclient.New(url.String(), opts...)
}
