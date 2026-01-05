package ffmpeg

import (
	"testing"

)

func Test_swresample_version_000(t *testing.T) {
	t.Log("swresample_version=", SWResample_version())
}

func Test_swresample_version_001(t *testing.T) {
	t.Log("swresample_configuration=", SWResample_configuration())
}

func Test_swresample_version_002(t *testing.T) {
	t.Log("swresample_license=", SWResample_license())
}
