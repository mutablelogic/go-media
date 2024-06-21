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

type DevicesCmd struct {
	Filter string `arg:"" optional:"" help:"Filter by mimetype, name or .ext" type:"string"`
}

func (cmd *MuxersCmd) Run(globals *Globals) error {
	manager := media.NewManager()
	return Run(cmd.Filter, Outputs(manager, media.ANY, cmd.Filter))
}

func (cmd *DemuxersCmd) Run(globals *Globals) error {
	manager := media.NewManager()
	return Run(cmd.Filter, Inputs(manager, media.ANY, cmd.Filter))
}

func (cmd *DevicesCmd) Run(globals *Globals) error {
	manager := media.NewManager()
	var formats []media.Format
	formats = append(formats, Inputs(manager, media.DEVICE, cmd.Filter)...)
	formats = append(formats, Outputs(manager, media.DEVICE, cmd.Filter)...)

	if len(formats) == 0 {
		fmt.Printf("No devices found for %q\n", cmd.Filter)
		return nil
	}

	var result []media.Device
	for _, format := range formats {
		result = append(result, manager.Devices(format)...)
	}

	writer := tablewriter.New(os.Stdout, tablewriter.OptHeader(), tablewriter.OptOutputText())
	return writer.Write(result)
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

func Inputs(manager media.Manager, mediatype media.MediaType, filter string) []media.Format {
	if filter == "" {
		return manager.InputFormats(mediatype)
	} else {
		return manager.InputFormats(mediatype, filter)
	}
}

func Outputs(manager media.Manager, mediatype media.MediaType, filter string) []media.Format {
	if filter == "" {
		return manager.OutputFormats(mediatype)
	} else {
		return manager.OutputFormats(mediatype, filter)
	}
}
