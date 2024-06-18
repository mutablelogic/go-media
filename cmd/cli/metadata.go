package main

import (
	"os"

	// Packages

	"github.com/djthorpe/go-tablewriter"
	"github.com/mutablelogic/go-media"
)

type MetadataCmd struct {
	Path string `arg required help:"Media file" type:"path"`
}

func (cmd *MetadataCmd) Run(globals *Globals) error {
	reader, err := media.Open(cmd.Path, nil)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Print metadata
	opts := []tablewriter.TableOpt{
		tablewriter.OptHeader(),
		tablewriter.OptOutputText(),
	}
	return tablewriter.New(os.Stdout, opts...).Write(reader.Metadata())
}
