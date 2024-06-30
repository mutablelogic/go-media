package generator

import (
	"encoding/json"
	"errors"
	"math"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

type sine struct {
	frame     *ff.AVFrame
	frequency float64 // in Hz
	volume    float64 // in decibels
}

var _ Generator = (*sine)(nil)

////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	frameDuration = 10 * time.Millisecond // Each frame is 10ms of audio
)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new sine wave generator with one channel using float32
// for samples. The frequency in Hz, volume in decibels and samplerate
// (ie, 44100) for the audio stream are passed as arguments.
func NewSine(freq, volume float64, par *ffmpeg.Par) (*sine, error) {
	sine := new(sine)

	// Check parameters
	if par.CodecType() != ff.AVMEDIA_TYPE_AUDIO {
		return nil, errors.New("invalid codec type")
	} else if par.ChannelLayout().NumChannels() != 1 {
		return nil, errors.New("invalid channel layout, only mono is supported")
	} else if par.SampleFormat() != ff.AV_SAMPLE_FMT_FLT && par.SampleFormat() != ff.AV_SAMPLE_FMT_FLTP {
		return nil, errors.New("invalid sample format, only float32 is supported")
	}
	if freq <= 0 {
		return nil, errors.New("invalid frequency")
	}
	if volume <= -100 {
		return nil, errors.New("invalid volume")
	}
	if par.Samplerate() <= 0 {
		return nil, errors.New("invalid samplerate")
	}
	if par.FrameSize() <= 0 {
		par.SetFrameSize(int(float64(par.Samplerate()) * frameDuration.Seconds()))
	}

	// Create a frame
	frame := ff.AVUtil_frame_alloc()
	if frame == nil {
		return nil, errors.New("failed to allocate frame")
	}

	frame.SetSampleFormat(ff.AV_SAMPLE_FMT_FLT) // float32
	if err := frame.SetChannelLayout(ff.AV_CHANNEL_LAYOUT_MONO); err != nil {
		return nil, err
	}
	frame.SetSampleRate(par.Samplerate())
	frame.SetNumSamples(par.FrameSize())
	frame.SetTimeBase(ff.AVUtil_rational(1, par.Samplerate()))
	frame.SetPts(ff.AV_NOPTS_VALUE)

	// Allocate buffer
	if err := ff.AVUtil_frame_get_buffer(frame, false); err != nil {
		return nil, err
	} else {
		sine.frame = frame
		sine.frequency = freq
		sine.volume = volume
	}

	// Return success
	return sine, nil
}

// Free resources for the generator
func (s *sine) Close() error {
	ff.AVUtil_frame_free(s.frame)
	s.frame = nil
	return nil
}

////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (s *sine) String() string {
	data, _ := json.MarshalIndent(s.frame, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the first and subsequent frames of raw audio data
func (s *sine) Frame() media.Frame {
	if err := ff.AVUtil_frame_make_writable(s.frame); err != nil {
		return nil
	}

	// Set the Pts
	if s.frame.Pts() == ff.AV_NOPTS_VALUE {
		s.frame.SetPts(0)
	} else {
		s.frame.SetPts(s.frame.Pts() + int64(s.frame.NumSamples()))
	}

	// Calculate current phase and volume
	t := ff.AVUtil_rational_q2d(s.frame.TimeBase()) * float64(s.frame.Pts())
	volume := math.Pow(10, s.volume/20.0)
	data := s.frame.Float32(0)

	// Generate sine wave
	for n := 0; n < s.frame.NumSamples(); n++ {
		sampleTime := t + float64(n)/float64(s.frame.SampleRate())
		data[n] = float32(math.Sin(2.0*math.Pi*s.frequency*sampleTime) * volume)
	}

	// Return the frame
	return ffmpeg.NewFrame(s.frame, 0)
}
