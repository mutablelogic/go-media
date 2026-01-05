package ffmpeg

import (
	"testing"

)

func Test_avutil_version_000(t *testing.T) {
	t.Log("avutil_version=", AVUtil_version())
}

func Test_avutil_version_001(t *testing.T) {
	t.Log("avutil_configuration=", AVUtil_configuration())
}

func Test_avutil_version_002(t *testing.T) {
	t.Log("avutil_license=", AVUtil_license())
}
