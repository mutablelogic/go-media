package ffmpeg_test

import (
	"testing"

	// Package imports
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

func Test_avutil_image_000(t *testing.T) {
	assert := assert.New(t)

	data, linesize, bufsize, err := AVUtil_image_alloc(320, 240, AV_PIX_FMT_YUV420P, 16)
	if !assert.NoError(err) {
		t.Fatal(err)
	}
	assert.NotNil(data)
	assert.NotNil(linesize)
	assert.NotZero(bufsize)

	t.Log("data=", data)
	t.Log("linesize=", linesize)
	t.Log("bufsize=", bufsize)

	AVUtil_image_free(data)
}

func Test_avutil_image_001(t *testing.T) {
	assert := assert.New(t)

	sizes, err := AVUtil_image_plane_sizes_ex(320, 240, AV_PIX_FMT_YUV420P)
	if !assert.NoError(err) {
		t.Fatal(err)
	}
	t.Log(sizes)
}
