package ffmpeg_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

func Test_avformat_output_001(t *testing.T) {
	// Iterate over all output formats
	var opaque uintptr
	for {
		muxer := AVFormat_muxer_iterate(&opaque)
		if muxer == nil {
			break
		}

		t.Log(muxer)
	}
}
