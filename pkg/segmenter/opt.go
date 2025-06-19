package segmenter

import (
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
)

///////////////////////////////////////////////////////////////////////////////////
// TYPES

type Opt func(*opts) error

type opts struct {
	SilenceThreshold float64       // Silence threshold
	SilenceDuration  time.Duration // Duration of silence to consider a segment boundary
}

///////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DefaultSilenceThreshold = 0.0005          // Default silence threshold
	DefaultSilenceDuration  = time.Second * 2 // Default silence duration
)

///////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func applyOpts(opt ...Opt) (*opts, error) {
	var o opts
	for _, fn := range opt {
		if err := fn(&o); err != nil {
			return nil, err
		}
	}
	return &o, nil
}

///////////////////////////////////////////////////////////////////////////////////
// TYPES

func WithDefaultSilenceThreshold() Opt {
	return func(o *opts) error {
		o.SilenceThreshold = DefaultSilenceThreshold
		o.SilenceDuration = DefaultSilenceDuration
		return nil
	}
}

func WithSilenceDuration(v time.Duration) Opt {
	return func(o *opts) error {
		if v < time.Millisecond*100 {
			return media.ErrBadParameter.Withf("silence duration %s is too short, must be at least 100ms", v)
		} else {
			o.SilenceDuration = v
		}
		return nil
	}
}
