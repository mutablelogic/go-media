//go:build sdl2

package sdl

import (
	"errors"
	"fmt"

	// Packages
	"github.com/veandco/go-sdl2/sdl"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Audio represents an SDL audio device for playback.
type Audio struct {
	device sdl.AudioDeviceID
	spec   *sdl.AudioSpec
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewAudio creates a new SDL audio device with the specified parameters.
// sampleRate is in Hz (e.g., 44100, 48000)
// channels is the number of audio channels (1 for mono, 2 for stereo)
// samples is the audio buffer size in samples (power of 2, e.g., 4096)
func (c *Context) NewAudio(sampleRate int32, channels uint8, samples uint16) (*Audio, error) {
	spec := &sdl.AudioSpec{
		Freq:     sampleRate,
		Format:   sdl.AUDIO_F32, // 32-bit float
		Channels: channels,
		Samples:  samples,
	}

	device, err := sdl.OpenAudioDevice("", false, spec, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("sdl.OpenAudioDevice: %w", err)
	}

	return &Audio{
		device: device,
		spec:   spec,
	}, nil
}

// Close closes the audio device.
func (a *Audio) Close() error {
	if a.device == 0 {
		return nil
	}
	sdl.CloseAudioDevice(a.device)
	a.device = 0
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Queue queues audio data for playback.
// The data should be in the format specified when creating the device (float32).
func (a *Audio) Queue(data []byte) error {
	if a.device == 0 {
		return errors.New("audio device not open")
	}

	if err := sdl.QueueAudio(a.device, data); err != nil {
		return fmt.Errorf("sdl.QueueAudio: %w", err)
	}

	return nil
}

// Pause pauses audio playback.
func (a *Audio) Pause() error {
	if a.device == 0 {
		return errors.New("audio device not open")
	}
	sdl.PauseAudioDevice(a.device, true)
	return nil
}

// Resume resumes audio playback.
func (a *Audio) Resume() error {
	if a.device == 0 {
		return errors.New("audio device not open")
	}
	sdl.PauseAudioDevice(a.device, false)
	return nil
}

// QueuedSize returns the number of bytes currently queued for playback.
func (a *Audio) QueuedSize() uint32 {
	if a.device == 0 {
		return 0
	}
	return sdl.GetQueuedAudioSize(a.device)
}

// Clear clears any queued audio data.
func (a *Audio) Clear() error {
	if a.device == 0 {
		return errors.New("audio device not open")
	}
	sdl.ClearQueuedAudio(a.device)
	return nil
}

// Spec returns the audio device specification.
func (a *Audio) Spec() *sdl.AudioSpec {
	return a.spec
}
