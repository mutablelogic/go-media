package ffmpeg_test

import (
	"testing"

	// Packages
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

func Test_avutil_frame_000(t *testing.T) {
	assert := assert.New(t)

	frame := AVUtil_frame_alloc()
	if !assert.NotNil(frame) {
		t.SkipNow()
	}
	AVUtil_frame_free(frame)
}
