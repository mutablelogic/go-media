package main

import (
	// Packages
	cmd "github.com/mutablelogic/go-media/gomedia/cmd"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLI struct {
	cmd.CLICommands
	cmd.ServerCommands
}
