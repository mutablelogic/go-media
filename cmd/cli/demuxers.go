package main

import (
	"fmt"
	"os"

	// Packages
	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-media"
)

type DemuxersCmd struct {
	Filter string `arg:"" optional:"" help:"Filter by mimetype, name or .ext" type:"string"`
}

func (cmd *DemuxersCmd) Run(globals *Globals) error {
	manager := media.NewManager()
	formats := manager.InputFormats(cmd.Filter)
	writer := tablewriter.New(os.Stdout, tablewriter.OptHeader(), tablewriter.OptOutputText())
	if len(formats) == 0 {
		fmt.Printf("No demuxers found for %q\n", cmd.Filter)
		return nil
	} else {
		return writer.Write(formats)
	}
}
