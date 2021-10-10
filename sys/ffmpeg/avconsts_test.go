package ffmpeg_test

import (
	"testing"

	. "github.com/mutablelogic/go-media/sys/ffmpeg"
)

func Test_Consts_001(t *testing.T) {
	for v := AV_PIX_FMT_MIN; v <= AV_PIX_FMT_MAX; v++ {
		t.Log(v)
	}
}
