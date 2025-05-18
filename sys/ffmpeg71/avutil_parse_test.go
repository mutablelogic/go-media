package ffmpeg_test

import (
	"testing"

	// Package imports
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

func Test_avutil_parse_000(t *testing.T) {
	assert := assert.New(t)
	x, y, err := AVUtil_parse_video_size("1920x1080")
	if !assert.NoError(err) {
		t.Fatal(err)
	}
	assert.Equal(1920, x)
	assert.Equal(1080, y)
}
