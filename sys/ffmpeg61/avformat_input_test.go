package ffmpeg_test

import (
	"testing"

	// Packages

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

func Test_avformat_input_001(t *testing.T) {
	// Iterate over all input formats
	var opaque uintptr
	for {
		demuxer := AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}

		t.Log(demuxer)
	}
}
