package segmenter

import (
	"context"
	"errors"
	"io"
	"slices"
	"time"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	frame "github.com/mutablelogic/go-media/frame"
	reader "github.com/mutablelogic/go-media/reader"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	writer "github.com/mutablelogic/go-media/writer"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Segmenter reads audio samples from a reader and segments them into
// chunks. It can be used to process audio samples in chunks for speech
// recognition, audio analysis, etc.
type Segmenter struct {
	opts
	reader      *reader.Reader
	stream      int // index of the audio stream to decode
	sampleRate  int
	segmentSize int // number of samples per segment, or zero
	frameSize   int // number of samples per frame which is analysed

	// State while decoding
	segment  Segment
	analyser *analyser
	ts       int64 // number of samples output before the current segment
	silence  silence
}

// silence is the run of silent frames at the end of the current segment
type silence struct {
	start  int  // index of the first silent sample in the segment, or -1 if not silent
	frame  int  // index of the first silent frame in the analyser's segment
	n      int  // number of silent samples
	output bool // whether the segment before the silence has been output
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// frameDuration is the duration of each frame of samples which is analysed
// for silence without voice activity detection. Shorter frames make the RMS
// energy of quiet speech pauses cross the silence threshold too often. With
// voice activity detection, frames are VADFrameDuration long
const frameDuration = 100 * time.Millisecond

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Open creates a new segmenter from a file path or URL.
// The sampleRate is the output sample rate in Hz (e.g., 16000 for speech).
// The output is always mono (single channel).
//
// Options:
//   - WithSegments(duration) - output segments of approximately this duration
//   - WithSegments(time, time, ...) - output segments cut near these times
//   - WithMode(mode, silence, threshold) - also break segments where silence has
//     lasted at least the silence duration, detected by RMS energy (ModeSilence)
//     or voice activity detection (ModeVAD...)
//   - WithStream(index) - the audio stream to decode
//   - WithReaderOpt(opt) - options for the reader
func Open(url string, sampleRate int, opts ...Opt) (*Segmenter, error) {
	s, err := newSegmenter(sampleRate, opts...)
	if err != nil {
		return nil, err
	}
	r, err := reader.Open(url, s.readerOpts...)
	if err != nil {
		return nil, err
	}
	return s.open(r)
}

// NewReader creates a new segmenter from an io.Reader.
// See Open for parameter documentation.
func NewReader(r io.Reader, sampleRate int, opts ...Opt) (*Segmenter, error) {
	s, err := newSegmenter(sampleRate, opts...)
	if err != nil {
		return nil, err
	}
	rr, err := reader.NewReader(r, s.readerOpts...)
	if err != nil {
		return nil, err
	}
	return s.open(rr)
}

func newSegmenter(sampleRate int, opts ...Opt) (*Segmenter, error) {
	s := new(Segmenter)

	// Apply options
	o, err := applyOpts(opts...)
	if err != nil {
		return nil, err
	}
	s.opts = *o

	// Validate sample rate
	if sampleRate <= 0 {
		return nil, gomedia.ErrBadParameter.With("sample rate must be positive")
	}
	s.sampleRate = sampleRate

	// Calculate segment and frame sizes in samples (mono). With boundaries,
	// short segments are joined together up to each one
	if s.SegmentSize > 0 {
		s.segmentSize = int(s.SegmentSize.Seconds() * float64(sampleRate))
	} else if len(s.Boundaries) > 0 {
		s.segmentSize = int(pieceDuration.Seconds() * float64(sampleRate))
	}
	if _, vad := s.Mode.vad(); vad {
		s.frameSize = int(VADFrameDuration.Seconds() * float64(sampleRate))
	} else {
		s.frameSize = max(1, int(frameDuration.Seconds()*float64(sampleRate)))
	}

	return s, nil
}

// open selects the audio stream to decode from r, and closes r on error
func (s *Segmenter) open(r *reader.Reader) (*Segmenter, error) {
	var audio []int
	for index, stream := range r.Streams() {
		if stream.Par().CodecType() == ff.AVMEDIA_TYPE_AUDIO {
			audio = append(audio, index)
		}
	}
	slices.Sort(audio)

	switch {
	case s.opts.stream >= 0 && !slices.Contains(audio, s.opts.stream):
		return nil, errors.Join(gomedia.ErrBadParameter.Withf("stream %d is not an audio stream", s.opts.stream), r.Close())
	case s.opts.stream >= 0:
		s.stream = s.opts.stream
	case len(audio) == 0:
		return nil, errors.Join(gomedia.ErrNotFound.With("no audio stream"), r.Close())
	default:
		s.stream = audio[0]
	}

	s.reader = r
	return s, nil
}

// Close releases all resources associated with the segmenter.
func (s *Segmenter) Close() error {
	var result error
	if s.reader != nil {
		result = errors.Join(result, s.reader.Close())
	}
	s.reader = nil
	s.segment = Segment{}
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

// Stream returns the index of the audio stream which is decoded.
func (s *Segmenter) Stream() int {
	return s.stream
}

// Decode decodes the audio stream into mono float32 samples at the
// configured sample rate, and calls fn for each segment. When the context is
// cancelled, such as to stop recording from a device, the samples so far are
// output as the last segment, and the context's error is returned.
func (s *Segmenter) Decode(ctx context.Context, fn SegmentFn) error {
	if fn == nil {
		return gomedia.ErrBadParameter.With("callback function is nil")
	}
	if s.reader == nil {
		return gomedia.ErrInternalError.With("segmenter is closed")
	}

	// Convert to mono float32 at the sample rate, in frames of frameSize
	resampler, release, err := newResampler(ff.AV_SAMPLE_FMT_FLT, s.sampleRate, s.frameSize)
	if err != nil {
		return err
	}
	defer release()
	if s.analyser, err = newAnalyser(s.sampleRate, s.Mode, s.SilenceThreshold); err != nil {
		return err
	}
	defer func() {
		s.analyser.Close()
		s.analyser = nil
	}()

	// With voice activity detection, also convert to mono int16 at the
	// sample rate it uses
	var vadResampler *writer.Resampler
	if _, vad := s.Mode.vad(); vad {
		var release func()
		if vadResampler, release, err = newResampler(ff.AV_SAMPLE_FMT_S16, vadSampleRate, vadFrameSize()); err != nil {
			return err
		}
		defer release()
	}
	vad := func(f frame.Frame) error {
		return s.analyser.writeVAD(f.(*frame.AudioFrame).Int16(0))
	}

	// Reset state
	if s.segmentSize > 0 {
		s.segment.Samples = make([]float32, 0, s.segmentSize)
	} else {
		s.segment.Samples = make([]float32, 0, s.sampleRate) // 1 second default capacity
	}
	s.ts = 0
	s.silence = silence{start: -1}

	// With boundaries, join segments together up to each one
	var j *joiner
	if len(s.Boundaries) > 0 {
		j = newJoiner(slices.Clone(s.Boundaries), fn)
		fn = j.write
	}

	// Each resampled frame is analysed, and output any segments. A callback
	// returning io.EOF stops decoding, after which nothing more is output
	var stopped bool
	process := func(f frame.Frame) error {
		err := s.process(f.(*frame.AudioFrame).Float32(0), fn)
		if errors.Is(err, io.EOF) {
			stopped = true
		}
		return err
	}

	// Decode frames, detecting voice activity in each one before analysing it
	decoder, err := reader.NewDecoder(func(f frame.Frame) error {
		if vadResampler != nil {
			if err := vadResampler.Process(f, vad); err != nil {
				return err
			}
		}
		return resampler.Process(f, process)
	})
	if err != nil {
		return err
	}
	if err := decoder.Add(s.stream); err != nil {
		return err
	}
	result := s.reader.Decode(ctx, nil, decoder)
	if result != nil && !errors.Is(result, context.Canceled) && !errors.Is(result, context.DeadlineExceeded) {
		return result
	}
	if stopped {
		return nil
	}

	// Flush the resamplers, and output any remaining samples
	if vadResampler != nil {
		if err := vadResampler.Process(nil, vad); err != nil {
			return err
		}
	}
	if err := resampler.Process(nil, process); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	if len(s.segment.Samples) > 0 {
		if err := s.flush(fn, len(s.segment.Samples)); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
	if j != nil {
		if err := j.close(); err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	}

	return result
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// newResampler returns a resampler which converts audio frames to mono
// samples in a sample format and rate, in frames of frameSize samples, and a
// function which releases it. The resampler uses the codec parameters until
// it's released
func newResampler(format ff.AVSampleFormat, sampleRate, frameSize int) (*writer.Resampler, func(), error) {
	par := ff.AVCodec_parameters_alloc()
	if par == nil {
		return nil, nil, gomedia.ErrInternalError.With("failed to allocate codec parameters")
	}
	par.SetCodecType(ff.AVMEDIA_TYPE_AUDIO)
	par.SetSampleFormat(format)
	par.SetSampleRate(sampleRate)
	var layout ff.AVChannelLayout
	ff.AVUtil_channel_layout_default(&layout, 1)
	if err := par.SetChannelLayout(layout); err != nil {
		ff.AVCodec_parameters_free(par)
		return nil, nil, err
	}
	resampler, err := writer.NewResampler(par, frameSize, ff.AVRational{})
	if err != nil {
		ff.AVCodec_parameters_free(par)
		return nil, nil, err
	}
	return resampler, func() {
		resampler.Close()
		ff.AVCodec_parameters_free(par)
	}, nil
}

// process analyses a frame of samples, and outputs segments. A segment is
// output as soon as the silence at its end has lasted SilenceSize, or when it
// reaches the segment size. The silence is then output as a segment of its
// own, when the silence ends or reaches the segment size
func (s *Segmenter) process(samples []float32, fn SegmentFn) error {
	if len(samples) == 0 {
		return nil
	}
	m, err := s.analyser.measure(samples)
	if err != nil {
		return err
	}

	if s.silent(m) {
		// Silent frame: start/continue silence tracking.
		if s.silence.start < 0 {
			s.silence = silence{start: len(s.segment.Samples), frame: s.analyser.count()}
		}
	} else if s.silence.start >= 0 {
		// Non-silent frame: output silence which lasted long enough, before
		// the frame starts the next segment
		var err error
		if s.silence.output {
			err = s.flush(fn, len(s.segment.Samples))
		}
		s.silence = silence{start: -1}
		if err != nil {
			return err
		}
	}

	// Append samples to buffer
	s.segment.Samples = append(s.segment.Samples, samples...)
	s.analyser.add(samples, m)

	// Output the segment before a silence once it has lasted long enough. If
	// the segment before it is too short, it's output with the silence so far
	if s.silence.start >= 0 && !s.silence.output {
		s.silence.n += len(samples)
		if s.duration(int64(s.silence.n)) >= s.SilenceSize {
			s.silence.output = true
			switch {
			case s.silence.start >= s.minSamples():
				if err := s.flush(fn, s.silence.start); err != nil {
					return err
				}
			case s.silence.start > 0:
				if err := s.flush(fn, len(s.segment.Samples)); err != nil {
					return err
				}
			}
		}
	}

	// Fixed-size segment split
	if s.segmentSize > 0 && len(s.segment.Samples) >= s.segmentSize {
		if err := s.flush(fn, len(s.segment.Samples)); err != nil {
			return err
		}
	}

	return nil
}

// flush outputs the first n samples in the buffer as a segment, and removes
// them from it
func (s *Segmenter) flush(fn SegmentFn, n int) error {
	frames := s.analyser.count()

	// The segment is silence when it all is, and silence detection is enabled
	silent := s.silence.start == 0 && s.silence.output
	if s.silence.start >= 0 {
		if n == s.silence.start {
			frames = s.silence.frame
		}
		s.silence.start, s.silence.frame = max(0, s.silence.start-n), max(0, s.silence.frame-frames)
	}
	s.segment.Start = s.duration(s.ts)
	s.segment.Duration = s.duration(s.ts+int64(n)) - s.segment.Start
	s.segment.Class, s.segment.Stats = s.analyser.next(frames, silent)

	samples := s.segment.Samples
	s.segment.Samples = samples[:n]
	err := fn(&s.segment)
	s.ts += int64(n)
	s.segment.Samples = samples[:copy(samples, samples[n:])]
	return err
}

// duration returns the duration of n samples
func (s *Segmenter) duration(n int64) time.Duration {
	return time.Duration(n) * time.Second / time.Duration(s.sampleRate)
}

// minSamples returns the minimum number of samples before a silence cut is allowed.
// This prevents very short segments.
func (s *Segmenter) minSamples() int {
	if s.segmentSize > 0 {
		return s.segmentSize / 2 // At least half the segment size
	}
	return s.sampleRate / 10 // At least 100ms
}

// silent returns true when a frame, with measurement m, is silence
func (s *Segmenter) silent(m measurement) bool {
	if s.SilenceThreshold > 0 && m.rms < s.SilenceThreshold {
		return true
	}
	return m.vad && !m.speech
}
