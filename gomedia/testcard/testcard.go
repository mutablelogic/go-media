// Package testcard implements a default gomedia/schema.Source: a generated
// sine wave tone, suitable as a fallback for a live encoding job when no
// other source is available. Video (a picture test card) is not implemented
// yet — only the audio tone.
package testcard

import (
	"context"
	"math"
	"time"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	frame "github.com/mutablelogic/go-media/frame"
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	profile "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// TestCard is a Source that generates a mono sine wave tone. It does not
// pace itself to real time — NextFrame returns as soon as the next chunk is
// generated. Pacing is the caller's job: it's dictated by the destination
// (a live streaming job wants real-time pacing, a batch re-encode doesn't),
// not something a source should assume.
type TestCard struct {
	audio     *profile.AudioProfile
	frequency float64 // Hz
	volume    float64 // dB

	frame   *frame.AudioFrame
	started bool // Whether the first frame has been generated
}

var _ schema.Source = (*TestCard)(nil)

// Opt configures optional TestCard parameters.
type Opt func(*TestCard) error

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	DefaultFrequency = 1000.0 // 1kHz tone
	DefaultVolume    = -20.0  // dB, a conservative default level

	// frameDuration is not driven by any target encoder's fixed frame size
	// (the test card doesn't know what it'll be encoded as) — it's just a
	// reasonable chunk size for a live source.
	frameDuration = 20 * time.Millisecond
)

//////////////////////////////////////////////////////////////////////////////
// PUBLIC OPTIONS

// WithFrequency sets the tone's frequency in Hz. Default is 1kHz.
func WithFrequency(hz float64) Opt {
	return func(t *TestCard) error {
		if hz <= 0 {
			return gomedia.ErrBadParameter.With("frequency must be positive")
		}
		t.frequency = hz
		return nil
	}
}

// WithVolume sets the tone's volume in decibels. Default is -20dB.
func WithVolume(db float64) Opt {
	return func(t *TestCard) error {
		if db <= -100 {
			return gomedia.ErrBadParameter.With("volume must be greater than -100dB")
		}
		t.volume = db
		return nil
	}
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// New creates a test card source generating a sine wave tone shaped by the
// given audio profile's sample rate, sample format and channel layout. Only
// mono, float32 profiles are supported for now (e.g. a profile.AudioProfile
// for "pcm_f32le" with channel_layout "mono").
func New(audio *profile.AudioProfile, opts ...Opt) (*TestCard, error) {
	if audio == nil {
		return nil, gomedia.ErrBadParameter.With("nil audio profile")
	}

	par := audio.Par()
	if par.ChannelLayout().NumChannels() != 1 {
		return nil, gomedia.ErrBadParameter.With("only mono is supported")
	}
	if par.SampleFormat() != ff.AV_SAMPLE_FMT_FLT && par.SampleFormat() != ff.AV_SAMPLE_FMT_FLTP {
		return nil, gomedia.ErrBadParameter.With("only float32 sample formats are supported")
	}
	if par.SampleRate() <= 0 {
		return nil, gomedia.ErrBadParameter.With("audio profile has no sample rate")
	}

	self := &TestCard{
		audio:     audio,
		frequency: DefaultFrequency,
		volume:    DefaultVolume,
	}
	for _, opt := range opts {
		if err := opt(self); err != nil {
			return nil, err
		}
	}

	// Allocate the frame once and reuse it for every call to NextFrame.
	f, err := frame.NewAudioFrame(0)
	if err != nil {
		return nil, err
	}
	f.SetSampleFormat(par.SampleFormat())
	f.SetSampleRate(par.SampleRate())
	if err := f.SetChannelLayout(par.ChannelLayout()); err != nil {
		f.Close()
		return nil, err
	}
	f.SetNumSamples(int(float64(par.SampleRate()) * frameDuration.Seconds()))
	if err := f.AllocateBuffers(); err != nil {
		f.Close()
		return nil, err
	}

	self.frame = f
	return self, nil
}

// Close releases the test card's frame.
func (t *TestCard) Close() error {
	if t.frame != nil {
		err := t.frame.Close()
		t.frame = nil
		return err
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Streams returns the single mono audio stream this test card provides.
func (t *TestCard) Streams() []profile.Profile {
	return []profile.Profile{t.audio}
}

// NextFrame generates the next chunk of sine wave and returns immediately —
// it does not pace itself to real time. A caller that wants real-time
// delivery (e.g. a live streaming job) is responsible for its own pacing
// based on the returned frame's Pts/SampleRate.
func (t *TestCard) NextFrame(ctx context.Context) (frame.Frame, error) {
	if t.frame == nil {
		return nil, gomedia.ErrInternalError.With("test card is closed")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Reusing the same buffer across calls requires exclusive ownership —
	// the encoder from a prior Encode call may still hold a reference.
	if err := t.frame.MakeWritable(); err != nil {
		return nil, err
	}

	// Advance the timeline. Pts is a running sample count (not tied to any
	// AVFrame.TimeBase, which isn't meaningfully set on a pre-encode raw
	// frame) — elapsed seconds is just Pts/SampleRate.
	if !t.started {
		t.frame.SetPts(0)
		t.started = true
	} else {
		t.frame.IncPts(int64(t.frame.NumSamples()))
	}
	ts := float64(t.frame.Pts()) / float64(t.frame.SampleRate())

	// Generate the sine wave for this chunk
	amplitude := math.Pow(10, t.volume/20.0)
	samples := t.frame.Float32(0)
	sampleRate := float64(t.frame.SampleRate())
	for n := range samples {
		sampleTime := ts + float64(n)/sampleRate
		samples[n] = float32(math.Sin(2.0*math.Pi*t.frequency*sampleTime) * amplitude)
	}

	return t.frame, nil
}
