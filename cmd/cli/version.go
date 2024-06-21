package main

import (
	"os"

	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-media"
)

type VersionCmd struct{}

func (v *VersionCmd) Run(globals *Globals) error {
	opts := []tablewriter.TableOpt{
		tablewriter.OptOutputText(),
		tablewriter.OptDelimiter(' '),
	}
	return tablewriter.New(os.Stdout, opts...).Write(media.NewManager().Version())
}
