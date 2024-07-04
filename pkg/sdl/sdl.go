package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	// Packages
	"github.com/mutablelogic/go-media"
	"github.com/mutablelogic/go-media/pkg/ffmpeg"
	sdl "github.com/veandco/go-sdl2/sdl"
)

type Context struct {
}

type Window struct {
	*sdl.Window
	*sdl.Renderer
	*sdl.Texture
}

type Surface sdl.Surface

// Create a new SDL object which can output audio and video
func NewSDL() (*Context, error) {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		return nil, err
	}
	return &Context{}, nil
}

func (s *Context) Close() error {
	sdl.Quit()
	return nil
}

func (s *Context) PushQuitEvent() {
	sdl.PushEvent(&sdl.QuitEvent{
		Type: sdl.QUIT,
	})
}

func (s *Context) NewWindow(title string, width, height int32) (*Window, error) {
	window, err := sdl.CreateWindow(
		title,
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		width, height,
		sdl.WINDOW_SHOWN|sdl.WINDOW_BORDERLESS)
	if err != nil {
		return nil, err
	}
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		window.Destroy()
		return nil, err
	}
	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_IYUV, sdl.TEXTUREACCESS_STREAMING, width, height)
	if err != nil {
		renderer.Destroy()
		window.Destroy()
		return nil, err
	}

	return &Window{window, renderer, texture}, nil
}

func (w *Window) Close() error {
	var result error
	if err := (*sdl.Texture)(w.Texture).Destroy(); err != nil {
		result = errors.Join(result, err)
	}
	if err := (*sdl.Renderer)(w.Renderer).Destroy(); err != nil {
		result = errors.Join(result, err)
	}
	if err := (*sdl.Window)(w.Window).Destroy(); err != nil {
		result = errors.Join(result, err)
	}
	w.Renderer = nil
	w.Window = nil

	// Return any errors
	return result
}

func (w *Window) Flush() error {
	if err := w.Renderer.Copy(w.Texture, nil, nil); err != nil {
		return err
	}
	w.Renderer.Present()
	return nil
}

func (w *Window) RenderFrame(frame *ffmpeg.Frame) error {
	return w.UpdateYUV(
		nil,
		frame.Bytes(0),
		frame.Stride(0),
		frame.Bytes(1),
		frame.Stride(1),
		frame.Bytes(2),
		frame.Stride(2),
	)
}

func (s *Context) RunLoop() {
	runtime.LockOSThread()
	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
				break
			}
		}
	}
}

func main() {
	sdl, err := NewSDL()
	if err != nil {
		log.Fatal(err)
	}
	defer sdl.Close()

	// Open video
	input, err := ffmpeg.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Decode frames in a goroutine
	var result error
	var wg sync.WaitGroup
	var w, h int32

	// Decoder map function
	mapfn := func(stream int, par *ffmpeg.Par) (*ffmpeg.Par, error) {
		if stream == input.BestStream(media.VIDEO) {
			w = int32(par.Width())
			h = int32(par.Height())
			return par, nil
		}
		return nil, nil
	}

	ch := make(chan *ffmpeg.Frame)

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := input.Decode(context.Background(), mapfn, func(stream int, frame *ffmpeg.Frame) error {
			ch <- frame
			return nil
		})
		if err != nil {
			result = errors.Join(result, err)
		}

		// Close channel
		close(ch)

		// Quit event
		sdl.PushQuitEvent()
	}()

	// HACK
	time.Sleep(100 * time.Millisecond)
	if w == 0 || h == 0 {
		log.Fatal("No video stream found")
	}

	title := filepath.Base(os.Args[1])
	meta := input.Metadata("title")
	if len(meta) > 0 {
		title = meta[0].Value()
	}

	window, err := sdl.NewWindow(title, w, h)
	if err != nil {
		log.Fatal(err)
	}
	defer window.Close()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for frame := range ch {
			if err := window.RenderFrame(frame); err != nil {
				fmt.Println("Error rendering frame:", err)
			}
			if err := window.Flush(); err != nil {
				fmt.Println("Error flushing frame:", err)
			}
		}
	}()

	// Run the SDL loop
	sdl.RunLoop()

	// Wait until all goroutines have finished
	wg.Wait()

	// Return any errors
	if result != nil {
		log.Fatal(result)
	}
}
