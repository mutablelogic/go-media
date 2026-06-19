package main

import (
	"fmt"
	"os"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ProbeCommands struct {
	Probe ProbeCommand `cmd:"" help:"Probe media file or stream" group:"PROCESS"`
}

type ProbeCommand struct {
	schema.ProbeRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ProbeCommand) Run(globals *Globals) error {
	// If the path is "-", use stdin
	if cmd.Input == "-" {
		cmd.Reader = os.Stdin
		cmd.Input = ""
	}

	// Call manager method
	response, err := globals.manager.Probe(globals.ctx, &cmd.ProbeRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}
