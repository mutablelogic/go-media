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

func formatFromPath(manager media.Manager, filter media.MediaType, path string) (media.Format, string) {
	if m := reDevice.FindStringSubmatch(path); m != nil {

		fmts := manager.InputFormats(filter|media.DEVICE, m[1])
		if len(fmts) > 0 {
			return fmts[0], m[2]
		}
	}
	return nil, path
}

func (cmd *ProbeCmd) Run(globals *Globals) error {
	var format media.Format

	// Get format and path
	manager := globals.manager
	format, path := formatFromPath(manager, media.NONE, cmd.Path)

	// Open a reader
	reader, err := manager.Open(path, format, cmd.Opts)
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
