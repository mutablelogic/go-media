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

// Create a new segmenter with a reader r which segments into raw audio of 'dur'
// length. If dur is zero then no segmenting is performed, the whole
// audio file is read and output in one go, which could cause some memory issues.
// The sample rate is the number of samples per second.
//
// At the moment, the audio format is auto-detected, but there should be
// a way to specify the audio format.
func NewReader(r io.Reader, dur time.Duration, sample_rate int, opts ...Opt) (*Segmenter, error) {
	segmenter := new(Segmenter)

	// Apply options
	if o, err := applyOpts(opts...); err != nil {
		return nil, err
	} else {
		segmenter.opts = *o
	}

	// Check arguments
	if dur < 0 || sample_rate <= 0 {
		return nil, media.ErrBadParameter.With("invalid duration or sample rate arguments")
	} else {
		segmenter.sample_rate = sample_rate
	}

	// Sample buffer is duration * sample rate, assuming mono
	segmenter.n = int(dur.Seconds() * float64(sample_rate))

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

		// Calculate the energy of the frame - root mean squared and normalize between 0 and 1
		var sum float32
		var energy float64
		for _, sample := range data {
			sum += float32(sample) * float32(sample)
		}
		energy = math.Sqrt(float64(sum)/float64(len(data))) / float64(math.MaxInt16)

		// If silence detection is enabled, check if the energy is below the threshold
		var cut bool
		if s.SilenceThreshold > 0 && energy < s.SilenceThreshold {
			// If the energy is below the threshold, we consider it silence
			if s.sts == -1 {
				// If this is the first silence, set the timestamp
				s.sts = frame.Ts()
			} else if frame.Ts()-s.sts >= s.SilenceDuration.Seconds() {
				// Cut when the buffer size is greater than 10 seconds
				if len(s.buf_flt) >= s.sample_rate*10 {
					cut = true
				}
				s.sts = -1 // Reset the silence timestamp
			}
		}

		// Append float32 samples from plane 0 to buffer
		s.buf_flt = append(s.buf_flt, frame.Float32(0)...)

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

		// Calculate the energy of the frame - root mean squared and normalize between 0 and 1
		var sum float32
		var energy float64
		for _, sample := range data {
			sum += float32(sample) * float32(sample)
		}
		energy = math.Sqrt(float64(sum)/float64(len(data))) / float64(math.MaxInt16)

		// If silence detection is enabled, check if the energy is below the threshold
		var cut bool
		if s.SilenceThreshold > 0 && energy < s.SilenceThreshold {
			// If the energy is below the threshold, we consider it silence
			if s.sts == -1 {
				// If this is the first silence, set the timestamp
				s.sts = frame.Ts()
			} else if frame.Ts()-s.sts >= s.SilenceDuration.Seconds() {
				// Cut when the buffer size is greater than 10 seconds
				if len(s.buf_s16) >= s.sample_rate*10 {
					cut = true
				}
				s.sts = -1 // Reset the silence timestamp
			}
		}

		// Append int16 samples from plane 0 to buffer
		s.buf_s16 = append(s.buf_s16, data...)

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
	// TODO: Pad any remaining samples with zeros if the buffer is not full
	return fn(s.ts, s.buf_flt)
}

func (s *Segmenter) segment_s16(fn SegmentFuncInt16) error {
	// TODO: Pad any remaining samples with zeros if the buffer is not full
	return fn(s.ts, s.buf_s16)
}
