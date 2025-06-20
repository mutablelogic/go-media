package segmenter

import (
	"context"
	"errors"
	"io"
	"math"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// A segmenter reads audio samples from a reader and segments them into
// fixed-size chunks. The segmenter can be used to process audio samples
type Segmenter struct {
	opts
	ts          time.Duration
	sts         float64 // silence timestamps
	sample_rate int
	n           int
	buf_flt     []float32
	buf_s16     []int16
	reader      *ffmpeg.Reader
}

// SegmentFuncFloat is a callback function which is called for the next
// segment of audio samples. The first argument is the timestamp of the segment.
type SegmentFuncFloat32 func(time.Duration, []float32) error

// SegmentFuncInt16 is a callback function which is called for the next
// segment of audio samples. The first argument is the timestamp of the segment.
type SegmentFuncInt16 func(time.Duration, []int16) error

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	Int16Gain = float64(math.MaxInt16) // Gain for converting int16 to float32
)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new segmenter with a reader r which segments into raw audio.
// The sample rate is the number of samples per second.
//
// Setting option WithSegmentSize will cause the segmenter to segment the audio
// into fixed-size chunks approximately of the specified duration.
//
// Setting option WithDefaultSilenceThreshold will cause the segmenter to break
// into smaller chunks, if silence is detected. The length of the silence is
// specified by the WithSilenceDuration option, which defaults to 2 seconds.
//
// At the moment, the audio format is auto-detected, but there should be
// a way to specify the audio format. The output samples are always single-channel
// (mono).
func NewReader(r io.Reader, sample_rate int, opts ...Opt) (*Segmenter, error) {
	segmenter := new(Segmenter)

	// Apply options
	if o, err := applyOpts(opts...); err != nil {
		return nil, err
	} else {
		segmenter.opts = *o
	}

	// Check arguments
	if sample_rate <= 0 {
		return nil, media.ErrBadParameter.With("invalid duration or sample rate arguments")
	} else {
		segmenter.sample_rate = sample_rate
	}

	// Sample buffer is duration * sample rate, assuming mono
	segmenter.n = int(segmenter.opts.SegmentSize.Seconds() * float64(sample_rate))

	// Open the file
	media, err := ffmpeg.NewReader(r)
	if err != nil {
		return nil, err
	} else {
		segmenter.reader = media
	}

	return segmenter, nil
}

// Close the segmenter
func (s *Segmenter) Close() error {
	var result error

	if s.reader != nil {
		result = errors.Join(result, s.reader.Close())
	}
	s.reader = nil
	s.buf_flt = nil
	s.buf_s16 = nil

	// Return any errors
	return result
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return current timestamp
func (s *Segmenter) Duration() time.Duration {
	return s.reader.Duration()
}

// Segments are output through a callback, with the samples and a timestamp
// At the moment the "best" audio stream is used, based on ffmpeg heuristic.
func (s *Segmenter) DecodeFloat32(ctx context.Context, fn SegmentFuncFloat32) error {
	// Check input parameters
	if fn == nil {
		return media.ErrBadParameter.With("SegmentFunc is nil")
	}

	// Map function chooses the best audio stream
	mapFunc := func(stream int, params *ffmpeg.Par) (*ffmpeg.Par, error) {
		if stream == s.reader.BestStream(media.AUDIO) {
			return ffmpeg.NewAudioPar("flt", "mono", s.sample_rate)
		}
		// Ignore no-audio streams
		return nil, nil
	}

	// Allocate the buffer
	s.buf_flt = make([]float32, 0, s.n)

	// Reset the silence timestamp
	s.sts = -1

	// Decode samples and segment
	if err := s.reader.Decode(ctx, mapFunc, func(stream int, frame *ffmpeg.Frame) error {
		// Ignore null frames
		if frame == nil {
			return nil
		}

		// Return if the frame is empty
		data := frame.Float32(0)
		if len(data) == 0 {
			return nil
		}

		// Calculate the energy of the frame and determine if we should "cut" the segment
		_, cut := s.detect_silence(frame.Ts(), func() float64 {
			var sum float32
			for _, sample := range data {
				sum += float32(sample) * float32(sample)
			}
			return math.Sqrt(float64(sum)/float64(len(data))) / float64(math.MaxInt16)
		})

		// Append float32 samples from plane 0 to buffer
		s.buf_flt = append(s.buf_flt, frame.Float32(0)...)

		// TODO: If we don't have enough samples for a segment, or we are not cutting,

		// n != 0 and len(buf) >= n we have a segment to process
		if (s.n != 0 && len(s.buf_flt) >= s.n) || cut {
			if err := s.segment_flt(fn); err != nil {
				return err
			}

			// Increment the timestamp
			s.ts += time.Duration(len(s.buf_flt)) * time.Second / time.Duration(s.sample_rate)

			// Clear the buffer
			s.buf_flt = s.buf_flt[:0]
		}

		// Continue processing
		return nil
	}); err != nil {
		return err
	}

	// Output any remaining samples
	if len(s.buf_flt) > 0 {
		if err := s.segment_flt(fn); err != nil {
			return err
		}
	}

	// Increment the timestamp
	s.ts += time.Duration(len(s.buf_flt)) * time.Second / time.Duration(s.sample_rate)

	// Return success
	return nil
}

func (s *Segmenter) detect_silence(ts float64, energy_fn func() float64) (float64, bool) {
	energy := energy_fn()

	// Segmenting or Silence detection is not enabled
	if s.SegmentSize == 0 || s.SilenceThreshold == 0 {
		return energy, false
	}

	// If energy is above the threshold, reset the silence timestamp
	if energy >= s.SilenceThreshold {
		s.sts = -1
		return energy, false
	}

	// Set the first frame of silence
	if s.sts == -1 {
		s.sts = ts
		return energy, false
	}

	// Calculate the silence duration, and consider whether we consider this
	// a segment boundary.
	silence_duration := ts - s.sts
	return energy, silence_duration >= s.SilenceSize.Seconds()
}

// Segments are output through a callback, with the samples and a timestamp
// At the moment the "best" audio stream is used, based on ffmpeg heuristic.
func (s *Segmenter) DecodeInt16(ctx context.Context, fn SegmentFuncInt16) error {
	// Check input parameters
	if fn == nil {
		return media.ErrBadParameter.With("SegmentFunc is nil")
	}

	// Map function chooses the best audio stream
	mapFunc := func(stream int, params *ffmpeg.Par) (*ffmpeg.Par, error) {
		if stream == s.reader.BestStream(media.AUDIO) {
			return ffmpeg.NewAudioPar("s16", "mono", s.sample_rate)
		}
		// Ignore no-audio streams
		return nil, nil
	}

	// Allocate the buffer
	s.buf_s16 = make([]int16, 0, s.n)

	// Reset the silence timestamp
	s.sts = -1

	// Decode samples and segment
	if err := s.reader.Decode(ctx, mapFunc, func(stream int, frame *ffmpeg.Frame) error {
		// Ignore null frames
		if frame == nil {
			return nil
		}

		// Return if the frame is empty
		data := frame.Int16(0)
		if len(data) == 0 {
			return nil
		}

		// Calculate the energy of the frame and determine if we should "cut" the segment
		_, cut := s.detect_silence(frame.Ts(), func() float64 {
			var sum float32
			for _, sample := range data {
				sum += float32(sample) * float32(sample)
			}
			return math.Sqrt(float64(sum)/float64(len(data))) / float64(math.MaxInt16)
		})

		// Append int16 samples from plane 0 to buffer
		s.buf_s16 = append(s.buf_s16, data...)

		// TODO: If we don't have enough samples for a segment, or we are not cutting
		if cut && len(s.buf_s16) < (s.n>>1) {
			cut = false
		}

		// n != 0 and len(buf) >= n we have a segment to process
		if (s.n != 0 && len(s.buf_s16) >= s.n) || cut {
			if err := s.segment_s16(fn); err != nil {
				return err
			}

			// Increment the timestamp
			s.ts += time.Duration(len(s.buf_s16)) * time.Second / time.Duration(s.sample_rate)

			// Clear the buffer
			s.buf_s16 = s.buf_s16[:0]
		}

		// Continue processing
		return nil
	}); err != nil {
		return err
	}

	// Output any remaining samples
	if len(s.buf_s16) > 0 {
		if err := s.segment_s16(fn); err != nil {
			return err
		}
	}

	// Increment the timestamp
	s.ts += time.Duration(len(s.buf_s16)) * time.Second / time.Duration(s.sample_rate)

	// Return success
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (s *Segmenter) segment_flt(fn SegmentFuncFloat32) error {
	return fn(s.ts, s.buf_flt)
}

func (s *Segmenter) segment_s16(fn SegmentFuncInt16) error {
	return fn(s.ts, s.buf_s16)
}
