package main

import (
	"fmt"

	// Packages
	segmenter "github.com/mutablelogic/go-media/segmenter"
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type SegmentCmd struct {
	SegmentOpts
	Path string `arg:"" name:"path" help:"Path or URL of the audio to segment."`
	OutputOpts
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *SegmentCmd) Run(ctx server.Cmd) error {
	opts, err := cmd.options()
	if err != nil {
		return err
	}
	out, err := cmd.open(cmd.Rate)
	if err != nil {
		return err
	}
	defer out.Close()

	seg, err := segmenter.Open(cmd.Path, cmd.Rate, opts...)
	if err != nil {
		return err
	}
	defer seg.Close()

	// Print each segment as soon as it's output
	fmt.Print(out.header())
	return seg.Decode(ctx.Context(), func(s *segmenter.Segment) error {
		row, err := out.write(s)
		if err != nil {
			return err
		}
		fmt.Print(row)
		return nil
	})
}
