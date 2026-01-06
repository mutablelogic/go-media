package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	// Packages
	kong "github.com/alecthomas/kong"
	chromaprintschema "github.com/mutablelogic/go-media/pkg/chromaprint/schema"
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	task "github.com/mutablelogic/go-media/pkg/ffmpeg/task"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Globals struct {
	Debug          bool   `name:"debug" help:"Enable debug logging"`
	ChromaprintKey string `name:"chromaprint-key" env:"CHROMAPRINT_KEY" help:"AcoustID API key for chromaprint lookups"`

	// Private fields
	ctx     context.Context
	cancel  context.CancelFunc
	manager *task.Manager
}

type CLI struct {
	Globals
	ListAudioChannels ListAudioChannelsCommand `cmd:"" help:"List audio channel layouts" group:"LIST"`
	ListCodecs        ListCodecsCommand        `cmd:"" help:"List codecs" group:"LIST"`
	ListFormats       ListFormatsCommand       `cmd:"" help:"List input, output formats and devices" group:"LIST"`
	ListPixelFormats  ListPixelFormatsCommand  `cmd:"" help:"List pixel formats" group:"LIST"`
	ListSampleFormats ListSampleFormatsCommand `cmd:"" help:"List sample formats" group:"LIST"`
	Probe             ProbeCommand             `cmd:"" help:"Probe media file or stream" group:"FILE"`
	AudioLookup       AudioLookupCommand       `cmd:"" help:"Generate audio fingerprint and perform AcoustID lookup" group:"FILE"`
	Remux             RemuxCommand             `cmd:"" help:"Remux media file or stream" group:"FILE"`
}

type ListAudioChannelsCommand struct {
	schema.ListAudioChannelLayoutRequest
}

type ListCodecsCommand struct {
	schema.ListCodecRequest
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

type ProbeCommand struct {
	schema.ProbeRequest
}

type AudioLookupCommand struct {
	chromaprintschema.AudioFingerprintRequest
}

type RemuxCommand struct {
	schema.RemuxRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (cmd *ListAudioChannelsCommand) Run(globals *Globals) error {
	// Call manager method
	response, err := globals.manager.ListAudioChannelLayout(globals.ctx, &cmd.ListAudioChannelLayoutRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}

func (cmd *ListCodecsCommand) Run(globals *Globals) error {
	// Call manager method
	response, err := globals.manager.ListCodec(globals.ctx, &cmd.ListCodecRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}

func (cmd *ListFormatsCommand) Run(globals *Globals) error {
	// Call manager method
	response, err := globals.manager.ListFormat(globals.ctx, &cmd.ListFormatRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}

func (cmd *ListPixelFormatsCommand) Run(globals *Globals) error {
	// Call manager method
	response, err := globals.manager.ListPixelFormat(globals.ctx, &cmd.ListPixelFormatRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}

func (cmd *ListSampleFormatsCommand) Run(globals *Globals) error {
	// Call manager method
	response, err := globals.manager.ListSampleFormat(globals.ctx, &cmd.ListSampleFormatRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}

func (cmd *ProbeCommand) Run(globals *Globals) error {
	// If the path is "-", use stdin
	if cmd.Input == "-" {
		cmd.Reader = os.Stdin
		cmd.Input = ""
	}

	// Call manager method
	response, err := globals.manager.Probe(globals.ctx, &cmd.ProbeRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}

func (cmd *AudioLookupCommand) Run(globals *Globals) error {
	// If the path is "-", use stdin
	if cmd.Input == "-" {
		cmd.Reader = os.Stdin
		cmd.Input = ""
	}

	// Call manager method
	response, err := globals.manager.AudioFingerprint(globals.ctx, &cmd.AudioFingerprintRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}

func (cmd *RemuxCommand) Run(globals *Globals) error {
	// If the path is "-", use stdin
	if cmd.Input == "-" {
		cmd.Reader = os.Stdin
		cmd.Input = ""
	}

	// Call manager method
	response, err := globals.manager.Remux(globals.ctx, os.Stdout, &cmd.RemuxRequest)
	if err != nil {
		return err
	}

	// Print response
	fmt.Println(response)
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// TYPES

func main() {
	cli := new(CLI)
	ctx := kong.Parse(cli,
		kong.UsageOnError(),
	)

	// Create the context and cancel function
	cli.Globals.ctx, cli.Globals.cancel = context.WithCancel(context.Background())
	defer cli.Globals.cancel()

	// Set up signal handling to force exit on CTRL+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nInterrupted")
		cli.Globals.cancel()
		os.Exit(130) // Standard exit code for CTRL+C
	}()

	// Set options
	opts := []task.Opt{}
	opts = append(opts, task.WithTraceFn(func(msg string) {
		fmt.Fprintln(os.Stderr, msg)
	}, cli.Globals.Debug))
	if cli.Globals.ChromaprintKey != "" {
		opts = append(opts, task.WithChromaprintKey(cli.Globals.ChromaprintKey))
	}

	// Create manager
	manager, err := task.NewManager(opts...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	} else {
		cli.Globals.manager = manager
	}

	// Call the Run() method of the selected parsed command.
	if err := ctx.Run(&cli.Globals); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
