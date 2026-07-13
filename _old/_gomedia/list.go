package main

import (
	"fmt"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type ListCommands struct {
	ListAudioChannels ListAudioChannelsCommand `cmd:"" help:"List audio channel layouts" group:"LIST"`
	ListCodecs        ListCodecsCommand        `cmd:"" help:"List codecs" group:"LIST"`
	ListFormats       ListFormatsCommand       `cmd:"" help:"List input, output formats and devices" group:"LIST"`
	ListPixelFormats  ListPixelFormatsCommand  `cmd:"" help:"List pixel formats" group:"LIST"`
	ListSampleFormats ListSampleFormatsCommand `cmd:"" help:"List sample formats" group:"LIST"`
	ListFilters       ListFiltersCommand       `cmd:"" help:"List filters" group:"LIST"`
}

type ListAudioChannelsCommand struct {
	schema.ListAudioChannelLayoutRequest
}

type ListCodecsCommand struct {
	schema.ListCodecRequest
}

type ListFiltersCommand struct {
	schema.ListFilterRequest
}

type ListFormatsCommand struct {
	schema.ListFormatRequest
}

type ListPixelFormatsCommand struct {
	schema.ListPixelFormatRequest
}

type ListSampleFormatsCommand struct {
	schema.ListSampleFormatRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ListAudioChannelsCommand) Run(globals *Globals) error {
	// Call manager method
	response, err := globals.manager.ListAudioChannelLayouts(globals.ctx, &cmd.ListAudioChannelLayoutRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}

func (cmd *ListCodecsCommand) Run(globals *Globals) error {
	// Call manager method
	response, err := globals.manager.ListCodecs(globals.ctx, &cmd.ListCodecRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}

func (cmd *ListFiltersCommand) Run(globals *Globals) error {
	// Call manager method
	response, err := globals.manager.ListFilters(globals.ctx, &cmd.ListFilterRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}

func (cmd *ListFormatsCommand) Run(globals *Globals) error {
	// Call manager method
	response, err := globals.manager.ListFormats(globals.ctx, &cmd.ListFormatRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}

func (cmd *ListPixelFormatsCommand) Run(globals *Globals) error {
	// Call manager method
	response, err := globals.manager.ListPixelFormats(globals.ctx, &cmd.ListPixelFormatRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}

func (cmd *ListSampleFormatsCommand) Run(globals *Globals) error {
	// Call manager method
	response, err := globals.manager.ListSampleFormats(globals.ctx, &cmd.ListSampleFormatRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}
