package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	// Packages
	kong "github.com/alecthomas/kong"
	"github.com/mutablelogic/go-media"
	chromaprintschema "github.com/mutablelogic/go-media/pkg/chromaprint/schema"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	task "github.com/mutablelogic/go-media/pkg/ffmpeg/task"
	sdl "github.com/mutablelogic/go-media/pkg/sdl"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	sdlraw "github.com/veandco/go-sdl2/sdl"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Globals struct {
	Debug          bool   `name:"debug" help:"Enable debug logging"`
	ChromaprintKey string `name:"chromaprint-key" env:"CHROMAPRINT_KEY" help:"AcoustID API key for chromaprint lookups"`
	Endpoint       string `name:"url" env:"GOMEDIA_ENDPOINT" help:"Server endpoint URL"`

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
	ListFilters       ListFiltersCommand       `cmd:"" help:"List filters" group:"LIST"`
	Probe             ProbeCommand             `cmd:"" help:"Probe media file or stream" group:"FILE"`
	AudioLookup       AudioLookupCommand       `cmd:"" help:"Generate audio fingerprint and perform AcoustID lookup" group:"FILE"`
	Remux             RemuxCommand             `cmd:"" help:"Remux media file or stream" group:"FILE"`
	Decode            DecodeCommand            `cmd:"" help:"Decode media file or stream to JSON" group:"FILE"`
	Play              PlayCommand              `cmd:"" help:"Play media file with SDL" group:"FILE"`
	Server            ServerCommands           `cmd:"" help:"Run server."`
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

type ProbeCommand struct {
	schema.ProbeRequest
}

type AudioLookupCommand struct {
	chromaprintschema.AudioFingerprintRequest
}

type RemuxCommand struct {
	schema.RemuxRequest
}

type DecodeCommand struct {
	schema.DecodeRequest
}

type PlayCommand struct {
	schema.DecodeRequest
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

func (cmd *ProbeCommand) Run(globals *Globals) error {
	// If the path is "-", use stdin
	if cmd.Input == "-" {
		cmd.Reader = os.Stdin
		cmd.Input = ""
	}

	// Heuristic: if input is MPEG-TS, hint demuxer and enlarge probe/analyze windows for SPS/PPS
	if cmd.Input != "" && strings.EqualFold(filepath.Ext(cmd.Input), ".ts") {
		if cmd.ProbeRequest.InputFormat == "" {
			cmd.ProbeRequest.InputFormat = "mpegts"
		}
		if len(cmd.ProbeRequest.InputOpts) == 0 {
			cmd.ProbeRequest.InputOpts = []string{
				"probesize=5000000",        // 5MB probe
				"analyzeduration=10000000", // 10s analyze
				"fflags=+genpts",
				"discardcorrupt=1",
			}
		}
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

func (cmd *DecodeCommand) Run(globals *Globals) error {
	// If the path is "-", use stdin
	if cmd.Input == "-" {
		cmd.Reader = os.Stdin
		cmd.Input = ""
	}

	// Call manager method - outputs JSON to stdout
	err := globals.manager.Decode(globals.ctx, os.Stdout, &cmd.DecodeRequest)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *PlayCommand) Run(globals *Globals) error {
	// If the path is "-", use stdin
	if cmd.Input == "-" {
		cmd.Reader = os.Stdin
		cmd.Input = ""
	}

	// Heuristic: if input is MPEG-TS, hint demuxer and enlarge probe/analyze windows for SPS/PPS
	if cmd.Input != "" && strings.EqualFold(filepath.Ext(cmd.Input), ".ts") {
		if cmd.DecodeRequest.InputFormat == "" {
			cmd.DecodeRequest.InputFormat = "mpegts"
		}
		if len(cmd.DecodeRequest.InputOpts) == 0 {
			cmd.DecodeRequest.InputOpts = []string{
				"probesize=5000000",        // 5MB probe
				"analyzeduration=10000000", // 10s analyze
				"fflags=+genpts",
				"discardcorrupt=1",
			}
		}
	}

	// Create SDL context
	ctx, err := sdl.New(sdlraw.INIT_VIDEO | sdlraw.INIT_AUDIO)
	if err != nil {
		return fmt.Errorf("sdl.New: %w", err)
	}
	defer ctx.Close()

	// Open media file to get metadata for window creation
	var reader *ffmpeg.Reader
	opt := ffmpeg.WithInput(cmd.DecodeRequest.InputFormat, cmd.DecodeRequest.InputOpts...)
	if cmd.Reader != nil {
		reader, err = ffmpeg.NewReader(cmd.Reader, opt)
	} else {
		reader, err = ffmpeg.Open(cmd.Input, opt)
	}
	if err != nil {
		return fmt.Errorf("open input: %w", err)
	}

	// Get video stream dimensions to create window
	var window *sdl.Window
	if streams := reader.Streams(media.VIDEO); len(streams) > 0 {
		codecPar := streams[0].AVStream.CodecPar()
		if codecPar != nil {
			width := codecPar.Width()
			height := codecPar.Height()
			if width > 0 && height > 0 {
				window, err = ctx.NewWindow("Media Player", int32(width), int32(height))
				if err != nil {
					reader.Close()
					return fmt.Errorf("create window: %w", err)
				}
			}
		}
	}
	reader.Close()

	// Create player with the window
	player := ctx.NewPlayer()
	if window != nil {
		player.SetWindow(window)
	}
	defer player.Close()

	loop, err := sdl.NewFrameLoop(ctx, func(frame *ffmpeg.Frame) error {
		return player.PlayFrame(ctx, frame)
	}, 100, sdl.WithFrameDelayFunc(player.VideoDelay))
	if err != nil {
		return fmt.Errorf("frame loop: %w", err)
	}
	loop.Start()
	defer loop.Stop()

	frameWriter := sdl.NewFrameWriter(loop)

	// Start decode in background (SDL must run on main thread on macOS)
	errCh := make(chan error, 1)
	go func() {
		// Open the input reader for decoding
		var reader *ffmpeg.Reader
		var err error
		opt := ffmpeg.WithInput(cmd.DecodeRequest.InputFormat, cmd.DecodeRequest.InputOpts...)
		if cmd.Reader != nil {
			reader, err = ffmpeg.NewReader(cmd.Reader, opt)
		} else {
			reader, err = ffmpeg.Open(cmd.Input, opt)
		}
		if err != nil {
			loop.CloseInput()
			errCh <- fmt.Errorf("open input: %w", err)
			return
		}
		defer reader.Close()

		// Map function to configure decoder output formats for SDL compatibility
		mapfn := func(streamIndex int, srcPar *ffmpeg.Par) (*ffmpeg.Par, error) {
			switch srcPar.Type() {
			case media.VIDEO:
				// Convert video to yuv420p (SDL-compatible)
				size := fmt.Sprintf("%dx%d", srcPar.Width(), srcPar.Height())
				return ffmpeg.NewVideoPar("yuv420p", size, 0)
			case media.AUDIO:
				// Convert audio to planar float32 (SDL-compatible)
				chLayout := srcPar.ChannelLayout()
				chLayoutStr, err := ff.AVUtil_channel_layout_describe(&chLayout)
				if err != nil {
					return nil, err
				}
				return ffmpeg.NewAudioPar("fltp", chLayoutStr, srcPar.SampleRate())
			default:
				// Ignore other streams
				return nil, nil
			}
		}

		// Decode with resampling to SDL formats
		err = reader.Demux(globals.ctx, mapfn, func(streamIndex int, frame *ffmpeg.Frame) error {
			return frameWriter.WriteFrame(streamIndex, frame)
		}, nil)

		loop.CloseInput()
		errCh <- err
	}()

	// Run SDL event loop on main thread
	if err := ctx.Run(globals.ctx); err != nil && err != context.Canceled {
		return fmt.Errorf("SDL event loop: %w", err)
	}

	// Wait for decode to finish
	if err := <-errCh; err != nil && err != context.Canceled {
		return err
	}

	return nil
}

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
