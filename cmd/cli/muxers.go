package main

import (
	"fmt"
	"os"

	// Packages
	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-media"
)

type MuxersCmd struct {
	Filter string `arg:"" optional:"" help:"Filter by mimetype, name or .ext" type:"string"`
}

func (cmd *MuxersCmd) Run(globals *Globals) error {
	manager := media.NewManager()
	formats := manager.OutputFormats(cmd.Filter)
	writer := tablewriter.New(os.Stdout, tablewriter.OptHeader(), tablewriter.OptOutputText())
	if len(formats) == 0 {
		fmt.Printf("No muxers found for %q\n", cmd.Filter)
		return nil
	} else {
		return writer.Write(formats)
	}
}
