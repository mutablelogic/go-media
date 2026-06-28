package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	// Packages
	exif "github.com/mutablelogic/go-media/pkg/exif"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ExifCommands struct {
	ExifProbe ExifProbeCommand `cmd:"" help:"Parse EXIF data" group:"PROCESS"`
}

type ExifProbeCommand struct {
	Input  string    `arg:"" optional:"" help:"Input file path (use - for stdin)"`
	Reader io.Reader `kong:"-"`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ExifProbeCommand) Run(globals *Globals) error {
	var (
		e   *exif.EXIF
		err error
	)

	if cmd.Input == "-" || (cmd.Input == "" && cmd.Reader != nil) {
		r := cmd.Reader
		if r == nil {
			r = os.Stdin
		}
		e, err = exif.Read(r)
	} else {
		e, err = exif.Open(cmd.Input)
	}
	if err != nil {
		return err
	}
	defer e.Close()

	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
