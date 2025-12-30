package avfilter

import (
	"github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type opt struct {
	inputs  []*ff.AVFilterInOut
	outputs []*ff.AVFilterInOut
}

type Opt func(*opt) error

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func WithInput(filter, name string) Opt {
	return func(o *opt) error {
		return media.ErrNotImplemented
	}
}

func WithOutput(filter, name string) Opt {
	return func(o *opt) error {
		return media.ErrNotImplemented
	}
}
