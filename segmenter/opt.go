package segmenter

import (
	"fmt"
	"time"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	reader "github.com/mutablelogic/go-media/reader"
	libfvad "github.com/mutablelogic/go-media/sys/libfvad"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Opt is a function which applies options to a Segmenter
type Opt func(*opts) error

// Mode sets how silence is detected, to break segments at silence boundaries
type Mode int

type opts struct {
	SegmentSize      time.Duration   // Segment size, zero means no fixed segmenting
	Boundaries       []time.Duration // Times to cut segments at, approximately
	Mode             Mode            // Silence detection mode
	SilenceSize      time.Duration   // Size of silence to consider a segment boundary
	SilenceThreshold float64         // Silence threshold (RMS energy 0.0-1.0), zero means none
	stream           int             // Audio stream to decode, or -1 for the first one
	readerOpts       []reader.Opt    // Reader options
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DefaultSilenceThreshold = 0.01                   // Default silence threshold (RMS)
	DefaultSilenceDuration  = time.Millisecond * 500 // Default silence duration
	MinSegmentDuration      = time.Millisecond * 100 // Minimum segment/silence duration
)

// Silence detection modes. The voice activity detection modes are in
// increasing order of aggressiveness: a more aggressive mode is more likely
// to be right when it reports speech, but misses more of it.
const (
	ModeNone              Mode = iota // No silence detection
	ModeSilence                       // Silence is where the RMS energy is below the threshold
	ModeVADQuality                    // Silence is where there's no speech, or the RMS energy is below the threshold
	ModeVADLowBitrate                 // See ModeVADQuality
	ModeVADAggressive                 // See ModeVADQuality
	ModeVADVeryAggressive             // See ModeVADQuality
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func applyOpts(opt ...Opt) (*opts, error) {
	o := opts{stream: -1}
	for _, fn := range opt {
		if err := fn(&o); err != nil {
			return nil, err
		}
	}
	return &o, nil
}

///////////////////////////////////////////////////////////////////////////////
// OPTIONS

// WithReaderOpt passes a reader.Opt to the reader, such as reader.WithInput
// to set the input format of raw audio.
func WithReaderOpt(opt reader.Opt) Opt {
	return func(o *opts) error {
		o.readerOpts = append(o.readerOpts, opt)
		return nil
	}
}

// WithStream sets the index of the audio stream to decode. Without it, the
// audio stream with the lowest index is decoded.
func WithStream(stream int) Opt {
	return func(o *opts) error {
		if stream < 0 {
			return gomedia.ErrBadParameter.Withf("invalid stream %d", stream)
		}
		o.stream = stream
		return nil
	}
}

// WithSegments sets how segments are cut. With one duration, it's the target
// segment duration, so segments are output when they reach approximately
// this duration (minimum 100ms). With more than one, they're the times from
// the start to cut segments at, in increasing order, not including the start
// or the end. Each cut is at a silence near the time if there is one, so one
// segment is output for each part of the audio between the times.
func WithSegments(v ...time.Duration) Opt {
	return func(o *opts) error {
		switch {
		case len(v) == 0:
			return gomedia.ErrBadParameter.With("no segment duration or boundaries")
		case len(v) == 1:
			if v[0] < MinSegmentDuration {
				return gomedia.ErrBadParameter.Withf("segment duration is too short, must be at least %v", MinSegmentDuration)
			}
			o.SegmentSize, o.Boundaries = v[0], nil
		default:
			for i, b := range v {
				if b <= 0 || (i > 0 && b <= v[i-1]) {
					return gomedia.ErrBadParameter.With("segment boundaries must be positive and in increasing order")
				}
			}
			o.SegmentSize, o.Boundaries = 0, v
		}
		return nil
	}
}

// WithMode sets how silence is detected, to break segments where silence
// has lasted at least the silence duration (minimum 100ms, or zero for
// DefaultSilenceDuration). The threshold is the RMS energy level (0.0-1.0)
// below which audio is silence. With ModeSilence, zero is
// DefaultSilenceThreshold. With the voice activity detection modes, which
// also classify segments as speech, music or other sound, zero is no
// threshold.
func WithMode(mode Mode, silence time.Duration, threshold float64) Opt {
	return func(o *opts) error {
		if mode < ModeNone || mode > ModeVADVeryAggressive {
			return gomedia.ErrBadParameter.Withf("invalid mode %v", mode)
		}
		if silence != 0 && silence < MinSegmentDuration {
			return gomedia.ErrBadParameter.Withf("silence duration is too short, must be at least %v", MinSegmentDuration)
		}
		if threshold < 0 || threshold > 1 {
			return gomedia.ErrBadParameter.Withf("silence threshold must be between 0 and 1, got %v", threshold)
		}
		if silence == 0 {
			silence = DefaultSilenceDuration
		}
		if threshold == 0 && mode == ModeSilence {
			threshold = DefaultSilenceThreshold
		}
		if mode == ModeNone {
			silence, threshold = 0, 0
		}
		o.Mode, o.SilenceSize, o.SilenceThreshold = mode, silence, threshold
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// vad returns the libfvad mode for a voice activity detection mode, or false
func (m Mode) vad() (libfvad.Mode, bool) {
	switch m {
	case ModeVADQuality:
		return libfvad.MODE_QUALITY, true
	case ModeVADLowBitrate:
		return libfvad.MODE_LOW_BITRATE, true
	case ModeVADAggressive:
		return libfvad.MODE_AGGRESSIVE, true
	case ModeVADVeryAggressive:
		return libfvad.MODE_VERY_AGGRESSIVE, true
	default:
		return 0, false
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m Mode) String() string {
	switch m {
	case ModeNone:
		return "ModeNone"
	case ModeSilence:
		return "ModeSilence"
	case ModeVADQuality:
		return "ModeVADQuality"
	case ModeVADLowBitrate:
		return "ModeVADLowBitrate"
	case ModeVADAggressive:
		return "ModeVADAggressive"
	case ModeVADVeryAggressive:
		return "ModeVADVeryAggressive"
	default:
		return fmt.Sprintf("Mode(%d)", int(m))
	}
}
