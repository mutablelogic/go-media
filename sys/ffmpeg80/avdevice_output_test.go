package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_avdevice_output_000(t *testing.T) {
	assert := assert.New(t)

	output := AVDevice_output_audio_device_first()
	count := 0
	for output != nil {
		count++
		t.Log("audio output=", output)

		devices, err := AVDevice_list_output_sinks(output, "", nil)
		if err == nil && devices != nil {
			t.Log("  devices=", devices)
			assert.GreaterOrEqual(devices.NumDevices(), 0)
			AVDevice_free_list_devices(devices)
		}

		output = AVDevice_output_audio_device_next(output)
	}

	t.Logf("Found %d audio output devices", count)
}

func Test_avdevice_output_001(t *testing.T) {
	assert := assert.New(t)

	output := AVDevice_output_video_device_first()
	count := 0
	for output != nil {
		count++
		t.Log("video output=", output)

		devices, err := AVDevice_list_output_sinks(output, "", nil)
		if err == nil && devices != nil {
			t.Log("  devices=", devices)
			assert.GreaterOrEqual(devices.NumDevices(), 0)
			AVDevice_free_list_devices(devices)
		}

		output = AVDevice_output_video_device_next(output)
	}

	t.Logf("Found %d video output devices", count)
}
