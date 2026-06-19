//go:build !client

package main

import (
	// Packages
	mediacmd "github.com/mutablelogic/go-media/media/cmd"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CLI struct {
	mediacmd.ServerCommands
}
