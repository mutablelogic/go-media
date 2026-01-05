package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	server "github.com/mutablelogic/go-server"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type CodecCommands struct {
	Formats        ListFormats        `cmd:"" group:"MISC" help:"List transport formats"`
	Codecs         ListCodecs         `cmd:"" group:"MISC" help:"List codecs"`
	SampleFormats  ListSampleFormats  `cmd:"" group:"MISC" help:"List sample formats"`
	PixelFormats   ListPixelFormats   `cmd:"" group:"MISC" help:"List pixel formats"`
	ChannelLayouts ListChannelLayouts `cmd:"" group:"MISC" help:"List channel layouts"`
}

type ListFormats struct {
	Name []string `arg:"" help:"Filter by name" optional:""`
	Type string   `cmd:"" help:"Type of codecs to list" enum:"any,audio,video,subtitle,input,output,device" default:"any"`
}

type ListCodecs struct {
	Name []string `arg:"" help:"Filter by name" optional:""`
	Type string   `cmd:"" help:"Type of codecs to list" enum:"any,audio,video,subtitle" default:"any"`
}

type ListSampleFormats struct {
	Name []string `arg:"" help:"Filter by name" optional:""`
}

type ListPixelFormats struct {
	Name []string `arg:"" help:"Filter by name" optional:""`
}

type ListChannelLayouts struct {
	Name []string `arg:"" help:"Filter by name" optional:""`
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ListFormats) Run(app server.Cmd) error {
	manager, err := ffmpeg.NewManager()
	if err != nil {
		return err
	}

	// Filter by codec type
	t := media.ANY
	switch cmd.Type {
	case "audio":
		t = media.AUDIO
	case "video":
		t = media.VIDEO
	case "subtitle":
		t = media.SUBTITLE
	case "input":
		t = media.INPUT
	case "output":
		t = media.OUTPUT
	case "device":
		t = media.DEVICE
	}

	// Write codecs to standard output
	return write(os.Stdout, formatToMeta(manager.Formats(t, cmd.Name...)), nil)
}

func formatToMeta(formats []media.Format) []media.Metadata {
	result := make([]media.Metadata, len(formats))
	for i, format := range formats {
		result[i] = ffmpeg.NewMetadata(format.Name(), format)
	}
	return result
}

func (cmd *ListCodecs) Run(app server.Cmd) error {
	manager, err := ffmpeg.NewManager()
	if err != nil {
		return err
	}

	// Filter by codec type
	t := media.ANY
	switch cmd.Type {
	case "audio":
		t = media.AUDIO
	case "video":
		t = media.VIDEO
	case "subtitle":
		t = media.SUBTITLE
	}

	// Write codecs to standard output
	return write(os.Stdout, manager.Codecs(t, cmd.Name...), nil)
}

func (cmd *ListSampleFormats) Run(app server.Cmd) error {
	manager, err := ffmpeg.NewManager()
	if err != nil {
		return err
	}

	// Write sample formats to standard output
	return write(os.Stdout, manager.SampleFormats(), func(m media.Metadata) bool {
		return len(cmd.Name) == 0 || slices.Contains(cmd.Name, m.Key())
	})
}

func (cmd *ListPixelFormats) Run(app server.Cmd) error {
	manager, err := ffmpeg.NewManager()
	if err != nil {
		return err
	}

	// Write pixel formats to standard output
	return write(os.Stdout, manager.PixelFormats(), func(m media.Metadata) bool {
		return len(cmd.Name) == 0 || slices.Contains(cmd.Name, m.Key())
	})
}

func (cmd *ListChannelLayouts) Run(app server.Cmd) error {
	manager, err := ffmpeg.NewManager()
	if err != nil {
		return err
	}

	// Write channel layouts to standard output
	return write(os.Stdout, manager.ChannelLayouts(), func(m media.Metadata) bool {
		return len(cmd.Name) == 0 || slices.Contains(cmd.Name, m.Key())
	})
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func write(w io.Writer, meta []media.Metadata, fn func(media.Metadata) bool) error {
	result := make(map[string]any, len(meta))
	for _, value := range meta {
		if fn != nil && !fn(value) {
			continue
		}
		result[value.Key()] = value.Any()
	}
	if len(result) == 0 {
		return fmt.Errorf("no results found")
	}

	// Print information in JSON format
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(data))
	return nil
}
