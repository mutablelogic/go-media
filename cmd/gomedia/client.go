//go:build client && !cli

package main

import (
	profile "github.com/mutablelogic/go-media/profile/cmd"
	servercmd "github.com/mutablelogic/go-server/pkg/cmd"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLI struct {
	profile.ClientCommands
	servercmd.OpenAPICommands
}
