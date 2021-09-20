package main

import (
	"context"
	"fmt"

	// Packages
	config "github.com/djthorpe/go-media/pkg/config"
)

var (
	GetHelp = Command{
		Keyword:     "help",
		Description: "Print this help",
		Fn:          GetHelpFn,
	}
	GetVersion = Command{
		Keyword:     "version",
		Description: "Print version of command",
		Fn:          GetVersionFn,
	}
)

func GetHelpFn(_ context.Context, cmd *Command, args []string) error {
	cmd.Usage()
	return nil
}

func GetVersionFn(_ context.Context, cmd *Command, args []string) error {
	fmt.Fprintln(cmd.Output(), "Version:")
	config.PrintVersion(cmd.Output())
	return nil
}
