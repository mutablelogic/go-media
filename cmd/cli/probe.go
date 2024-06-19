package main

import (
	"encoding/json"
	"fmt"
	"regexp"

	// Packages
	"github.com/mutablelogic/go-media"
)

type ProbeCmd struct {
	Path  string `arg:"" required:"" help:"Media file or device name" type:"string"`
	Audio bool   `name:"audio" short:"a" help:"Probe audio stream" type:"bool"`
	Video bool   `name:"video" short:"v" help:"Probe video stream" type:"bool"`
}

var (
	reDevice = regexp.MustCompile(`^([a-zA-Z0-9]+):(.*)$`)
)

func (cmd *ProbeCmd) Run(globals *Globals) error {
	var format media.Format

	manager := media.NewManager()
	filter := media.NONE
	if cmd.Audio {
		filter |= media.AUDIO
	}
	if cmd.Video {
		filter |= media.VIDEO
	}

	// Try device first
	if m := reDevice.FindStringSubmatch(cmd.Path); m != nil {
		cmd.Path = m[2]
		fmts := manager.InputFormats(filter|media.DEVICE, m[1])
		if len(fmts) == 1 {
			format = fmts[0]
		} else if len(fmts) > 1 {
			return fmt.Errorf("ambigious device name %q, use -audio or -video", m[1])
		}
	}

	// Open the media file or device
	reader, err := manager.Open(cmd.Path, format)
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
