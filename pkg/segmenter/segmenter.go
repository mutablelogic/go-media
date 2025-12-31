package segmenter

import (
	"context"
	"errors"
	"io"
	"math"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg80"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Segmenter reads audio samples from a reader and segments them into
// fixed-size chunks. It can be used to process audio samples in chunks
// for speech recognition, audio analysis, etc.
type Segmenter struct {
	opts
	reader      *ffmpeg.Reader
	sampleRate  int
	segmentSize int           // number of samples per segment
	ts          time.Duration // current timestamp
	silenceTs   float64       // timestamp when silence started (-1 if not in silence)
	bufFloat32  []float32
	bufInt16    []int16
}

// SegmentFuncFloat32 is a callback function which is called for each segment
// of audio samples. The first argument is the timestamp of the segment start.
// Return nil to continue, io.EOF to stop early, or any other error to abort.
type SegmentFuncFloat32 func(timestamp time.Duration, samples []float32) error

// SegmentFuncInt16 is a callback function which is called for each segment
// of audio samples. The first argument is the timestamp of the segment start.
// Return nil to continue, io.EOF to stop early, or any other error to abort.
type SegmentFuncInt16 func(timestamp time.Duration, samples []int16) error

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a new segmenter from a file path or URL.
// The sampleRate is the output sample rate in Hz (e.g., 16000 for speech).
// The output is always mono (single channel).
//
// Options:
//   - WithSegmentSize(duration) - output segments of approximately this duration
//   - WithDefaultSilence() - also break segments on silence boundaries
//   - WithSilenceThreshold(threshold) - custom silence threshold (0.0-1.0)
//   - WithSilenceSize(duration) - minimum silence duration to trigger a break
func New(path string, sampleRate int, opts ...Opt) (*Segmenter, error) {
	s := new(Segmenter)

	// Apply options
	o, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}
	s.opts = *o

	// Validate sample rate
	if sampleRate <= 0 {
		return nil, media.ErrBadParameter.With("sample rate must be positive")
	}
	s.sampleRate = sampleRate

	// Calculate segment size in samples (mono)
	if s.SegmentSize > 0 {
		s.segmentSize = int(s.SegmentSize.Seconds() * float64(sampleRate))
	}

	// Open the file
	reader, err := ffmpeg.Open(path)
	if err != nil {
		return nil, err
	}
	s.reader = reader

	return s, nil
}

// NewFromReader creates a new segmenter from an io.Reader.
// See New for parameter documentation.
func NewFromReader(r io.Reader, sampleRate int, opts ...Opt) (*Segmenter, error) {
	s := new(Segmenter)

	// Apply options
	o, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}
	s.opts = *o

	// Validate sample rate
	if sampleRate <= 0 {
		return nil, media.ErrBadParameter.With("sample rate must be positive")
	}
	s.sampleRate = sampleRate

	// Calculate segment size in samples (mono)
	if s.SegmentSize > 0 {
		s.segmentSize = int(s.SegmentSize.Seconds() * float64(sampleRate))
	}

	// Open the reader
	reader, err := ffmpeg.NewReader(r)
	if err != nil {
		return nil, err
	}
	s.reader = reader

	return s, nil
}

// Close releases all resources associated with the segmenter.
func (s *Segmenter) Close() error {
	var result error
	if s.reader != nil {
		result = errors.Join(result, s.reader.Close())
	}
	s.reader = nil
	s.bufFloat32 = nil
	s.bufInt16 = nil
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Duration returns the total duration of the media stream.
// Returns zero if duration is unknown.
func (s *Segmenter) Duration() time.Duration {
	return s.reader.Duration()
}

// SampleRate returns the configured output sample rate.
func (s *Segmenter) SampleRate() int {
	return s.sampleRate
}

// DecodeFloat32 decodes the audio stream into float32 samples and calls the
// callback function for each segment. Samples are in the range [-1.0, 1.0].
//
// The "best" audio stream is automatically selected. Output is mono at the
// configured sample rate.
func (s *Segmenter) DecodeFloat32(ctx context.Context, fn SegmentFuncFloat32) error {
	if fn == nil {
		return media.ErrBadParameter.With("callback function is nil")
	}

	// Map function selects best audio stream and converts to float32 mono
	mapFunc := func(stream int, params *ffmpeg.Par) (*ffmpeg.Par, error) {
		if stream == s.reader.BestStream(media.AUDIO) {
			return ffmpeg.NewAudioPar("flt", "mono", s.sampleRate)
		}
		return nil, nil // Ignore other streams
	}

	// Pre-allocate buffer
	if s.segmentSize > 0 {
		s.bufFloat32 = make([]float32, 0, s.segmentSize)
	} else {
		s.bufFloat32 = make([]float32, 0, s.sampleRate) // 1 second default capacity
	}

	// Reset state
	s.ts = 0
	s.silenceTs = -1

	// Decode frames
	if err := s.reader.Demux(ctx, mapFunc, func(stream int, frame *ffmpeg.Frame) error {
		if frame == nil {
			return nil
		}

		// Get samples from plane 0 (mono)
		samples := frame.Float32(0)
		if len(samples) == 0 {
			return nil
		}

		// Calculate RMS energy and check for silence-based cut
		cut := s.detectSilence(frame.Pts(), samples)

		// Append samples to buffer
		s.bufFloat32 = append(s.bufFloat32, samples...)

		// Check if we should output a segment
		shouldOutput := false
		if s.segmentSize > 0 && len(s.bufFloat32) >= s.segmentSize {
			shouldOutput = true
		}
		if cut && len(s.bufFloat32) >= s.minSamples() {
			shouldOutput = true
		}

		if shouldOutput {
			// Capture buffer length before callback (for timestamp calculation)
			bufLen := len(s.bufFloat32)

			// Call the callback with the segment
			if err := fn(s.ts, s.bufFloat32); err != nil {
				return err
			}

			// Update timestamp for next segment
			s.ts += time.Duration(bufLen) * time.Second / time.Duration(s.sampleRate)

			// Clear buffer
			s.bufFloat32 = s.bufFloat32[:0]
		}

		return nil
	}); err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	// Output any remaining samples
	if len(s.bufFloat32) > 0 {
		if err := fn(s.ts, s.bufFloat32); err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	}

	return nil
}

// DecodeInt16 decodes the audio stream into int16 samples and calls the
// callback function for each segment. Samples are in the range [-32768, 32767].
//
// The "best" audio stream is automatically selected. Output is mono at the
// configured sample rate.
func (s *Segmenter) DecodeInt16(ctx context.Context, fn SegmentFuncInt16) error {
	if fn == nil {
		return media.ErrBadParameter.With("callback function is nil")
	}

	// Map function selects best audio stream and converts to int16 mono
	mapFunc := func(stream int, params *ffmpeg.Par) (*ffmpeg.Par, error) {
		if stream == s.reader.BestStream(media.AUDIO) {
			return ffmpeg.NewAudioPar("s16", "mono", s.sampleRate)
		}
		return nil, nil // Ignore other streams
	}

	// Pre-allocate buffer
	if s.segmentSize > 0 {
		s.bufInt16 = make([]int16, 0, s.segmentSize)
	} else {
		s.bufInt16 = make([]int16, 0, s.sampleRate) // 1 second default capacity
	}

	// Reset state
	s.ts = 0
	s.silenceTs = -1

	// Decode frames
	if err := s.reader.Demux(ctx, mapFunc, func(stream int, frame *ffmpeg.Frame) error {
		if frame == nil {
			return nil
		}

		// Get samples from plane 0 (mono)
		samples := frame.Int16(0)
		if len(samples) == 0 {
			return nil
		}

		// Calculate RMS energy and check for silence-based cut
		cut := s.detectSilenceInt16(frame.Pts(), samples)

		// Append samples to buffer
		s.bufInt16 = append(s.bufInt16, samples...)

		// Check if we should output a segment
		shouldOutput := false
		if s.segmentSize > 0 && len(s.bufInt16) >= s.segmentSize {
			shouldOutput = true
		}
		if cut && len(s.bufInt16) >= s.minSamples() {
			shouldOutput = true
		}

		if shouldOutput {
			// Capture buffer length before callback (for timestamp calculation)
			bufLen := len(s.bufInt16)

			// Call the callback with the segment
			if err := fn(s.ts, s.bufInt16); err != nil {
				return err
			}

			// Update timestamp for next segment
			s.ts += time.Duration(bufLen) * time.Second / time.Duration(s.sampleRate)

			// Clear buffer
			s.bufInt16 = s.bufInt16[:0]
		}

		return nil
	}); err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	// Output any remaining samples
	if len(s.bufInt16) > 0 {
		if err := fn(s.ts, s.bufInt16); err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// minSamples returns the minimum number of samples before a silence cut is allowed.
// This prevents very short segments.
func (s *Segmenter) minSamples() int {
	if s.segmentSize > 0 {
		return s.segmentSize / 2 // At least half the segment size
	}
	return s.sampleRate / 10 // At least 100ms
}

// detectSilence checks if we should cut based on silence detection (float32 samples).
// Returns true if silence has been detected for long enough to trigger a cut.
func (s *Segmenter) detectSilence(pts int64, samples []float32) bool {
	// Silence detection not enabled
	if s.SilenceThreshold == 0 {
		return false
	}

	// Calculate RMS energy for float32 samples (already in -1.0 to 1.0 range)
	var sum float64
	for _, sample := range samples {
		sum += float64(sample) * float64(sample)
	}
	energy := math.Sqrt(sum / float64(len(samples)))

	return s.checkSilenceDuration(float64(pts), energy)
}

// detectSilenceInt16 checks if we should cut based on silence detection (int16 samples).
// Returns true if silence has been detected for long enough to trigger a cut.
func (s *Segmenter) detectSilenceInt16(pts int64, samples []int16) bool {
	// Silence detection not enabled
	if s.SilenceThreshold == 0 {
		return false
	}

	// Calculate RMS energy for int16 samples, normalized to 0.0-1.0 range
	var sum float64
	for _, sample := range samples {
		norm := float64(sample) / float64(math.MaxInt16)
		sum += norm * norm
	}
	energy := math.Sqrt(sum / float64(len(samples)))

	return s.checkSilenceDuration(float64(pts), energy)
}

// checkSilenceDuration checks if we've been in silence long enough to trigger a cut.
func (s *Segmenter) checkSilenceDuration(ts float64, energy float64) bool {
	// If energy is above threshold, we're not in silence
	if energy >= s.SilenceThreshold {
		s.silenceTs = -1
		return false
	}

	// First frame of silence - record timestamp
	if s.silenceTs < 0 {
		s.silenceTs = ts
		return false
	}

	// Check if we've been in silence long enough
	silenceDuration := ts - s.silenceTs
	return silenceDuration >= s.SilenceSize.Seconds()
}
