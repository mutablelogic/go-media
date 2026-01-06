package sdl

import (
	"context"
	"errors"
	"fmt"

	// Packages
	"github.com/veandco/go-sdl3/sdl"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Context manages SDL initialization and the event loop.
type Context struct {
	flags  sdl.InitFlags
	events map[uint32]func(any)
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new SDL context with the specified initialization flags.
// Common flags are sdl.INIT_VIDEO, sdl.INIT_AUDIO, or combine them.
func New(flags sdl.InitFlags) (*Context, error) {
	if err := sdl.Init(flags); err != nil {
		return nil, fmt.Errorf("sdl.Init: %w", err)
	}

	return &Context{
		flags:  flags,
		events: make(map[uint32]func(any)),
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
func (c *Context) Register(fn func(userInfo any)) uint32 {
	evt := sdl.RegisterEvents(1)
	c.events[evt] = fn
	return evt
}

// Post posts a custom event to the event queue.
func (c *Context) Post(evt uint32, userInfo any) error {
	return sdl.PushEvent(&sdl.UserEvent{
		Type: evt,
		Data: userInfo,
	})
}

// Run starts the SDL event loop and blocks until the context is cancelled
// or a quit event is received.
func (c *Context) Run(ctx context.Context) error {
	// Register an event that quits when context is cancelled
	evtCancel := sdl.RegisterEvents(1)
	go func() {
		<-ctx.Done()
		sdl.PushEvent(&sdl.UserEvent{
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

		// Wait for an event with timeout
		event, err := sdl.WaitEventTimeout(100)
		if err != nil {
			if errors.Is(err, sdl.ErrTimedOut) {
				continue
			}
			return fmt.Errorf("sdl.WaitEventTimeout: %w", err)
		}

		if event == nil {
			continue
		}

		// Handle event
		switch e := event.(type) {
		case *sdl.QuitEvent:
			return nil
		case *sdl.UserEvent:
			if e.Type == evtCancel {
				return ctx.Err()
			}
			if handler, exists := c.events[e.Type]; exists {
				handler(e.Data)
			}
		}
	}
}
