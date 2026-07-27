package manager

import (
	"context"

	// Packages
	task "github.com/mutablelogic/go-media/gomedia/task"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Media struct {
	opt
	tasks *task.Manager
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new media object
func New(ctx context.Context, opts ...Opt) (_ *Media, err error) {
	self := new(Media)
	if err := self.apply(opts); err != nil {
		return nil, err
	}
	self.tasks = task.NewManager(self.opt.tracer)

	// Return the media manager
	return self, nil
}
