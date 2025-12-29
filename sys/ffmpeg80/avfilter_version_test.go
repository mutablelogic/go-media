package ffmpeg

import (
	"testing"
)

func Test_avfilter_version_000(t *testing.T) {
	t.Log("avfilter_version=", AVFilter_version())
}

func Test_avfilter_version_001(t *testing.T) {
	t.Log("avfilter_configuration=", AVFilter_configuration())
}

func Test_avfilter_version_002(t *testing.T) {
	t.Log("avfilter_license=", AVFilter_license())
}
