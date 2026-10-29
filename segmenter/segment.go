package segmenter

import (
	"fmt"
	"math"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Segment is a segment of mono audio samples, passed to a SegmentFn. It's
// only valid during the call, as its samples are reused for the next one.
type Segment struct {
	Start    time.Duration // Timestamp of the first sample
	Duration time.Duration // Duration of the samples
	Samples  []float32     // Samples, in the range [-1.0, 1.0]
	Class    Class         // Kind of audio
	Stats    Stats         // Measurements of the samples
}

// Class is the kind of audio in a segment
type Class int

// Stats are measurements of a segment's samples. Except for RMS and Peak,
// they're measured over 20ms frames
type Stats struct {
	Speech        float64 // Fraction of frames with speech (0.0-1.0), with a ModeVAD mode
	RMS           float64 // RMS energy of the samples (0.0-1.0)
	Peak          float64 // Largest absolute sample value (0.0-1.0)
	Flatness      float64 // Mean spectral flatness, from tonal (0.0) to noise (1.0)
	ZCR           float64 // Mean zero-crossing rate, in crossings per sample
	ZCRVariation  float64 // Standard deviation of the zero-crossing rate, divided by its mean
	LowEnergy     float64 // Fraction of frames with less than half the mean RMS energy
	Modulation    float64 // Fraction of the energy's modulation at the syllable rate of speech (2-8Hz)
	Centroid      float64 // Mean spectral centroid, in Hz
	Flux          float64 // Mean spectral flux, how much the spectrum changes between frames (0.0-2.0)
	FluxVariation float64 // Standard deviation of the spectral flux, divided by its mean
	Rhythm        float64 // Strength of a steady beat of 60-180 BPM in the energy onsets (0.0-1.0)
	Harmonicity   float64 // Mean harmonicity, from noise (0.0) to a periodic sound with a pitch of 100-1000Hz (1.0)
	Voiced        float64 // Fraction of frames with a harmonicity of at least 0.5
}

// SegmentFn is called for each segment. Return nil to continue, io.EOF to
// stop early, or any other error to abort.
type SegmentFn func(*Segment) error

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// Classes of audio. A segment is silence when its RMS energy is below the
// silence threshold, or DefaultSilenceThreshold when there's none.
// Otherwise, it's classified further only with a ModeVAD mode.
const (
	ClassUnknown Class = iota // Not silence, and not classified further
	ClassSilence              // Silence
	ClassSpeech               // Mostly speech
	ClassMusic                // Tonal sound which isn't speech, such as music
	ClassOther                // Other sound, such as noise
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Int16 returns a copy of the samples, in the range [-32767, 32767].
func (s *Segment) Int16() []int16 {
	result := make([]int16, len(s.Samples))
	for i, sample := range s.Samples {
		result[i] = toInt16(sample)
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (c Class) String() string {
	switch c {
	case ClassUnknown:
		return "ClassUnknown"
	case ClassSilence:
		return "ClassSilence"
	case ClassSpeech:
		return "ClassSpeech"
	case ClassMusic:
		return "ClassMusic"
	case ClassOther:
		return "ClassOther"
	default:
		return fmt.Sprintf("Class(%d)", int(c))
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func toInt16(sample float32) int16 {
	return int16(math.Round(float64(max(-1, min(1, sample))) * math.MaxInt16))
}
