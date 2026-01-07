//go:build sdl2

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	sdl "github.com/mutablelogic/go-media/pkg/sdl"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	sdlraw "github.com/veandco/go-sdl2/sdl"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type PlayCommands struct {
	Play PlayCommand `cmd:"" help:"Play media file with SDL" group:"FILE"`
}

type PlayCommand struct {
	schema.DecodeRequest
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

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
