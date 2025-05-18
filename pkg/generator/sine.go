package generator

import (
	"encoding/json"
	"errors"
	"math"
	"time"

	// Packages
	"github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

type sine struct {
	frame     *ffmpeg.Frame
	frequency float64 // in Hz
	volume    float64 // in decibels
}

var _ Generator = (*sine)(nil)

////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	frameDuration = 20 * time.Millisecond // Each frame is 20ms of audio
)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new sine wave generator with one channel using float32
// for samples. The frequency in Hz, volume in decibels and samplerate
// (ie, 44100) for the audio stream are passed as arguments.
func NewSine(freq, volume float64, par *ffmpeg.Par) (*sine, error) {
	sine := new(sine)

	// Check parameters
	if par.Type() != media.AUDIO {
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
	frame, err := ffmpeg.NewFrame(par)
	if err != nil {
		return nil, err
	}

	// Allocate buffer
	if err := frame.AllocateBuffers(); err != nil {
		return nil, errors.Join(err, frame.Close())
	}

	// Set parameters
	sine.frame = frame
	sine.frequency = freq
	sine.volume = volume

	// Return success
	return sine, nil
}

// Free resources for the generator
func (s *sine) Close() error {
	result := s.frame.Close()
	s.frame = nil
	return result
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
func (s *sine) Frame() *ffmpeg.Frame {
	// Make a writable copy if the frame is not writable
	if err := s.frame.MakeWritable(); err != nil {
		return nil
	}

	// Set the Pts
	if s.frame.Pts() == ffmpeg.PTS_UNDEFINED {
		s.frame.SetPts(0)
	} else {
		s.frame.IncPts(int64(s.frame.NumSamples()))
	}

	// Calculate current phase and volume
	t := s.frame.Ts() // Timestamp in seconds
	volume := math.Pow(10, s.volume/20.0)
	data := s.frame.Float32(0)

	// Generate sine wave
	for n := 0; n < s.frame.NumSamples(); n++ {
		sampleTime := t + float64(n)/float64(s.frame.SampleRate())
		data[n] = float32(math.Sin(2.0*math.Pi*s.frequency*sampleTime) * volume)
	}

	// Return the frame
	return s.frame
}
