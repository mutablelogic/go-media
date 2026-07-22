package writer

import (
	profile "github.com/mutablelogic/go-media/profile/schema"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
}

type Opt func(o *opt) error

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func WithProfile(id int, profile *profile.Output) Opt {
	return func(o *opt) error {
		return nil
	}
}
