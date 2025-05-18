//go:build !container

package ffmpeg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

func Test_avdevice_input_000(t *testing.T) {
	assert := assert.New(t)
	input := AVDevice_input_audio_device_first()
	for {
		if input == nil {
			break
		}
		t.Log("audio input=", input)
		devices, err := AVDevice_list_input_sources(input, "", nil)
		if assert.NoError(err) {
			if devices != nil {
				t.Log("  devices=", devices)
				AVDevice_free_list_devices(devices)
			}
		}

		input = AVDevice_input_audio_device_next(input)

	}
}

func Test_avdevice_input_001(t *testing.T) {
	assert := assert.New(t)
	input := AVDevice_input_video_device_first()
	for {
		if input == nil {
			break
		}
		t.Log("video input=", input)
		devices, err := AVDevice_list_input_sources(input, "", nil)
		if assert.NoError(err) {
			if devices != nil {
				t.Log("  devices=", devices)
				AVDevice_free_list_devices(devices)
			}
		}

		input = AVDevice_input_video_device_next(input)
	}
}
