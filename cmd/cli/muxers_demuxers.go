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

type MuxersCmd struct {
	Filter string `arg:"" optional:"" help:"Filter by mimetype, name or .ext" type:"string"`
}

func (cmd *MuxersCmd) Run(globals *Globals) error {
	manager := media.NewManager()
	var formats []media.Format
	if cmd.Filter == "" {
		formats = manager.OutputFormats(media.ANY)
	} else {
		formats = manager.OutputFormats(media.ANY, cmd.Filter)
	}
	return Run(cmd.Filter, formats)
}

func (cmd *DemuxersCmd) Run(globals *Globals) error {
	manager := media.NewManager()
	var formats []media.Format
	if cmd.Filter == "" {
		formats = manager.InputFormats(media.ANY)
	} else {
		formats = manager.InputFormats(media.ANY, cmd.Filter)
	}
	return Run(cmd.Filter, formats)
}

func Run(filter string, formats []media.Format) error {
	writer := tablewriter.New(os.Stdout, tablewriter.OptHeader(), tablewriter.OptOutputText())
	if len(formats) == 0 {
		fmt.Printf("No (de)muxers found for %q\n", filter)
		return nil
	} else {
		return writer.Write(formats)
	}
}
