package manager

import (
	"context"
	"slices"
	"sync"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	schema "github.com/mutablelogic/go-media/task/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// subscriber pairs a callback with a unique key identifying its Subscribe
// call - func values aren't comparable, so removing a specific one from
// subscribers needs something else to search on.
type subscriber struct {
	key *struct{}
	fn  func(*schema.Event)
}

// broadcaster fans out task events (see schema.Event) to any number of
// subscribers. It has its own lifecycle, run alongside a Manager's own
// task-tracking loop, so every blocked Subscribe call unblocks either when
// its own ctx is done or when the broadcaster itself stops - whichever
// comes first.
type broadcaster struct {
	sync.Mutex
	subscribers []subscriber
	done        chan struct{} // closed once Run's ctx is done
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newBroadcaster() *broadcaster {
	return &broadcaster{
		done: make(chan struct{}),
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Run blocks until ctx is done, at which point every subscriber currently
// blocked in Subscribe unblocks and returns. Must be called exactly once.
func (b *broadcaster) Run(ctx context.Context) error {
	<-ctx.Done()
	close(b.done)
	return nil
}

// Subscribe registers fn to be called for every event emitted (via emit)
// until ctx is done or the broadcaster's own Run returns, whichever comes
// first, at which point fn is unregistered and Subscribe returns.
func (b *broadcaster) Subscribe(ctx context.Context, fn func(*schema.Event)) error {
	if fn == nil {
		return gomedia.ErrBadParameter.With("nil callback function")
	}

	// A fresh pointer makes a unique key without needing a counter - each
	// call gets its own distinct *struct{}, live for as long as this
	// subscription is, so DeleteFunc below only ever removes this one entry.
	key := new(struct{})

	b.Lock()
	b.subscribers = append(b.subscribers, subscriber{key: key, fn: fn})
	b.Unlock()

	defer func() {
		b.Lock()
		b.subscribers = slices.DeleteFunc(b.subscribers, func(s subscriber) bool {
			return s.key == key
		})
		b.Unlock()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.done:
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// emit calls every registered subscriber with a fresh Event. Doesn't
// block when fn is called
func (b *broadcaster) emit(name schema.EventName, status schema.Status) {
	b.Lock()
	if len(b.subscribers) == 0 {
		b.Unlock()
		return
	}
	fns := make([]func(*schema.Event), 0, len(b.subscribers))
	for _, s := range b.subscribers {
		fns = append(fns, s.fn)
	}
	b.Unlock()

	event := &schema.Event{Name: name, Status: status}
	for _, fn := range fns {
		fn(event)
	}
}
