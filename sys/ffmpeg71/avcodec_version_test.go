package ffmpeg_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

func Test_avcodec_version_000(t *testing.T) {
	t.Log("avcodec_version=", AVCodec_version())
}

func Test_avcodec_version_001(t *testing.T) {
	t.Log("avcodec_configuration=", AVCodec_configuration())
}

func Test_avcodec_version_002(t *testing.T) {
	t.Log("avcodec_license=", AVCodec_license())
}
