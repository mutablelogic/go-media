package ffmpeg

import (
	"testing"
)

func Test_avdevice_version_000(t *testing.T) {
	t.Log("avdevice_version=", AVDevice_version())
}

func Test_avdevice_version_001(t *testing.T) {
	t.Log("avdevice_configuration=", AVDevice_configuration())
}

func Test_avdevice_version_002(t *testing.T) {
	t.Log("avdevice_license=", AVDevice_license())
}
