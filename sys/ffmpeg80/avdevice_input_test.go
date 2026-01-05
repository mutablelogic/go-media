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
		if assert.NoError(err) {
			if devices != nil {
				t.Log("  devices=", devices)
				assert.GreaterOrEqual(devices.NumDevices(), 0)
				AVDevice_free_list_devices(devices)
			}
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
		if assert.NoError(err) {
			if devices != nil {
				t.Log("  devices=", devices)
				assert.GreaterOrEqual(devices.NumDevices(), 0)
				AVDevice_free_list_devices(devices)
			}
		}

		input = AVDevice_input_video_device_next(input)
	}

	t.Logf("Found %d video input devices", count)
}
