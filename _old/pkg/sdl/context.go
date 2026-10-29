//go:build sdl2

package sdl

import (
	"context"
	"fmt"

	// Packages
	"github.com/veandco/go-sdl2/sdl"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Context manages SDL initialization and the event loop.
type Context struct {
	flags  uint32
	events map[uint32]func(interface{})
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new SDL context with the specified initialization flags.
// Common flags are sdl.INIT_VIDEO, sdl.INIT_AUDIO, or combine them.
func New(flags uint32) (*Context, error) {
	if err := sdl.Init(flags); err != nil {
		return nil, fmt.Errorf("sdl.Init: %w", err)
	}

	return &Context{
		flags:  flags,
		events: make(map[uint32]func(interface{})),
	}, nil
}

// Close shuts down SDL and releases all resources.
func (c *Context) Close() error {
	sdl.Quit()
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// EVENT HANDLING

// Register registers a custom event handler. Returns the event ID that
// can be used with Post to trigger the handler.
func (c *Context) Register(fn func(userInfo interface{})) uint32 {
	evt := sdl.RegisterEvents(1)
	c.events[evt] = fn
	return evt
}

// Post posts a custom event to the event queue.
func (c *Context) Post(evt uint32, userInfo interface{}) error {
	_, err := sdl.PushEvent(&sdl.UserEvent{
		Type: evt,
		Code: 0,
	})
	if err != nil {
		return fmt.Errorf("sdl.PushEvent: %w", err)
	}
	return nil
}

// Run starts the SDL event loop and blocks until the context is cancelled
// or a quit event is received.
func (c *Context) Run(ctx context.Context) error {
	// Register an event that quits when context is cancelled
	evtCancel := sdl.RegisterEvents(1)
	go func() {
		<-ctx.Done()
		_, _ = sdl.PushEvent(&sdl.UserEvent{
			Type: evtCancel,
		})
	}()

	// Event loop
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Pump events from the system queue to SDL's event queue
		sdl.PumpEvents()

		// Poll for events (non-blocking)
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			// Handle event
			switch e := event.(type) {
			case *sdl.QuitEvent:
				return nil
			case *sdl.UserEvent:
				if e.Type == evtCancel {
					return ctx.Err()
				}
				if handler, exists := c.events[e.Type]; exists {
					handler(nil)
				}
			}
		}

		// Small delay to avoid busy-waiting
		sdl.Delay(10)
	}
}
