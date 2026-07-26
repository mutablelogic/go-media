package cmd

import (
	// Packages
	manager "github.com/mutablelogic/go-media/gomedia/manager"
	server "github.com/mutablelogic/go-server"

	// Imports
	_ "github.com/mutablelogic/go-media/metadata/application"
	_ "github.com/mutablelogic/go-media/metadata/audio"
	_ "github.com/mutablelogic/go-media/metadata/image"
	_ "github.com/mutablelogic/go-media/metadata/video"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLICommands struct {
	MetadataCLICommands
	CapabilitiesCLICommands
	EncodingCLICommands
}

type BaseCmd struct {
	ChromaprintKey string `name:"chromaprint-key" env:"CHROMAPRINT_KEY" help:"AcoustID API key for chromaprint lookups"`
}

type MetadataCLICommands struct {
	Metadata MetadataCmd `cmd:"" name:"metadata" help:"Extract metadata." group:"METADATA"`
	Artwork  ArtworkCmd  `cmd:"" name:"artwork" help:"Extract artwork." group:"METADATA"`
	Probe    ProbeCmd    `cmd:"" name:"probe" help:"Probe media file." group:"METADATA"`
	MetadataChromaprintCLICommands
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (runner *BaseCmd) IsJSONOutput(ctx server.Cmd) (bool, int) {
	width := ctx.IsTerm()
	return ctx.IsDebug() || width == 0, width
}

func (runner *BaseCmd) WithManager(ctx server.Cmd, fn func(*manager.Media) error) error {
	// Client opts
	_, clientopts, err := ctx.ClientEndpoint()
	if err != nil {
		return err
	}

	// Set basic mamager options
	opts := []manager.Opt{
		manager.WithTracer(ctx.Tracer()),
	}

	// Chromaprint key
	if runner.ChromaprintKey != "" {
		opts = append(opts, manager.WithAcoustIDKey(runner.ChromaprintKey, clientopts...))
	}

	// Create a manager and then call the function with the manager, returning any error
	if manager, err := manager.New(ctx.Context(), opts...); err != nil {
		return err
	} else {
		return fn(manager)
	}
}
