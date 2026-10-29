package thumbnailer

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Sample is a frame sampled from a video, and its stats
type Sample struct {
	Time     time.Duration // Time of the frame
	Keyframe bool          // Whether it was sampled as a keyframe
	Change   float64       // How different it is from the sample before (0.0-1.0)
	Stats    Stats         // How good a thumbnail it makes
	hist     histogram
}

// Scene is a part of a video between scene changes. As frames are sampled,
// its start is the time of its first sample, which is up to the time between
// samples after the scene changed
type Scene struct {
	Start      time.Duration // Time of its first sample
	Duration   time.Duration // Until the next scene starts, or the end
	Transition Transition    // How it starts
	First      int           // Index of its first sample
	Count      int           // Number of samples in it, which aren't blank
	Best       int           // Index of its best sample, or -1 if they're all blank
}

// Transition is how a scene starts
type Transition int

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	TransitionStart Transition = iota // The start of the video
	TransitionCut                     // A cut from the scene before
	TransitionFade                    // A fade, through blank frames, from the scene before
)

// Thresholds for a scene change, between samples
const (
	minCut      = 0.2  // Minimum change for a cut
	cutRatio    = 4    // A cut is at least this many times the median change
	flashReturn = 0.5  // A change is a flash when the sample after it is this close to the sample before
	trimEnds    = 0.03 // Fraction of the video at each end where thumbnails are avoided
	minDiverse  = 0.15 // Minimum change between thumbnails
)

///////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

// detect sets how different each sample is from the one before, and returns
// the scenes, which end at the duration of the video, or after the last
// sample if it's unknown. A scene changes where a sample is much more
// different from the one before than is usual for the video, unless the
// sample after it is like the one before, which is a flash, and the change
// back from a flash isn't a cut either. Blank samples,
// such as a fade through black, end a scene, and the next sample which
// isn't blank starts the next one
func detect(samples []Sample, duration time.Duration) []Scene {
	if len(samples) == 0 {
		return nil
	}

	// How different each sample is from the one before, and what's usual
	changes := make([]float64, 0, len(samples))
	for i := 1; i < len(samples); i++ {
		samples[i].Change = distance(samples[i-1].hist, samples[i].hist)
		if !samples[i].Stats.Blank && !samples[i-1].Stats.Blank {
			changes = append(changes, samples[i].Change)
		}
	}
	threshold := minCut
	if len(changes) > 0 {
		threshold = max(minCut, cutRatio*median(changes))
	}

	var scenes []Scene
	start := func(i int, transition Transition) {
		scenes = append(scenes, Scene{Start: samples[i].Time, Transition: transition, First: i, Best: -1})
	}
	start(0, TransitionStart)
	// A scene starts with a sample which isn't blank, so blank samples only
	// end a scene once a sample which isn't blank has been seen, rather than
	// at the start of the video
	seen := !samples[0].Stats.Blank
	faded, flash := false, false
	for i := 1; i < len(samples); i++ {
		switch {
		case samples[i].Stats.Blank:
			faded = faded || seen
			flash = false
		case faded:
			start(i, TransitionFade)
			faded, flash = false, false
		case flash:
			flash = false
		case samples[i].Change >= threshold:
			if flash = isFlash(samples, i); !flash {
				start(i, TransitionCut)
			}
		}
		seen = seen || !samples[i].Stats.Blank
	}

	// The duration, number of samples and best sample of each scene
	end := duration
	if last := samples[len(samples)-1].Time; end <= last {
		end = last
	}
	for i := range scenes {
		next := len(samples)
		if i+1 < len(scenes) {
			next = scenes[i+1].First
			scenes[i].Duration = scenes[i+1].Start - scenes[i].Start
		} else {
			scenes[i].Duration = end - scenes[i].Start
		}
		for j := scenes[i].First; j < next; j++ {
			if !samples[j].Stats.Blank {
				scenes[i].Count++
			}
			if samples[j].Stats.Score > 0 && (scenes[i].Best < 0 || samples[j].Stats.Score > samples[scenes[i].Best].Stats.Score) {
				scenes[i].Best = j
			}
		}
	}
	return scenes
}

// isFlash returns true when a sample is a change from the one before, but
// the sample after it is like the one before, such as a camera flash
func isFlash(samples []Sample, i int) bool {
	if i+1 >= len(samples) || samples[i+1].Stats.Blank {
		return false
	}
	return distance(samples[i-1].hist, samples[i+1].hist) < flashReturn*samples[i].Change
}

// choose returns the indexes of up to n samples for thumbnails, in time
// order: the best sample of each scene, avoiding the ends of the video if
// there are enough scenes, ranked by score and by how long the scene is,
// and each different from those chosen before it. With fewer scenes than n,
// more samples are chosen from the longest scenes
func choose(samples []Sample, scenes []Scene, n int, duration time.Duration) []int {
	if n <= 0 || len(samples) == 0 {
		return nil
	}

	// The scenes with a best sample, away from the ends if there are enough
	var candidates []Scene
	for _, s := range scenes {
		if s.Best >= 0 {
			candidates = append(candidates, s)
		}
	}
	if duration > 0 {
		lo, hi := time.Duration(float64(duration)*trimEnds), time.Duration(float64(duration)*(1-trimEnds))
		var inside []Scene
		for _, s := range candidates {
			if t := samples[s.Best].Time; t >= lo && t <= hi {
				inside = append(inside, s)
			}
		}
		if len(inside) >= n {
			candidates = inside
		}
	}

	// Rank by score, and by how long the scene is compared to the average
	var mean float64
	for _, s := range candidates {
		mean += s.Duration.Seconds()
	}
	if len(candidates) > 0 {
		mean /= float64(len(candidates))
	}
	rank := func(s Scene) float64 {
		length := 1.0
		if mean > 0 {
			length = 0.5 + 0.5*min(1, s.Duration.Seconds()/mean)
		}
		return samples[s.Best].Stats.Score * length
	}
	sort.SliceStable(candidates, func(i, j int) bool { return rank(candidates[i]) > rank(candidates[j]) })

	// Choose the best, each different from those chosen before, then the rest
	var result []int
	chosen := func(i int) bool { return slices.Contains(result, i) }
	diverse := func(i int) bool {
		for _, j := range result {
			if distance(samples[i].hist, samples[j].hist) < minDiverse {
				return false
			}
		}
		return true
	}
	for _, s := range candidates {
		if len(result) < n && diverse(s.Best) {
			result = append(result, s.Best)
		}
	}
	for _, s := range candidates {
		if len(result) < n && !chosen(s.Best) {
			result = append(result, s.Best)
		}
	}

	// With too few scenes, more samples from the longest scenes, as far
	// from those chosen as possible
	for len(result) < n {
		best, bestTime := -1, time.Duration(-1)
		for i, sample := range samples {
			if chosen(i) || sample.Stats.Score == 0 {
				continue
			}
			nearest := time.Duration(math.MaxInt64)
			for _, j := range result {
				nearest = min(nearest, (sample.Time - samples[j].Time).Abs())
			}
			if nearest > bestTime {
				best, bestTime = i, nearest
			}
		}
		if best < 0 {
			break
		}
		result = append(result, best)
	}

	sort.Ints(result)
	return result
}

func median(values []float64) float64 {
	sorted := slices.Clone(values)
	slices.Sort(sorted)
	return sorted[len(sorted)/2]
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t Transition) String() string {
	switch t {
	case TransitionStart:
		return "start"
	case TransitionCut:
		return "cut"
	case TransitionFade:
		return "fade"
	default:
		return fmt.Sprintf("Transition(%d)", int(t))
	}
}
