package main

import (

	// Packages
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"unsafe"

	"github.com/mutablelogic/go-media/pkg/ffmpeg"
	"github.com/mutablelogic/go-media/pkg/sdl"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

type Player struct {
	input      *ffmpeg.Reader
	ctx        *ffmpeg.Context
	audio      *ffmpeg.Par
	video      *ffmpeg.Par
	videoevent uint32
	audioevent uint32
}

func NewPlayer() *Player {
	return &Player{}
}

func (p *Player) Close() error {
	var result error

	// Close resources
	if p.ctx != nil {
		result = errors.Join(result, p.ctx.Close())
	}
	if p.input != nil {
		result = errors.Join(result, p.input.Close())
	}

	// Return any errors
	return result
}

func (p *Player) OpenUrl(url string) error {
	input, err := ffmpeg.Open(url)
	if err != nil {
		return err
	}
	p.input = input

	// Map input streams - find best audio and video streams
	ctx, err := p.input.Map(func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if stream == p.input.BestStream(VIDEO) {
			p.video = par
			return par, nil
		} else if stream == p.input.BestStream(AUDIO) {
			p.audio = par
			return par, nil
		} else {
			return nil, nil
		}
	})
	if err != nil {
		return err
	} else {
		p.ctx = ctx
		p.input = input
	}
	return nil
}

func (p *Player) Type() Type {
	t := NONE
	if p.video != nil {
		t |= VIDEO
	}
	if p.audio != nil {
		t |= AUDIO
	}
	return t
}

// Return media title
func (p *Player) Title() string {
	title := p.input.Metadata("title")
	if len(title) > 0 {
		return title[0].Value()
	}
	return fmt.Sprint(p.Type())
}

func (p *Player) Play(ctx context.Context) error {
	var window *sdl.Window
	var audio *sdl.Audio

	// Create a new SDL context
	sdl, err := sdl.New(p.Type())
	if err != nil {
		return err
	}
	defer sdl.Close()

	// Create a window for video
	if p.video != nil {
		if w, err := sdl.NewVideo(p.Title(), p.video); err != nil {
			return err
		} else {
			window = w
		}
		defer window.Close()

		// Register a method to push video rendering
		p.videoevent = sdl.Register(func(frame unsafe.Pointer) {
			var result error
			frame_ := (*ffmpeg.Frame)(frame)
			if err := window.RenderFrame(frame_); err != nil {
				result = errors.Join(result, err)
			}
			if err := window.Flush(); err != nil {
				result = errors.Join(result, err)
			}
			if err := frame_.Close(); err != nil {
				result = errors.Join(result, err)
			}
			if result != nil {
				fmt.Fprintln(os.Stderr, result)
			}
			/*
				// Pause to present the frame at the correct PTS
				if pts != ffmpeg.TS_UNDEFINED && pts < frame.Ts() {
					pause := frame.Ts() - pts
					if pause > 0 {
						sdl.Delay(uint32(pause * 1000))
					}
				}

				// Set current timestamp
				pts = frame.Ts()

				// Render the frame, release the frame resources
				if err := w.RenderFrame(frame); err != nil {
					log.Print(err)
				} else if err := w.Flush(); err != nil {
					log.Print(err)
				} else if err := frame.Close(); err != nil {
					log.Print(err)
				}
			*/
		})
	}

	// Create audio
	if p.audio != nil {
		if a, err := sdl.NewAudio(p.audio); err != nil {
			return err
		} else {
			audio = a
		}
		defer audio.Close()

		// Register a method to push audio rendering
		p.audioevent = sdl.Register(func(frame unsafe.Pointer) {
			//frame_ := (*ffmpeg.Frame)(frame)
			//fmt.Println("TODO: Audio", frame_)
		})
	}

	// Start go routine to decode the audio and video frames
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := p.decode(ctx, sdl); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	// Run loop with events for audio and video
	var result error
	if err := sdl.Run(ctx); err != nil {
		result = err
	}

	// Wait for go routines to finish
	wg.Wait()

	// Return any errors
	return result
}

// Goroutine decoder
func (p *Player) decode(ctx context.Context, sdl *sdl.Context) error {
	return p.input.DecodeWithContext(ctx, p.ctx, func(stream int, frame *ffmpeg.Frame) error {
		if frame.Type().Is(VIDEO) {
			if copy, err := frame.Copy(); err != nil {
				fmt.Println("Unable to make a frame copy: ", err)
			} else {
				// TODO: Make a copy of the frame
				sdl.Post(p.videoevent, unsafe.Pointer(copy))
			}
		}
		if frame.Type().Is(AUDIO) {
			// TODO: Make a copy of the frame
			sdl.Post(p.audioevent, unsafe.Pointer(frame))
		}
		return nil
	})
}
