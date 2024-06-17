package ffmpeg_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
	"github.com/stretchr/testify/assert"
)

func Test_avutil_channel_layout_001(t *testing.T) {
	assert := assert.New(t)
	var iter uintptr
	for {
		layout := AVUtil_channel_layout_standard(&iter)
		if layout == nil {
			break
		}
		description, err := AVUtil_channel_layout_describe(layout)
		assert.NoError(err)

		t.Logf("AVChannelLayout: %q", description)
		t.Log("  .channels: ", AVUtil_get_channel_layout_nb_channels(layout))
	}
}
