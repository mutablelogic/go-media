package ffmpeg_test

import (
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

func Test_avformat_input_001(t *testing.T) {
	assert := assert.New(t)
	// Iterate over all input formats
	var opaque uintptr
	for {
		demuxer := AVFormat_demuxer_iterate(&opaque)
		if demuxer == nil {
			break
		}
		demuxer2 := AVFormat_find_input_format(demuxer.Name())
		assert.Equal(demuxer, demuxer2)
	}
}
