package main

import (
	"encoding/json"
	"fmt"

	// Packages
	version "github.com/mutablelogic/go-media/pkg/version"
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type VersionCommands struct {
	Version VersionCommand `cmd:"" group:"MISC" help:"Print version information"`
}

type VersionCommand struct{}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *VersionCommand) Run(app server.Cmd) error {
	// Print version information
	data, err := json.MarshalIndent(version.Map(), "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
