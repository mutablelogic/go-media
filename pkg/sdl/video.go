package sdl

import (
	"errors"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	sdl "github.com/veandco/go-sdl2/sdl"

	// Namespace imports
	. "github.com/mutablelogic/go-media"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Window struct {
	*sdl.Window
	*sdl.Renderer
	*sdl.Texture
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func (s *Context) NewVideo(title string, par *ffmpeg.Par) (*Window, error) {
	if !par.Type().Is(VIDEO) || par.Width() <= 0 || par.Height() <= 0 {
		return nil, errors.New("invalid video parameters")
	}
	window, err := sdl.CreateWindow(
		title,
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		int32(par.Width()), int32(par.Height()),
		sdl.WINDOW_SHOWN|sdl.WINDOW_BORDERLESS)
	if err != nil {
		return nil, err
	}
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		window.Destroy()
		return nil, err
	}
	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_IYUV, sdl.TEXTUREACCESS_STREAMING, int32(par.Width()), int32(par.Height()))
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
	w.Texture = nil
	w.Renderer = nil
	w.Window = nil

	// Return any errors
	return result
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (w *Window) Flush() error {
	if err := w.Renderer.Copy(w.Texture, nil, nil); err != nil {
		return err
	}
	w.Renderer.Present()
	return nil
}

func (w *Window) RenderFrame(frame *ffmpeg.Frame) error {
	if err := w.UpdateYUV(
		nil,
		frame.Bytes(0),
		frame.Stride(0),
		frame.Bytes(1),
		frame.Stride(1),
		frame.Bytes(2),
		frame.Stride(2),
	); err != nil {
		return err
	}

	// Return success
	return nil
}
