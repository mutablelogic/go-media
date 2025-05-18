package ffmpeg_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

func Test_swscale_version_000(t *testing.T) {
	t.Log("swscale_version=", SWScale_version())
}

func Test_swscale_version_001(t *testing.T) {
	t.Log("swscale_configuration=", SWScale_configuration())
}

func Test_swscale_version_002(t *testing.T) {
	t.Log("swscale_license=", SWScale_license())
}

func Test_swscale_version_003(t *testing.T) {
	t.Log("swscale_class=", SWScale_get_class())
}
