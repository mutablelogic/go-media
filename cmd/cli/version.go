package main

import (
	"os"

	"github.com/djthorpe/go-tablewriter"
)

type VersionCmd struct{}

func (v *VersionCmd) Run(globals *Globals) error {
	manager := globals.manager

	opts := []tablewriter.TableOpt{
		tablewriter.OptOutputText(),
		tablewriter.OptDelimiter(' '),
	}
	return tablewriter.New(os.Stdout, opts...).Write(manager.Version())
}
