package manager

import (
	"context"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Manager struct {
	opt
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new manager object
func New(ctx context.Context, opts ...Opt) (_ *Manager, err error) {
	self := new(Manager)
	if err := self.opt.apply(opts); err != nil {
		return nil, err
	}

	return self, nil
}
