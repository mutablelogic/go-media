package segmenter

import (
	"math"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// joiner joins short segments, which are cut at least every pieceDuration
// and at silences, into one segment for each part of the audio between the
// boundaries. Each cut is at the edge between two short segments nearest a
// boundary, preferring an edge next to silence within boundaryTolerance
type joiner struct {
	fn         SegmentFn
	boundaries []time.Duration // boundaries which haven't been cut at yet
	pieces     []Segment       // short segments which haven't been output yet
	segment    Segment         // the segment which is output
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// pieceDuration is the maximum duration of the short segments which are
	// joined
	pieceDuration = time.Second

	// boundaryTolerance is how far from a boundary a cut at a silence can be
	boundaryTolerance = time.Second
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newJoiner(boundaries []time.Duration, fn SegmentFn) *joiner {
	return &joiner{fn: fn, boundaries: boundaries}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// write adds a short segment, and outputs segments for the boundaries which
// it's now possible to choose a cut for
func (j *joiner) write(s *Segment) error {
	j.pieces = append(j.pieces, Segment{
		Start:    s.Start,
		Duration: s.Duration,
		Samples:  append([]float32(nil), s.Samples...),
		Class:    s.Class,
		Stats:    s.Stats,
	})
	for len(j.boundaries) > 0 && j.end() >= j.boundaries[0]+boundaryTolerance {
		if err := j.cut(j.boundaries[0]); err != nil {
			return err
		}
	}
	return nil
}

// close outputs the segments which remain
func (j *joiner) close() error {
	for len(j.boundaries) > 0 && len(j.pieces) > 0 && j.boundaries[0] < j.end() {
		if err := j.cut(j.boundaries[0]); err != nil {
			return err
		}
	}
	if len(j.pieces) > 0 {
		return j.output(len(j.pieces))
	}
	return nil
}

// end returns the end of the short segments which haven't been output
func (j *joiner) end() time.Duration {
	if len(j.pieces) == 0 {
		return 0
	}
	last := j.pieces[len(j.pieces)-1]
	return last.Start + last.Duration
}

// cut outputs the short segments up to the edge nearest a boundary, and
// removes the boundaries before the cut
func (j *joiner) cut(boundary time.Duration) error {
	best, silent := -1, false
	var distance time.Duration
	for i, p := range j.pieces {
		edge := p.Start + p.Duration
		d := (edge - boundary).Abs()
		s := p.Class == ClassSilence || (i+1 < len(j.pieces) && j.pieces[i+1].Class == ClassSilence)
		s = s && d <= boundaryTolerance
		if best < 0 || (s && !silent) || (s == silent && d < distance) {
			best, silent, distance = i, s, d
		}
	}

	edge := j.pieces[best].Start + j.pieces[best].Duration
	for len(j.boundaries) > 0 && (j.boundaries[0] == boundary || j.boundaries[0] <= edge) {
		j.boundaries = j.boundaries[1:]
	}
	return j.output(best + 1)
}

// output joins the first n short segments into a segment, outputs it, and
// removes them
func (j *joiner) output(n int) error {
	pieces := j.pieces[:n]
	j.segment.Start = pieces[0].Start
	j.segment.Duration = pieces[n-1].Start + pieces[n-1].Duration - pieces[0].Start
	j.segment.Samples = j.segment.Samples[:0]
	for _, p := range pieces {
		j.segment.Samples = append(j.segment.Samples, p.Samples...)
	}
	j.segment.Class, j.segment.Stats = join(pieces)
	err := j.fn(&j.segment)
	j.pieces = j.pieces[:copy(j.pieces, j.pieces[n:])]
	return err
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

// join returns the class and stats of segments which are joined together.
// The class is the one with the longest duration except silence, or silence
// if they all are. RMS is the RMS of all the samples, Peak is the largest,
// and the rest are weighted by duration
func join(segments []Segment) (Class, Stats) {
	var result Stats
	var total, energy float64
	durations := make(map[Class]time.Duration)
	for _, s := range segments {
		d := s.Duration.Seconds()
		total += d
		energy += s.Stats.RMS * s.Stats.RMS * d
		result.Peak = max(result.Peak, s.Stats.Peak)
		result.Speech += s.Stats.Speech * d
		result.Flatness += s.Stats.Flatness * d
		result.ZCR += s.Stats.ZCR * d
		result.ZCRVariation += s.Stats.ZCRVariation * d
		result.LowEnergy += s.Stats.LowEnergy * d
		result.Modulation += s.Stats.Modulation * d
		result.Centroid += s.Stats.Centroid * d
		result.Flux += s.Stats.Flux * d
		result.FluxVariation += s.Stats.FluxVariation * d
		result.Rhythm += s.Stats.Rhythm * d
		result.Harmonicity += s.Stats.Harmonicity * d
		result.Voiced += s.Stats.Voiced * d
		durations[s.Class] += s.Duration
	}
	if total > 0 {
		result.RMS = math.Sqrt(energy / total)
		for _, v := range []*float64{
			&result.Speech, &result.Flatness, &result.ZCR, &result.ZCRVariation, &result.LowEnergy,
			&result.Modulation, &result.Centroid, &result.Flux, &result.FluxVariation, &result.Rhythm,
			&result.Harmonicity, &result.Voiced,
		} {
			*v /= total
		}
	}

	class := ClassSilence
	for _, c := range []Class{ClassSpeech, ClassMusic, ClassOther, ClassUnknown} {
		if durations[c] > 0 && (class == ClassSilence || durations[c] > durations[class]) {
			class = c
		}
	}
	return class, result
}
