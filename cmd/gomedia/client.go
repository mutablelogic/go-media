//go:build client && !cli

package main

import (
	gomedia "github.com/mutablelogic/go-media/gomedia/cmd"
	profile "github.com/mutablelogic/go-media/profile/cmd"
	servercmd "github.com/mutablelogic/go-server/pkg/cmd"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLI struct {
	gomedia.ClientCommands
	profile.ClientCommands
	servercmd.OpenAPICommands
}
