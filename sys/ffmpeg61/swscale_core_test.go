package ffmpeg_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
	"github.com/stretchr/testify/assert"
)

func Test_swscale_core_000(t *testing.T) {
	assert := assert.New(t)
	ctx := SWScale_alloc_context()
	if !assert.NotNil(ctx) {
		t.SkipNow()
	}
	SWScale_free_context(ctx)
}

func Test_swscale_core_001(t *testing.T) {
	assert := assert.New(t)
	ctx := SWScale_get_context(320, 240, AV_PIX_FMT_YUV420P, 640, 480, AV_PIX_FMT_RGB24, SWS_BILINEAR, nil, nil, nil)
	if !assert.NotNil(ctx) {
		t.SkipNow()
	}
	SWScale_free_context(ctx)
}
