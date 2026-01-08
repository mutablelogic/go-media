//go:build sdl2

package main

import (
	"context"
	"fmt"
	"os"
	"time"

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

	// Create SDL context
	ctx, err := sdl.New(sdlraw.INIT_VIDEO | sdlraw.INIT_AUDIO)
	if err != nil {
		return fmt.Errorf("sdl.New: %w", err)
	}
	defer ctx.Close()

	// Open media file to get metadata for window creation
	var reader *ffmpeg.Reader
	var audioStreamIdx = -1
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
	var bufferStreams []*schema.Stream

	if streams := reader.Streams(media.VIDEO); len(streams) > 0 {
		bufferStreams = append(bufferStreams, streams[0])
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
	if streams := reader.Streams(media.AUDIO); len(streams) > 0 {
		bufferStreams = append(bufferStreams, streams[0])
		audioStreamIdx = streams[0].Index()
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
	}, 512 /*buffer*/, sdl.WithFrameDelayFunc(func(*ffmpeg.Frame) time.Duration { return 0 }))
	if err != nil {
		return fmt.Errorf("frame loop: %w", err)
	}
	loop.Start()
	defer loop.Stop()

	// Buffer will be created inside decoder goroutine after opening the file
	var buffer ffmpeg.FrameBuffer
	bufferReady := make(chan struct{})

	cancelScheduler := startScheduler(globals.ctx, bufferReady, &buffer, audioStreamIdx, loop)
	defer cancelScheduler()

	errCh := make(chan error, 1)
	startDecoder(globals.ctx, cmd, bufferReady, &buffer, cancelScheduler, loop, errCh)

	// Run SDL event loop on main thread
	fmt.Printf("[STARTUP] Starting SDL event loop on main thread...\n")
	if err := ctx.Run(globals.ctx); err != nil && err != context.Canceled {
		return fmt.Errorf("SDL event loop: %w", err)
	}

	// Wait for decode to finish
	if err := <-errCh; err != nil && err != context.Canceled {
		return err
	}

	return nil
}

// startScheduler launches a goroutine that paces frames based on a master clock
// and enqueues them onto the FrameLoop.
func startScheduler(parent context.Context, bufferReady <-chan struct{}, bufferPtr *ffmpeg.FrameBuffer, audioStreamIdx int, loop *sdl.FrameLoop) context.CancelFunc {
	schedulerCtx, cancel := context.WithCancel(parent)

	go func() {
		fmt.Printf("[SCHEDULER] Waiting for buffer...\n")
		<-bufferReady
		buffer := *bufferPtr
		if buffer == nil {
			fmt.Printf("[SCHEDULER] Buffer not initialized, exiting\n")
			return
		}
		fmt.Printf("[SCHEDULER] Buffer ready, waiting for initial frames...\n")

		for {
			if schedulerCtx.Err() != nil {
				return
			}
			stats := buffer.Stats()
			if stats.TotalFrames >= 4 || stats.AllClosed {
				fmt.Printf("[SCHEDULER] Buffer has %d frames, starting playback\n", stats.TotalFrames)
				break
			}
			time.Sleep(5 * time.Millisecond)
		}

		frameCount := 0
		noFrameCount := 0
		anchorSet := false
		anchorPTS := int64(0)
		anchorWall := time.Time{}
		masterStream := audioStreamIdx

		for {
			select {
			case <-schedulerCtx.Done():
				fmt.Printf("[SCHEDULER] Context cancelled, exiting\n")
				return
			default:
			}

			t0 := time.Now()
			schemaFrame, err := buffer.Next(-1)
			if wait := time.Since(t0); wait > 50*time.Millisecond {
				fmt.Printf("[SCHEDULER] buffer.Next() blocked for %v\n", wait)
			}
			if err != nil {
				fmt.Printf("[SCHEDULER] buffer.Next() error: %v\n", err)
				return
			}
			if schemaFrame == nil {
				noFrameCount++
				if noFrameCount%100 == 0 {
					fmt.Printf("[SCHEDULER] No frame ready (checked %d times), buffer stats: %+v\n", noFrameCount, buffer.Stats())
				}
				time.Sleep(5 * time.Millisecond)
				continue
			}
			noFrameCount = 0

			ffmpegFrame := (*ffmpeg.Frame)(schemaFrame.AVFrame)
			framePtsMs := schemaFrame.Pts
			copyFrame, err := ffmpegFrame.Copy()
			if err != nil {
				schemaFrame.Unref()
				fmt.Printf("[SCHEDULER] frame copy failed: %v\n", err)
				return
			}
			frameCount++

			if !anchorSet {
				if masterStream == -1 || schemaFrame.Stream == masterStream {
					anchorSet = true
					anchorPTS = schemaFrame.Pts
					anchorWall = time.Now()
					if masterStream == -1 {
						masterStream = schemaFrame.Stream
					}
				}
			}

			delta := time.Duration(0)
			if anchorSet {
				target := anchorWall.Add(time.Duration(framePtsMs-anchorPTS) * time.Millisecond)
				delta = time.Until(target)
			}

			if delta > 0 {
				time.Sleep(delta)
			} else if delta < 0 {
				late := -delta
				if late > 200*time.Millisecond && copyFrame.Type() == media.VIDEO {
					copyFrame.Close()
					schemaFrame.Unref()
					continue
				}
			}

			if schedulerCtx.Err() != nil {
				copyFrame.Close()
				schemaFrame.Unref()
				return
			}

			if err := loop.Enqueue(copyFrame); err != nil {
				copyFrame.Close()
				schemaFrame.Unref()
				if err.Error() == "frame queue full" {
					continue
				}
				if err.Error() == "frame loop stopped" {
					return
				}
				fmt.Printf("[SCHEDULER] Enqueue failed: %v\n", err)
				return
			}

			schemaFrame.Unref()

			if frameCount%100 == 0 {
				fmt.Printf("[SCHEDULER] Posted %d frames, buffer: %+v\n", frameCount, buffer.Stats())
			}
		}
	}()

	return cancel
}

// startDecoder launches the decode goroutine that fills the FrameBuffer and signals readiness.
func startDecoder(ctx context.Context, cmd *PlayCommand, bufferReady chan struct{}, bufferPtr *ffmpeg.FrameBuffer, cancelScheduler context.CancelFunc, loop *sdl.FrameLoop, errCh chan<- error) {
	fmt.Printf("[STARTUP] Starting decoder goroutine...\n")

	go func() {
		fmt.Printf("[DECODER] Started\n")

		var reader *ffmpeg.Reader
		var err error
		opt := ffmpeg.WithInput(cmd.DecodeRequest.InputFormat, cmd.DecodeRequest.InputOpts...)
		if cmd.Reader != nil {
			reader, err = ffmpeg.NewReader(cmd.Reader, opt)
		} else {
			reader, err = ffmpeg.Open(cmd.Input, opt)
		}
		if err != nil {
			fmt.Printf("[DECODER] Failed to open: %v\n", err)
			loop.CloseInput()
			errCh <- fmt.Errorf("open input: %w", err)
			return
		}
		defer reader.Close()
		fmt.Printf("[DECODER] File opened successfully\n")

		bufferTimebase := ff.AVUtil_rational_d2q(1.0/1000.0, 0)
		var bufferStreams []*schema.Stream
		if streams := reader.Streams(media.VIDEO); len(streams) > 0 {
			bufferStreams = append(bufferStreams, streams[0])
		}
		if streams := reader.Streams(media.AUDIO); len(streams) > 0 {
			bufferStreams = append(bufferStreams, streams[0])
		}

		fb, err2 := ffmpeg.NewFrameBuffer(bufferTimebase, 3*time.Second, bufferStreams...)
		if err2 != nil {
			close(bufferReady)
			errCh <- fmt.Errorf("create frame buffer: %w", err2)
			return
		}
		*bufferPtr = fb
		fmt.Printf("[DECODER] Buffer created with %d streams\n", len(bufferStreams))
		for i, s := range bufferStreams {
			fmt.Printf("[DECODER]   Stream %d: index=%d\n", i, s.Index())
		}
		close(bufferReady)

		mapfn := func(streamIndex int, srcPar *ffmpeg.Par) (*ffmpeg.Par, error) {
			fmt.Printf("[DECODER] Mapping stream %d, type=%v\n", streamIndex, srcPar.Type())
			switch srcPar.Type() {
			case media.VIDEO:
				size := fmt.Sprintf("%dx%d", srcPar.Width(), srcPar.Height())
				return ffmpeg.NewVideoPar("yuv420p", size, 0)
			case media.AUDIO:
				chLayout := srcPar.ChannelLayout()
				chLayoutStr, err := ff.AVUtil_channel_layout_describe(&chLayout)
				if err != nil {
					return nil, err
				}
				return ffmpeg.NewAudioPar("fltp", chLayoutStr, srcPar.SampleRate())
			default:
				return nil, nil
			}
		}

		framesPushed := 0
		err = reader.Demux(ctx, mapfn, func(streamIndex int, frame *ffmpeg.Frame) error {
			schemaFrame := schema.NewFrame((*ff.AVFrame)(frame), streamIndex)

			for {
				pushErr := fb.Push(schemaFrame)
				if pushErr == nil {
					break
				}
				if pushErr == ffmpeg.ErrBufferFull {
					select {
					case <-ctx.Done():
						schemaFrame.Unref()
						return ctx.Err()
					case <-time.After(10 * time.Millisecond):
					}
					continue
				}
				fmt.Printf("[DECODER] Push error: %v\n", pushErr)
				schemaFrame.Unref()
				return pushErr
			}

			framesPushed++
			schemaFrame.Unref()
			return nil
		}, nil)

		fmt.Printf("[DECODER] Decode complete, pushed %d frames total\n", framesPushed)
		loop.CloseInput()
		cancelScheduler()
		errCh <- err
	}()
}
