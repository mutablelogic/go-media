package main

import (
	"fmt"

	"github.com/mutablelogic/go-media/pkg/version"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type VersionCommands struct {
	Version VersionCommand `cmd:"" help:"Report library versions" group:"OTHER"`
}

type VersionCommand struct {
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *VersionCommand) Run(globals *Globals) error {
	version := version.Map()
	for _, v := range version {
		fmt.Printf("%s: %s\n", v.Key, v.Value)
	}
	return nil
}
