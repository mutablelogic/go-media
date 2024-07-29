package main

import (
	// Packages
	"github.com/djthorpe/go-tablewriter"
)

type CodecsCmd struct{}

type SampleFormatsCmd struct{}

type ChannelLayoutsCmd struct{}

type PixelFormatsCmd struct{}

func (cmd *CodecsCmd) Run(globals *Globals) error {
	manager := globals.manager
	writer := globals.writer
	codecs := manager.Codecs()
	return writer.Write(codecs, tablewriter.OptHeader())
}

func (cmd *SampleFormatsCmd) Run(globals *Globals) error {
	manager := globals.manager
	writer := globals.writer
	formats := manager.SampleFormats()
	return writer.Write(formats, tablewriter.OptHeader())
}

func (cmd *ChannelLayoutsCmd) Run(globals *Globals) error {
	manager := globals.manager
	writer := globals.writer
	layouts := manager.ChannelLayouts()
	return writer.Write(layouts, tablewriter.OptHeader())
}

func (cmd *PixelFormatsCmd) Run(globals *Globals) error {
	manager := globals.manager
	writer := globals.writer
	formats := manager.PixelFormats()
	return writer.Write(formats, tablewriter.OptHeader())
}
