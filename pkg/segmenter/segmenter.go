package segmenter

import (
	"context"
	"errors"
	"io"
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
	ts          time.Duration
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
// LIFECYCLE

// Create a new segmenter with a reader r which segments into raw audio of 'dur'
// length. If dur is zero then no segmenting is performed, the whole
// audio file is read and output in one go, which could cause some memory issues.
// The sample rate is the number of samples per second.
//
// At the moment, the audio format is auto-detected, but there should be
// a way to specify the audio format.
func NewReader(r io.Reader, dur time.Duration, sample_rate int) (*Segmenter, error) {
	segmenter := new(Segmenter)

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
	if s.n > 0 {
		s.buf_flt = make([]float32, 0, s.n)
	}

	// Decode samples and segment
	if err := s.reader.Decode(ctx, mapFunc, func(stream int, frame *ffmpeg.Frame) error {
		// We get null frames sometimes, ignore them
		if frame == nil {
			return nil
		}

		// Append float32 samples from plane 0 to buffer
		s.buf_flt = append(s.buf_flt, frame.Float32(0)...)

		// n != 0 and len(buf) >= n we have a segment to process
		if s.n != 0 && len(s.buf_flt) >= s.n {
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
	if s.n > 0 {
		s.buf_s16 = make([]int16, 0, s.n)
	}

	// Decode samples and segment
	if err := s.reader.Decode(ctx, mapFunc, func(stream int, frame *ffmpeg.Frame) error {
		// We get null frames sometimes, ignore them
		if frame == nil {
			return nil
		}

		// Append int16 samples from plane 0 to buffer
		s.buf_s16 = append(s.buf_s16, frame.Int16(0)...)

		// n != 0 and len(buf) >= n we have a segment to process
		if s.n != 0 && len(s.buf_s16) >= s.n {
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
	// Not segmenting
	if s.n == 0 {
		return fn(s.ts, s.buf_flt)
	}

	// Split into n-sized segments
	bufLength := len(s.buf_flt)
	ts := s.ts
	tsinc := time.Duration(s.n) * time.Second / time.Duration(s.sample_rate)
	for i := 0; i < bufLength; i += s.n {
		end := i + s.n
		var segment []float32
		if end <= bufLength {
			// If the segment fits exactly or there are enough items
			segment = s.buf_flt[i:end]
		} else {
			// If the segment is smaller than segmentSize, pad with zeros
			segment = make([]float32, s.n)
			copy(segment, s.buf_flt[i:bufLength])
		}
		if err := fn(ts, segment); err != nil {
			return err
		} else {
			ts += tsinc
		}
	}

	// Return success
	return nil
}

func (s *Segmenter) segment_s16(fn SegmentFuncInt16) error {
	// Not segmenting
	if s.n == 0 {
		return fn(s.ts, s.buf_s16)
	}

	// Split into n-sized segments
	bufLength := len(s.buf_s16)
	ts := s.ts
	tsinc := time.Duration(s.n) * time.Second / time.Duration(s.sample_rate)
	for i := 0; i < bufLength; i += s.n {
		end := i + s.n
		var segment []int16
		if end <= bufLength {
			// If the segment fits exactly or there are enough items
			segment = s.buf_s16[i:end]
		} else {
			// If the segment is smaller than segmentSize, pad with zeros
			segment = make([]int16, s.n)
			copy(segment, s.buf_s16[i:bufLength])
		}
		if err := fn(ts, segment); err != nil {
			return err
		} else {
			ts += tsinc
		}
	}

	// Return success
	return nil
}
