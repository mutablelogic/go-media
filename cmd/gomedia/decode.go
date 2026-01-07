package main

import (
	"os"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type DecodeCommands struct {
	Decode DecodeCommand `cmd:"" help:"Decode media file or stream to JSON" group:"PROCESS"`
}

type DecodeCommand struct {
	schema.DecodeRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *DecodeCommand) Run(globals *Globals) error {
	// If the path is "-", use stdin
	if cmd.Input == "-" {
		cmd.Reader = os.Stdin
		cmd.Input = ""
	}

	// Call manager method - outputs JSON to stdout
	err := globals.manager.Decode(globals.ctx, os.Stdout, &cmd.DecodeRequest)
	if err != nil {
		return err
	}

	return nil
}
