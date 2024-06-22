package main

import (
	"os"

	// Packages
	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-media"
)

type CodecsCmd struct{}

type SampleFormatsCmd struct{}

type ChannelLayoutsCmd struct{}

type PixelFormatsCmd struct{}

func (cmd *CodecsCmd) Run(globals *Globals) error {
	manager := media.NewManager()
	writer := tablewriter.New(os.Stdout, tablewriter.OptHeader(), tablewriter.OptOutputText())

	codecs := manager.Codecs()

	return writer.Write(codecs)
}

func (cmd *SampleFormatsCmd) Run(globals *Globals) error {
	manager := media.NewManager()
	writer := tablewriter.New(os.Stdout, tablewriter.OptHeader(), tablewriter.OptOutputText())
	return writer.Write(manager.SampleFormats())
}

func (cmd *ChannelLayoutsCmd) Run(globals *Globals) error {
	manager := media.NewManager()
	writer := tablewriter.New(os.Stdout, tablewriter.OptHeader(), tablewriter.OptOutputText())
	return writer.Write(manager.ChannelLayouts())
}

func (cmd *PixelFormatsCmd) Run(globals *Globals) error {
	manager := media.NewManager()
	writer := tablewriter.New(os.Stdout, tablewriter.OptHeader(), tablewriter.OptOutputText())
	return writer.Write(manager.PixelFormats())
}
