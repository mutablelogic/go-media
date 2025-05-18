package ffmpeg_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

func Test_avformat_version_000(t *testing.T) {
	t.Log("avformat_version=", AVFormat_version())
}

func Test_avformat_version_001(t *testing.T) {
	t.Log("avformat_configuration=", AVFormat_configuration())
}

func Test_avformat_version_002(t *testing.T) {
	t.Log("avformat_license=", AVFormat_license())
}
