package main

import (
	"encoding/json"
	"fmt"
	"regexp"

	// Packages
	"github.com/mutablelogic/go-media"
)

type ProbeCmd struct {
	Path string `arg:"" required:"" help:"Media file or device name" type:"string"`
	Opts string `name:"opts" short:"o" help:"Options for opening the media file or device, (ie, \"framerate=30 video_size=176x144\")"`
}

var (
	reDevice = regexp.MustCompile(`^([a-zA-Z0-9]+):([^\/].*|)$`)
)

func (cmd *ProbeCmd) Run(globals *Globals) error {
	var format media.Format

	manager := globals.manager

	filter := media.NONE

	// Try device first
	if m := reDevice.FindStringSubmatch(cmd.Path); m != nil {
		cmd.Path = m[2]
		fmts := manager.InputFormats(filter|media.DEVICE, m[1])
		if len(fmts) > 0 {
			format = fmts[0]
		}
	}

	// Open the media file or device
	reader, err := manager.Open(cmd.Path, format, cmd.Opts)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Print out probe data
	data, _ := json.MarshalIndent(reader, "", "  ")
	fmt.Println(string(data))

	// Return success
	return nil
}
