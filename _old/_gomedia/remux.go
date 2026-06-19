package main

import (
	"fmt"
	"os"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
)

// /////////////////////////////////////////////////////////////////////////////
// TYPES
type RemuxCommands struct {
	Remux RemuxCommand `cmd:"" help:"Remux media file or stream" group:"PROCESS"`
}

type RemuxCommand struct {
	schema.RemuxRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *RemuxCommand) Run(globals *Globals) error {
	// If the path is "-", use stdin
	if cmd.Input == "-" {
		cmd.Reader = os.Stdin
		cmd.Input = ""
	}

	// Call manager method
	response, err := globals.manager.Remux(globals.ctx, os.Stdout, &cmd.RemuxRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}
