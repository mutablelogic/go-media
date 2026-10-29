package segmenter

import (
	"math"
	"time"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	libfvad "github.com/mutablelogic/go-media/sys/libfvad"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// analyser measures frames of samples, and returns the class and stats of
// segments. It keeps the measurements of the current segment, and of earlier
// segments within ClassifyWindow, so a segment can be classified from the
// samples before it as well, without waiting for any after it
type analyser struct {
	threshold float64 // RMS energy below which a segment is silence
	features  *features
	frames    []record // measurements of earlier segments, then the current segment
	start     int      // index of the first frame of the current segment
	window    int      // number of frames to classify a segment with

	// Voice activity detection, if vad is not nil, of the samples at
	// vadSampleRate. Each decision is for VADFrameDuration of samples, and
	// is used for the frames of samples at sampleRate whose middle it covers
	vad        *libfvad.Context
	sampleRate int
	measured   int64  // number of samples measured at sampleRate
	decisions  []bool // decisions which haven't been used yet
	first      int64  // index of the first decision in decisions
	last       bool   // the last decision which was used, if hasLast is true
	hasLast    bool
}

// measurement is the result of measuring a frame of samples
type measurement struct {
	rms    float64 // RMS energy of the frame (0.0-1.0)
	vad    bool    // whether voice activity detection was used
	speech bool    // whether voice activity detection found speech
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// VADFrameDuration is the duration of each frame of samples which is
// analysed for voice activity
const VADFrameDuration = 20 * time.Millisecond

// vadSampleRate is the sample rate for voice activity detection, which is
// the rate libfvad processes all samples at
const vadSampleRate = 8000

// ClassifyWindow is the minimum duration of samples a segment is classified
// with. A shorter segment is classified with the samples before it as well,
// so it isn't output any later.
const ClassifyWindow = 6 * time.Second

// Thresholds to classify a segment which isn't silence. Speech has pauses
// between syllables and words, so many frames with low energy, a
// zero-crossing rate which varies as voiced and unvoiced sounds alternate,
// and a spectrum which changes quickly. When background sound fills the
// pauses, a zero-crossing rate which varies a lot is also speech. Music has a
// spectrum which changes slowly, and voice activity detection reports it as
// speech. The rest is other sound, such as noise. These were chosen from the
// MUSAN corpus (make test-musan), a music library and a news podcast
const (
	minSpeechLowEnergy     = 0.25 // Minimum fraction of frames with low energy for speech
	minSpeechZCRVar        = 0.4  // Minimum zero-crossing rate variation for speech
	minSpeechFlux          = 0.25 // Minimum spectral flux for speech
	minBackgroundLowEnergy = 0.15 // Minimum fraction of frames with low energy for speech with background sound
	minBackgroundZCRVar    = 0.8  // Minimum zero-crossing rate variation for speech with background sound
	maxMusicFlux           = 0.32 // Maximum spectral flux for music
	minMusicSpeech         = 0.8  // Minimum fraction of frames with speech for music
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// newAnalyser returns an analyser, which uses voice activity detection if
// the mode is a voice activity detection mode. A segment is silence when its
// RMS energy is below the threshold, or DefaultSilenceThreshold if zero
func newAnalyser(sampleRate int, m Mode, threshold float64) (*analyser, error) {
	a := &analyser{
		features:   newFeatures(sampleRate),
		window:     int(ClassifyWindow / featureFrameDuration),
		sampleRate: sampleRate,
	}
	if a.threshold = threshold; a.threshold == 0 {
		a.threshold = DefaultSilenceThreshold
	}
	mode, vad := m.vad()
	if !vad {
		return a, nil
	}
	if a.vad = libfvad.NewContext(); a.vad == nil {
		return nil, gomedia.ErrInternalError.With("failed to create VAD context")
	}
	if err := a.vad.SetSampleRate(vadSampleRate); err != nil {
		a.Close()
		return nil, err
	}
	if err := a.vad.SetMode(mode); err != nil {
		a.Close()
		return nil, err
	}
	return a, nil
}

// Close releases the VAD context
func (a *analyser) Close() {
	if a.vad != nil {
		a.vad.Free()
		a.vad = nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// vadFrameSize returns the number of samples at vadSampleRate which
// writeVAD is called with
func vadFrameSize() int {
	return int(VADFrameDuration.Seconds() * vadSampleRate)
}

// writeVAD detects voice activity in a frame of vadFrameSize samples at
// vadSampleRate. A shorter frame, at the end of the samples, is ignored
func (a *analyser) writeVAD(samples []int16) error {
	if a.vad == nil || len(samples) != vadFrameSize() {
		return nil
	}
	speech, err := a.vad.Process(samples)
	if err != nil {
		return err
	}
	a.decisions = append(a.decisions, speech)
	return nil
}

// measure returns the RMS energy of a frame of samples and, with voice
// activity detection, whether it has speech
func (a *analyser) measure(samples []float32) (measurement, error) {
	var result measurement
	var sum float64
	for _, sample := range samples {
		sum += float64(sample) * float64(sample)
	}
	if len(samples) > 0 {
		result.rms = math.Sqrt(sum / float64(len(samples)))
	}
	if a.vad == nil {
		return result, nil
	}

	// Use the decision for the middle of the frame, or the last decision if
	// it's not been made yet, as resampling delays the samples a little
	middle := a.measured + int64(len(samples)/2)
	a.measured += int64(len(samples))
	k := middle * vadSampleRate / int64(a.sampleRate) / int64(vadFrameSize())
	if i := k - a.first; i < int64(len(a.decisions)) {
		a.last, a.hasLast = a.decisions[i], true
		a.decisions, a.first = a.decisions[i:], k
	} else if len(a.decisions) > 0 {
		a.last, a.hasLast = a.decisions[len(a.decisions)-1], true
	}
	if a.hasLast {
		result.vad, result.speech = true, a.last
	}
	return result, nil
}

// add measures a frame of samples, with its measurement, which belongs to the
// current segment
func (a *analyser) add(samples []float32, m measurement) {
	first := len(a.frames)
	n := a.features.frameSize()
	for ; len(samples) >= n; samples = samples[n:] {
		a.frames = append(a.frames, a.features.measure(samples[:n]))
	}

	// Samples which don't fill a frame are added to the last one
	if len(samples) > 0 {
		f := a.features.measure(samples)
		if len(a.frames) > first {
			last := &a.frames[len(a.frames)-1]
			last.n += f.n
			last.sum += f.sum
			last.peak = max(last.peak, f.peak)
		} else {
			a.frames = append(a.frames, f)
		}
	}

	// A frame checked for voice activity is VADFrameDuration long, the same
	// as a feature frame
	if m.vad && len(a.frames) > first {
		a.frames[first].vad, a.frames[first].speech = true, m.speech
	}
}

// count returns the number of frames in the current segment
func (a *analyser) count() int {
	return len(a.frames) - a.start
}

// next returns the class and stats of the first n frames of the current
// segment, and starts the next segment after them. RMS and Peak are measured
// over the segment, and the rest over at least ClassifyWindow before its end.
// A segment which was detected as silence is always ClassSilence
func (a *analyser) next(n int, silent bool) (Class, Stats) {
	class, stats := a.classify(a.start+n, silent)

	// Keep the frames within the window before the next segment
	a.start += n
	if drop := a.start - a.window; drop > 0 {
		a.frames = a.frames[:copy(a.frames, a.frames[drop:])]
		a.start -= drop
	}
	return class, stats
}

// classify returns the class and stats of the current segment, up to the
// frame before end
func (a *analyser) classify(end int, silent bool) (Class, Stats) {
	var stats Stats
	var sum float64
	var samples int
	for _, f := range a.frames[a.start:end] {
		sum += f.sum
		samples += f.n
		stats.Peak = max(stats.Peak, f.peak)
	}
	if samples > 0 {
		stats.RMS = math.Sqrt(sum / float64(samples))
	}
	setFeatures(&stats, a.frames[max(0, min(a.start, end-a.window)):end])

	var class Class
	switch {
	case silent || stats.RMS < a.threshold:
		class = ClassSilence
	case a.vad == nil:
		class = ClassUnknown
	case stats.isSpeech():
		class = ClassSpeech
	case stats.Flux < maxMusicFlux && stats.Speech >= minMusicSpeech:
		class = ClassMusic
	default:
		class = ClassOther
	}
	return class, stats
}

// isSpeech returns true if the stats are those of speech, either with pauses,
// or with background sound which fills them
func (s Stats) isSpeech() bool {
	if s.Flux < minSpeechFlux {
		return false
	}
	pauses := s.LowEnergy >= minSpeechLowEnergy && s.ZCRVariation >= minSpeechZCRVar
	background := s.LowEnergy >= minBackgroundLowEnergy && s.ZCRVariation >= minBackgroundZCRVar
	return pauses || background
}
