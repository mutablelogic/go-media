package segmenter

import (
	"math"
	"math/bits"
	"math/cmplx"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// features measures each frame of samples, to tell speech, music and noise
// apart
type features struct {
	sampleRate int
	window     []float64    // Hann window for a frame
	buf        []complex128 // frame samples, zero-padded to a power of two, for the FFT
	power      []float64    // power spectrum of the frame, without the DC bin
	prev       []float64    // normalised magnitude spectrum of the previous frame
	cur        []float64    // normalised magnitude spectrum of the frame
	hasPrev    bool         // whether prev is set
}

// record is the measurement of a frame of samples
type record struct {
	n    int     // number of samples
	sum  float64 // sum of squared samples
	peak float64 // largest absolute sample value

	// Voice activity detection, if vad is true
	vad, speech bool

	// Features, if features is true, which isn't the case for a frame at the
	// end of the samples which is shorter than featureFrameDuration
	features    bool
	energy      float64 // RMS energy
	zcr         float64 // zero-crossing rate
	flatness    float64 // spectral flatness
	centroid    float64 // spectral centroid, in Hz
	flux        float64 // spectral flux, if hasFlux is true
	hasFlux     bool
	harmonicity float64 // autocorrelation at the strongest pitch
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// featureFrameDuration is the duration of each frame of samples which is
// measured for features
const featureFrameDuration = 20 * time.Millisecond

// Range of envelope modulation frequencies which is the syllable rate of
// speech, in Hz
const (
	minSyllableRate = 2
	maxSyllableRate = 8
)

// Range of tempos which are measured for rhythm, in beats per minute
const (
	minTempo = 60
	maxTempo = 180
)

// Range of pitches which are measured for harmonicity, in Hz, and the
// harmonicity above which a frame is voiced
const (
	minPitch = 100
	maxPitch = 1000
	voiced   = 0.5
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newFeatures(sampleRate int) *features {
	n := max(1, int(featureFrameDuration.Seconds()*float64(sampleRate)))
	size := 1 << bits.Len(uint(n-1))
	f := &features{
		sampleRate: sampleRate,
		window:     make([]float64, n),
		buf:        make([]complex128, size),
		power:      make([]float64, size/2),
		prev:       make([]float64, size/2),
		cur:        make([]float64, size/2),
	}
	for i := range f.window {
		f.window[i] = 0.5 - 0.5*math.Cos(2*math.Pi*float64(i)/float64(n))
	}
	return f
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// frameSize returns the number of samples in a frame
func (f *features) frameSize() int {
	return len(f.window)
}

// measure returns the measurement of a frame of samples, with features if
// there are frameSize samples
func (f *features) measure(samples []float32) record {
	result := record{n: len(samples)}
	var crossings int
	for i, sample := range samples {
		v := float64(sample)
		result.sum += v * v
		result.peak = max(result.peak, math.Abs(v))
		if i > 0 && (sample >= 0) != (samples[i-1] >= 0) {
			crossings++
		}
	}
	if len(samples) != f.frameSize() {
		return result
	}

	result.features = true
	result.energy = math.Sqrt(result.sum / float64(len(samples)))
	result.zcr = float64(crossings) / float64(len(samples)-1)
	f.spectrum(samples)
	result.flatness = f.spectralFlatness()
	result.centroid = f.spectralCentroid()
	result.flux, result.hasFlux = f.spectralFlux()
	result.harmonicity = f.harmonic(samples)
	return result
}

// spectrum sets the power spectrum of a frame of samples
func (f *features) spectrum(samples []float32) {
	for i := range f.buf {
		if i < len(samples) {
			f.buf[i] = complex(float64(samples[i])*f.window[i], 0)
		} else {
			f.buf[i] = 0
		}
	}
	fft(f.buf)

	// Skip the DC bin
	for i := range f.power {
		v := f.buf[i+1]
		f.power[i] = real(v)*real(v) + imag(v)*imag(v)
	}
}

// spectralFlatness returns the geometric mean of the power spectrum divided
// by its arithmetic mean, which is close to zero for a tone and close to one
// for white noise
func (f *features) spectralFlatness() float64 {
	// Add a small amount so silence has no zero bins
	const epsilon = 1e-12
	var logSum, sum float64
	for _, p := range f.power {
		logSum += math.Log(p + epsilon)
		sum += p + epsilon
	}
	n := float64(len(f.power))
	return math.Exp(logSum/n) / (sum / n)
}

// spectralCentroid returns the centre of mass of the power spectrum, in Hz
func (f *features) spectralCentroid() float64 {
	var weighted, sum float64
	for i, p := range f.power {
		weighted += p * f.frequency(i)
		sum += p
	}
	if sum == 0 {
		return 0
	}
	return weighted / sum
}

// spectralFlux returns the distance between the normalised magnitude
// spectrum of the frame and the previous one, from zero when it doesn't
// change to two when nothing is in common, or false if there's no previous
// frame or the frame is silent
func (f *features) spectralFlux() (float64, bool) {
	var sum float64
	for i, p := range f.power {
		f.cur[i] = math.Sqrt(p)
		sum += f.cur[i] * f.cur[i]
	}
	if sum == 0 {
		f.hasPrev = false
		return 0, false
	}
	norm := math.Sqrt(sum)
	var flux float64
	for i := range f.cur {
		f.cur[i] /= norm
		d := f.cur[i] - f.prev[i]
		flux += d * d
	}
	hasPrev := f.hasPrev
	f.prev, f.cur, f.hasPrev = f.cur, f.prev, true
	return flux, hasPrev
}

// frequency returns the frequency of a bin of the power spectrum, in Hz
func (f *features) frequency(bin int) float64 {
	return float64(bin+1) * float64(f.sampleRate) / float64(len(f.buf))
}

// harmonic returns the largest normalised autocorrelation of a frame of
// samples at a lag which is a pitch between minPitch and maxPitch, which is
// close to one for a periodic sound and close to zero for noise
func (f *features) harmonic(samples []float32) float64 {
	minLag := max(1, f.sampleRate/maxPitch)
	maxLag := min(len(samples)/2, f.sampleRate/minPitch)
	var result float64
	for lag := minLag; lag <= maxLag; lag++ {
		var xy, xx, yy float64
		for i := 0; i+lag < len(samples); i++ {
			x, y := float64(samples[i]), float64(samples[i+lag])
			xy += x * y
			xx += x * x
			yy += y * y
		}
		if xx > 0 && yy > 0 {
			result = max(result, xy/math.Sqrt(xx*yy))
		}
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

// setFeatures sets the speech and feature stats of frames
func setFeatures(stats *Stats, frames []record) {
	var energy, zcr, flatness, centroid, flux, harmonicity []float64
	var vad, speech int
	for _, f := range frames {
		if f.vad {
			vad++
			if f.speech {
				speech++
			}
		}
		if !f.features {
			continue
		}
		energy = append(energy, f.energy)
		zcr = append(zcr, f.zcr)
		flatness = append(flatness, f.flatness)
		centroid = append(centroid, f.centroid)
		harmonicity = append(harmonicity, f.harmonicity)
		if f.hasFlux {
			flux = append(flux, f.flux)
		}
	}
	if vad > 0 {
		stats.Speech = float64(speech) / float64(vad)
	}

	stats.Flatness = mean(flatness)
	stats.ZCR = mean(zcr)
	if stats.ZCR > 0 {
		stats.ZCRVariation = stddev(zcr) / stats.ZCR
	}

	// Fraction of frames with less than half the mean energy
	if threshold := mean(energy) / 2; len(energy) > 0 {
		var low int
		for _, e := range energy {
			if e < threshold {
				low++
			}
		}
		stats.LowEnergy = float64(low) / float64(len(energy))
	}

	rate := 1 / featureFrameDuration.Seconds()
	stats.Modulation = modulation(energy, rate)
	stats.Rhythm = rhythm(onsets(energy), rate)
	stats.Centroid = mean(centroid)
	stats.Flux = mean(flux)
	if stats.Flux > 0 {
		stats.FluxVariation = stddev(flux) / stats.Flux
	}
	stats.Harmonicity = mean(harmonicity)
	if len(harmonicity) > 0 {
		var n int
		for _, h := range harmonicity {
			if h >= voiced {
				n++
			}
		}
		stats.Voiced = float64(n) / float64(len(harmonicity))
	}
}

// modulation returns the fraction of the energy envelope's modulation which
// is at the syllable rate of speech. The envelope is sampled at rate Hz
func modulation(envelope []float64, rate float64) float64 {
	n := len(envelope)
	if n < 2 || float64(n)/rate < 1/minSyllableRate {
		return 0
	}

	// Remove the mean, and transform
	buf := make([]complex128, 1<<bits.Len(uint(n-1)))
	m := mean(envelope)
	for i, e := range envelope {
		buf[i] = complex(e-m, 0)
	}
	fft(buf)

	var band, total float64
	for i := 1; i <= len(buf)/2; i++ {
		freq := float64(i) * rate / float64(len(buf))
		p := cmplx.Abs(buf[i])
		p *= p
		total += p
		if freq >= minSyllableRate && freq <= maxSyllableRate {
			band += p
		}
	}
	if total == 0 {
		return 0
	}
	return band / total
}

// onsets returns how much the log energy rises from each frame to the next
func onsets(energy []float64) []float64 {
	const epsilon = 1e-6
	if len(energy) < 2 {
		return nil
	}
	result := make([]float64, len(energy)-1)
	for i := range result {
		result[i] = max(0, math.Log(energy[i+1]+epsilon)-math.Log(energy[i]+epsilon))
	}
	return result
}

// rhythm returns the largest normalised autocorrelation of a signal at a lag
// which is a tempo between minTempo and maxTempo, which is close to one for a
// steady beat. The signal is sampled at rate Hz, and must be longer than the
// slowest beat
func rhythm(signal []float64, rate float64) float64 {
	minLag := int(rate * 60 / maxTempo)
	maxLag := int(rate * 60 / minTempo)
	if len(signal) <= maxLag {
		return 0
	}

	// Remove the mean
	m := mean(signal)
	x := make([]float64, len(signal))
	var energy float64
	for i, v := range signal {
		x[i] = v - m
		energy += x[i] * x[i]
	}
	if energy == 0 {
		return 0
	}

	var result float64
	for lag := minLag; lag <= maxLag; lag++ {
		var xy, xx, yy float64
		for i := 0; i+lag < len(x); i++ {
			xy += x[i] * x[i+lag]
			xx += x[i] * x[i]
			yy += x[i+lag] * x[i+lag]
		}
		if xx > 0 && yy > 0 {
			result = max(result, xy/math.Sqrt(xx*yy))
		}
	}
	return result
}

// fft transforms x in place. The length of x must be a power of two
func fft(x []complex128) {
	n := len(x)
	if n < 2 {
		return
	}

	// Bit-reversal permutation
	shift := bits.UintSize - bits.Len(uint(n-1))
	for i := range x {
		if j := int(bits.Reverse(uint(i)) >> shift); i < j {
			x[i], x[j] = x[j], x[i]
		}
	}

	// Butterflies
	for size := 2; size <= n; size <<= 1 {
		step := cmplx.Exp(complex(0, -2*math.Pi/float64(size)))
		for start := 0; start < n; start += size {
			w := complex(1, 0)
			for k := range size / 2 {
				a, b := x[start+k], x[start+k+size/2]*w
				x[start+k], x[start+k+size/2] = a+b, a-b
				w *= step
			}
		}
	}
}

func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stddev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := mean(values)
	var sum float64
	for _, v := range values {
		sum += (v - m) * (v - m)
	}
	return math.Sqrt(sum / float64(len(values)))
}
