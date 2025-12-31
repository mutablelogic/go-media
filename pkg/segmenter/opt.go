package segmenter

import (
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
)

// WithFFMpegOpt wraps an ffmpeg.Opt as a segmenter.Opt for use with segmenter.NewFromReader
func WithFFMpegOpt(opt interface{}) Opt {
	return func(o *opts) error {
		if fn, ok := opt.(func(*opts) error); ok {
			return fn(o)
		}
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is a function which applies options to a Segmenter
type Opt func(*opts) error

type opts struct {
	SegmentSize      time.Duration // Segment size, zero means no fixed segmenting
	SilenceSize      time.Duration // Size of silence to consider a segment boundary
	SilenceThreshold float64       // Silence threshold (RMS energy 0.0-1.0)
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DefaultSilenceThreshold = 0.01                   // Default silence threshold (RMS)
	DefaultSilenceDuration  = time.Millisecond * 500 // Default silence duration
	MinSegmentDuration      = time.Millisecond * 100 // Minimum segment/silence duration
)

///////////////////////////////////////////////////////////////////////////////
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

///////////////////////////////////////////////////////////////////////////////
// OPTIONS

// WithSegmentSize sets the target segment duration. Segments will be output
// when they reach approximately this duration. Minimum is 100ms.
func WithSegmentSize(v time.Duration) Opt {
	return func(o *opts) error {
		if v < MinSegmentDuration {
			return media.ErrBadParameter.Withf("segment duration is too short, must be at least %v", MinSegmentDuration)
		}
		o.SegmentSize = v
		return nil
	}
}

// WithSilenceSize sets the silence duration that triggers a segment boundary.
// Only takes effect when silence detection is enabled. Minimum is 100ms.
func WithSilenceSize(v time.Duration) Opt {
	return func(o *opts) error {
		if v < MinSegmentDuration {
			return media.ErrBadParameter.Withf("silence duration is too short, must be at least %v", MinSegmentDuration)
		}
		o.SilenceSize = v
		return nil
	}
}

// WithSilenceThreshold enables silence detection with a custom threshold.
// The threshold is the RMS energy level (0.0-1.0) below which audio is
// considered silence. A typical value is 0.01 (1%).
func WithSilenceThreshold(threshold float64) Opt {
	return func(o *opts) error {
		if threshold <= 0 || threshold > 1 {
			return media.ErrBadParameter.Withf("silence threshold must be between 0 and 1, got %v", threshold)
		}
		o.SilenceThreshold = threshold
		if o.SilenceSize == 0 {
			o.SilenceSize = DefaultSilenceDuration
		}
		return nil
	}
}

// WithDefaultSilence enables silence detection with default threshold and duration.
// Uses threshold of 0.01 (1% RMS) and silence duration of 500ms.
func WithDefaultSilence() Opt {
	return func(o *opts) error {
		o.SilenceThreshold = DefaultSilenceThreshold
		o.SilenceSize = DefaultSilenceDuration
		return nil
	}
}
