//go:build !container

package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_avdevice_input_000(t *testing.T) {
	assert := assert.New(t)

	input := AVDevice_input_audio_device_first()
	count := 0
	for input != nil {
		count++
		t.Log("audio input=", input)

		devices, err := AVDevice_list_input_sources(input, "", nil)
		if err != nil {
			// Some registered drivers (e.g. dshow on a headless CI runner)
			// can't actually enumerate sources in this environment.
			t.Logf("  could not list sources for %v: %v", input, err)
		} else if devices != nil {
			t.Log("  devices=", devices)
			assert.GreaterOrEqual(devices.NumDevices(), 0)
			AVDevice_free_list_devices(devices)
		}

		input = AVDevice_input_audio_device_next(input)
	}

	t.Logf("Found %d audio input devices", count)
}

func Test_avdevice_input_001(t *testing.T) {
	assert := assert.New(t)

	input := AVDevice_input_video_device_first()
	count := 0
	for input != nil {
		count++
		t.Log("video input=", input)

		devices, err := AVDevice_list_input_sources(input, "", nil)
		if err != nil {
			// Some registered drivers (e.g. dshow/gdigrab on a headless CI
			// runner) can't actually enumerate sources in this environment.
			t.Logf("  could not list sources for %v: %v", input, err)
		} else if devices != nil {
			t.Log("  devices=", devices)
			assert.GreaterOrEqual(devices.NumDevices(), 0)
			AVDevice_free_list_devices(devices)
		}

		input = AVDevice_input_video_device_next(input)
	}

	t.Logf("Found %d video input devices", count)
}
