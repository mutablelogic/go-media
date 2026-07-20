package main

import (
	// Packages
	profile "github.com/mutablelogic/go-media/profile/cmd"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLI struct {
	profile.ServerCommands
	profile.ClientCommands
}
