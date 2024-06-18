package ffmpeg_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

func Test_avdevice_output_000(t *testing.T) {
	output := AVDevice_output_audio_device_first()
	for {
		if output == nil {
			break
		}
		t.Log("audio output=", output)
		devices, err := AVDevice_list_output_sinks(output, "", nil)
		if err == nil {
			t.Log("  devices=", devices)
		}
		AVDevice_free_list_devices(devices)

		output = AVDevice_output_audio_device_next(output)

	}
}

func Test_avdevice_output_001(t *testing.T) {
	output := AVDevice_output_video_device_first()
	for {
		if output == nil {
			break
		}
		t.Log("video output=", output)
		devices, err := AVDevice_list_output_sinks(output, "", nil)
		if err == nil {
			t.Log("  devices=", devices)
		}
		AVDevice_free_list_devices(devices)

		output = AVDevice_output_video_device_next(output)
	}
}
