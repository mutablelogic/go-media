package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"time"

	// Packages
	frame "github.com/mutablelogic/go-media/frame"
	reader "github.com/mutablelogic/go-media/reader"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	writer "github.com/mutablelogic/go-media/writer"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// tap decodes audio from a URL or device into mono int16 samples at a sample
// rate, measures each frame of them, and writes them to a pipe as raw s16le,
// which the segmenter reads
type tap struct {
	reader     *reader.Reader
	stream     int
	sampleRate int
	level      func(level) // called with the level of each frame
	pr         *io.PipeReader
	pw         *io.PipeWriter
}

// level is the level of a frame of samples
type level struct {
	time time.Duration // time of the end of the frame
	rms  float64       // RMS energy (0.0-1.0)
	peak float64       // largest absolute sample value (0.0-1.0)
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// newTap opens audio from a path, URL or device address, with an input format
// and options if they're set
func newTap(path string, stream *int, sampleRate int, format string, opts []string, fn func(level)) (*tap, error) {
	r, err := reader.Open(path, reader.WithInput(format, opts...))
	if err != nil {
		return nil, err
	}

	// Choose the audio stream
	var audio []int
	for index, s := range r.Streams() {
		if s.Par().CodecType() == ff.AVMEDIA_TYPE_AUDIO {
			audio = append(audio, index)
		}
	}
	slices.Sort(audio)
	s := &tap{reader: r, sampleRate: sampleRate, level: fn}
	switch {
	case stream != nil && slices.Contains(audio, *stream):
		s.stream = *stream
	case stream != nil:
		return nil, errors.Join(fmt.Errorf("stream %d is not an audio stream", *stream), r.Close())
	case len(audio) == 0:
		return nil, errors.Join(fmt.Errorf("no audio stream"), r.Close())
	default:
		s.stream = audio[0]
	}

	s.pr, s.pw = io.Pipe()
	return s, nil
}

// Close releases the reader
func (s *tap) Close() error {
	return errors.Join(s.pr.Close(), s.reader.Close())
}

// stop stops writing samples, if the segmenter has stopped reading them
func (s *tap) stop() {
	s.pr.CloseWithError(io.ErrClosedPipe)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Reader returns the pipe which the samples are written to, as raw s16le
func (s *tap) Reader() io.Reader {
	return s.pr
}

// inputOpts returns the reader input format and options for the samples
// written to the pipe. The format is known, so it's probed as little as
// possible, so the segmenter starts as soon as there are samples
func (s *tap) inputOpts() (string, []string) {
	return "s16le", []string{fmt.Sprintf("sample_rate=%d", s.sampleRate), "channels=1", "probesize=32", "analyzeduration=0", "fpsprobesize=0"}
}

// run decodes the audio, and writes the samples to the pipe, until the end of
// the audio or until the context is cancelled. Either way, the pipe is then
// closed, so the segmenter outputs the samples it has
func (s *tap) run(ctx context.Context) error {
	err := s.decode(ctx)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		err = nil
	}
	s.pw.CloseWithError(err)
	return err
}

func (s *tap) decode(ctx context.Context) error {
	par := ff.AVCodec_parameters_alloc()
	if par == nil {
		return fmt.Errorf("failed to allocate codec parameters")
	}
	defer ff.AVCodec_parameters_free(par)
	par.SetCodecType(ff.AVMEDIA_TYPE_AUDIO)
	par.SetSampleFormat(ff.AV_SAMPLE_FMT_S16)
	par.SetSampleRate(s.sampleRate)
	var layout ff.AVChannelLayout
	ff.AVUtil_channel_layout_default(&layout, 1)
	if err := par.SetChannelLayout(layout); err != nil {
		return err
	}
	resampler, err := writer.NewResampler(par, 0, ff.AVRational{})
	if err != nil {
		return err
	}
	defer resampler.Close()

	// Measure each frame of samples, and write them to the pipe
	var n int64
	var buf []byte
	write := func(f frame.Frame) error {
		samples := f.(*frame.AudioFrame).Int16(0)
		var sum, peak float64
		buf = buf[:0]
		for _, sample := range samples {
			v := float64(sample) / math.MaxInt16
			sum += v * v
			peak = max(peak, math.Abs(v))
			buf = binary.LittleEndian.AppendUint16(buf, uint16(sample))
		}
		n += int64(len(samples))
		if len(samples) > 0 && s.level != nil {
			s.level(level{
				time: time.Duration(n) * time.Second / time.Duration(s.sampleRate),
				rms:  math.Sqrt(sum / float64(len(samples))),
				peak: peak,
			})
		}
		_, err := s.pw.Write(buf)
		return err
	}

	decoder, err := reader.NewDecoder(func(f frame.Frame) error {
		return resampler.Process(f, write)
	})
	if err != nil {
		return err
	}
	if err := decoder.Add(s.stream); err != nil {
		return err
	}
	if err := s.reader.Decode(ctx, nil, decoder); err != nil {
		return err
	}
	return resampler.Process(nil, write)
}
