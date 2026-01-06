package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	// Packages
	kong "github.com/alecthomas/kong"
	"github.com/mutablelogic/go-media"
	chromaprintschema "github.com/mutablelogic/go-media/pkg/chromaprint/schema"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	task "github.com/mutablelogic/go-media/pkg/ffmpeg/task"
	sdl "github.com/mutablelogic/go-media/pkg/sdl"
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

	// Create SDL context
	ctx, err := sdl.New(sdlraw.INIT_VIDEO | sdlraw.INIT_AUDIO)
	if err != nil {
		return fmt.Errorf("sdl.New: %w", err)
	}
	defer ctx.Close()

	// Open media file to get metadata for window creation
	var reader *ffmpeg.Reader
	if cmd.Reader != nil {
		if cmd.Input != "" {
			reader, err = ffmpeg.NewReader(cmd.Reader, ffmpeg.WithInput(cmd.Input))
		} else {
			reader, err = ffmpeg.NewReader(cmd.Reader)
		}
	} else {
		reader, err = ffmpeg.Open(cmd.Input)
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

	// Create frame queue channel (larger buffer)
	frameCh := make(chan *ffmpeg.Frame, 100)
	decodeDone := make(chan struct{})
	var doneOnce sync.Once
	frameCount := 0
	var frameEvent uint32

	// Register SDL event for processing frames on main thread
	var eventStopped uint32 // atomic flag to stop posting new events
	frameEvent = ctx.Register(func(userInfo interface{}) {
		// Check if we should stop
		if atomic.LoadUint32(&eventStopped) != 0 {
			return
		}

		// Process ONE frame per event (not all queued frames)
		select {
		case frame, ok := <-frameCh:
			if !ok {
				// Channel closed, decode is done
				atomic.StoreUint32(&eventStopped, 1) // Stop posting events
				doneOnce.Do(func() { close(decodeDone) })
				return
			}
			frameCount++

			// Process frame on main thread
			if err := player.PlayFrame(ctx, frame); err != nil {
				// Skip frames with errors (like invalid planes)
				// Schedule next frame immediately if not stopped
				go func() {
					if atomic.LoadUint32(&eventStopped) == 0 {
						time.Sleep(1 * time.Millisecond)
						ctx.Post(frameEvent, nil)
					}
				}()
				return
			}

			// Render video if we have a window
			if window := player.Window(); window != nil {
				if err := window.Render(); err != nil {
					fmt.Fprintf(os.Stderr, "render error: %v\n", err)
					atomic.StoreUint32(&eventStopped, 1)
					doneOnce.Do(func() { close(decodeDone) })
					return
				}
			}

			// Always schedule next frame check with timing delay (~30fps) in background
			go func() {
				if atomic.LoadUint32(&eventStopped) == 0 {
					time.Sleep(33 * time.Millisecond)
					ctx.Post(frameEvent, nil)
				}
			}()
		default:
			// No frame ready yet, check again soon
			go func() {
				if atomic.LoadUint32(&eventStopped) == 0 {
					time.Sleep(10 * time.Millisecond)
					ctx.Post(frameEvent, nil)
				}
			}()
		}
	})

	// Create frame writer that queues frames
	frameWriter := &playFrameWriter{
		ctx:        ctx,
		frameCh:    frameCh,
		frameEvent: frameEvent,
	}

	// Start decode in background (SDL must run on main thread on macOS)
	errCh := make(chan error, 1)
	go func() {
		// Wait a bit for SDL event loop to start
		time.Sleep(100 * time.Millisecond)

		// Start decoding
		err := globals.manager.Decode(globals.ctx, frameWriter, &cmd.DecodeRequest)
		close(frameCh) // Signal completion

		// Post one more event to trigger final frame processing
		ctx.Post(frameEvent, nil)
		errCh <- err
	}()

	// Kick off the frame processing loop after SDL starts
	go func() {
		time.Sleep(200 * time.Millisecond)
		ctx.Post(frameEvent, nil)
	}()

	// Wait for decode to finish (but keep window open)
	go func() {
		<-decodeDone
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

///////////////////////////////////////////////////////////////////////////////
// FRAME WRITER FOR PLAYBACK

type playFrameWriter struct {
	ctx        *sdl.Context
	frameCh    chan *ffmpeg.Frame
	frameEvent uint32
	frameCount int
	videoCount int
	audioCount int
}

// Write satisfies io.Writer but discards output (we only use WriteFrame for playback)
func (w *playFrameWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (w *playFrameWriter) WriteFrame(streamIndex int, frame interface{}) error {
	f, ok := frame.(*ffmpeg.Frame)
	if !ok {
		return nil
	}

	w.frameCount++
	frameType := f.Type()
	switch frameType {
	case 1: // AUDIO
		w.audioCount++
	case 2: // VIDEO
		w.videoCount++
	}

	// Queue frame for processing on main thread
	w.frameCh <- f

	// Trigger SDL event to process frames on main thread
	if err := w.ctx.Post(w.frameEvent, nil); err != nil {
		return fmt.Errorf("post frame event: %w", err)
	}

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
