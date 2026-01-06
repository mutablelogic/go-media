package sdl

import (
	"errors"
	"fmt"

	// Packages
	"github.com/veandco/go-sdl3/sdl"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Window represents an SDL window with renderer and texture for video display.
type Window struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	width    int32
	height   int32
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewWindow creates a new SDL window with the specified title and dimensions.
// The window is created with a renderer and texture ready for video display.
func (c *Context) NewWindow(title string, width, height int32) (*Window, error) {
	if width <= 0 || height <= 0 {
		return nil, errors.New("invalid window dimensions")
	}

	// Create window
	window, err := sdl.CreateWindow(
		title,
		width, height,
		sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE,
	)
	if err != nil {
		return nil, fmt.Errorf("sdl.CreateWindow: %w", err)
	}

	// Create renderer
	renderer, err := sdl.CreateRenderer(window, nil)
	if err != nil {
		window.Destroy()
		return nil, fmt.Errorf("sdl.CreateRenderer: %w", err)
	}

	// Create texture for video frames (YUV format)
	texture, err := sdl.CreateTexture(
		renderer,
		sdl.PIXELFORMAT_IYUV,
		sdl.TEXTUREACCESS_STREAMING,
		width, height,
	)
	if err != nil {
		renderer.Destroy()
		window.Destroy()
		return nil, fmt.Errorf("sdl.CreateTexture: %w", err)
	}

	return &Window{
		window:   window,
		renderer: renderer,
		texture:  texture,
		width:    width,
		height:   height,
	}, nil
}

// Close destroys the window and releases all resources.
func (w *Window) Close() error {
	var result error

	if w.texture != nil {
		if err := w.texture.Destroy(); err != nil {
			result = errors.Join(result, err)
		}
		w.texture = nil
	}

	if w.renderer != nil {
		if err := w.renderer.Destroy(); err != nil {
			result = errors.Join(result, err)
		}
		w.renderer = nil
	}

	if w.window != nil {
		if err := w.window.Destroy(); err != nil {
			result = errors.Join(result, err)
		}
		w.window = nil
	}

	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Update updates the texture with new video frame data.
// The data should be in YUV420P format with planes laid out sequentially.
func (w *Window) Update(y, u, v []byte, yPitch, uPitch, vPitch int) error {
	if err := w.texture.UpdateYUV(nil, y, yPitch, u, uPitch, v, vPitch); err != nil {
		return fmt.Errorf("texture.UpdateYUV: %w", err)
	}
	return nil
}

// UpdateRGB updates the texture with RGB24 data.
// The data should be in packed RGB24 format (3 bytes per pixel).
func (w *Window) UpdateRGB(pixels []byte, pitch int) error {
	if err := w.texture.Update(nil, pixels, pitch); err != nil {
		return fmt.Errorf("texture.Update: %w", err)
	}
	return nil
}

// Render renders the current texture to the window.
func (w *Window) Render() error {
	if err := w.renderer.Clear(); err != nil {
		return fmt.Errorf("renderer.Clear: %w", err)
	}

	if err := w.renderer.Copy(w.texture, nil, nil); err != nil {
		return fmt.Errorf("renderer.Copy: %w", err)
	}

	if err := w.renderer.Present(); err != nil {
		return fmt.Errorf("renderer.Present: %w", err)
	}

	return nil
}

// Size returns the current window dimensions.
func (w *Window) Size() (width, height int32) {
	return w.width, w.height
}

// SetTitle sets the window title.
func (w *Window) SetTitle(title string) error {
	return w.window.SetTitle(title)
}
