package ffmpeg_test

import (
	"testing"

	// Pacakge imports
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_avdevice_000(t *testing.T) {
	t.Log("avdevice_version=", ffmpeg.AVDevice_version())
}

func Test_avdevice_001(t *testing.T) {
	t.Log("avdevice_configuration=", ffmpeg.AVDevice_configuration())
}

func Test_avdevice_002(t *testing.T) {
	t.Log("avdevice_license=", ffmpeg.AVDevice_license())
}

func Test_avdevice_003(t *testing.T) {
	device := ffmpeg.AVDevice_av_input_audio_device_first()
	for device != nil {
		t.Log("audio input device=", device)
		device = device.AVDevice_av_input_audio_device_next()
	}
}

func Test_avdevice_004(t *testing.T) {
	device := ffmpeg.AVDevice_av_input_video_device_first()
	for device != nil {
		t.Log("video input device=", device)
		device = device.AVDevice_av_input_video_device_next()
	}
}

func Test_avdevice_005(t *testing.T) {
	device := ffmpeg.AVDevice_av_output_audio_device_first()
	for device != nil {
		t.Log("audio output device=", device)
		device = device.AVDevice_av_output_audio_device_next()
	}
}

func Test_avdevice_006(t *testing.T) {
	device := ffmpeg.AVDevice_av_output_video_device_first()
	for device != nil {
		t.Log("video output device=", device)
		device = device.AVDevice_av_output_video_device_next()
	}
}
