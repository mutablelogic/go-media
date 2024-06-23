package main

import (
	"context"
	"fmt"

	// Packages
	"github.com/djthorpe/go-tablewriter"
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

	manager := globals.manager
	if cmd.Format != "" {
		if formats := manager.InputFormats(media.NONE, cmd.Format); len(formats) == 0 {
			return fmt.Errorf("unknown format %q", cmd.Format)
		} else if len(formats) > 1 {
			return fmt.Errorf("ambiguous format %q", cmd.Format)
		} else {
			format = formats[0]
		}
	}

	// Open media file
	reader, err := manager.Open(cmd.Path, format)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Create a decoder - copy streams
	decoder, err := reader.Decoder(nil)
	if err != nil {
		return err
	}

	// Demultiplex the stream
	writer := globals.writer
	header := []tablewriter.TableOpt{tablewriter.OptHeader()}
	return decoder.Demux(context.Background(), func(packet media.Packet) error {
		if packet == nil {
			return nil
		}
		if err := writer.Write(packet, header...); err != nil {
			return err
		}
		// Reset the header
		header = header[:0]
		return nil
	})
}
