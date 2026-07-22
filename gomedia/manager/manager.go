package manager

import (
	"context"
	"sync"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Media struct {
	opt

	sourcesMu sync.RWMutex
	sources   map[string]schema.Source
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new media object
func New(ctx context.Context, opts ...Opt) (_ *Media, err error) {
	self := new(Media)
	if err := self.apply(opts); err != nil {
		return nil, err
	}

	// Return the media manager
	return self, nil
}
