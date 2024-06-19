package main

import (
	"encoding/json"
	"fmt"

	// Packages
	"github.com/mutablelogic/go-media"
)

type DecodeCmd struct {
	Path   string `arg:"" required:"" help:"Media file" type:"path"`
	Format string `name:"format" short:"f" help:"Format of input file (name, .extension or mimetype)" type:"string"`
	Audio  *bool  `name:"audio" short:"a" help:"Output raw audio stream" type:"bool"`
	Video  *bool  `name:"video" short:"v" help:"Output raw video stream" type:"bool"`
}

func (cmd *DecodeCmd) Run(globals *Globals) error {
	var format media.Format

	manager := media.NewManager()
	if cmd.Format != "" {
		if formats := manager.InputFormats(media.NONE, cmd.Format); len(formats) == 0 {
			return fmt.Errorf("unknown format %q", cmd.Format)
		} else if len(formats) > 1 {
			return fmt.Errorf("ambiguous format %q", cmd.Format)
		} else {
			format = formats[0]
		}
	}

	reader, err := manager.Open(cmd.Path, format)
	if err != nil {
		return err
	}
	defer reader.Close()

	data, _ := json.MarshalIndent(reader, "", "  ")
	fmt.Println(string(data))

	return nil
}
