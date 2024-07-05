package ffmpeg_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

func Test_avcodec_packet_000(t *testing.T) {
	assert := assert.New(t)
	packet := AVCodec_packet_alloc()
	if !assert.NotNil(packet) {
		t.SkipNow()
	}

	if !assert.NoError(AVCodec_new_packet(packet, 1024)) {
		t.SkipNow()
	}
	if !assert.NoError(AVCodec_grow_packet(packet, 2048)) {
		t.SkipNow()
	}
	AVCodec_shrink_packet(packet, 1024)
	AVCodec_packet_unref(packet)
	AVCodec_packet_free(packet)
}
