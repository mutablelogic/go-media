package ffmpeg

import (
	"testing"
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
