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
	SegmentSize      time.Duration // Segment size, zero means no segmenting
	SilenceSize      time.Duration // Size of silence to consider a segment boundary
	SilenceThreshold float64       // Silence threshold
}

///////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DefaultSilenceThreshold = 0.01                   // Default silence threshold
	DefaultSilenceDuration  = time.Millisecond * 500 // Default silence duration
	MinDuration             = time.Millisecond * 250 // Minimum duration
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

func WithSegmentSize(v time.Duration) Opt {
	return func(o *opts) error {
		if v < MinDuration {
			return media.ErrBadParameter.Withf("segment duration is too short, must be at least %v", MinDuration)
		} else {
			o.SegmentSize = v
		}
		return nil
	}
}

func WithSilenceSize(v time.Duration) Opt {
	return func(o *opts) error {
		if v < MinDuration {
			return media.ErrBadParameter.Withf("silence duration is too short, must be at least %v", MinDuration)
		} else {
			o.SilenceSize = v
		}
		return nil
	}
}

func WithDefaultSilenceThreshold() Opt {
	return func(o *opts) error {
		o.SilenceThreshold = DefaultSilenceThreshold
		o.SilenceSize = DefaultSilenceDuration
		return nil
	}
}
