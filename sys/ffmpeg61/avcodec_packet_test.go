package ffmpeg_test

import (
	"testing"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
	"github.com/stretchr/testify/assert"
)

func Test_avcodec_packet_000(t *testing.T) {
	assert := assert.New(t)
	packet := AVCodec_av_packet_alloc()
	if !assert.NotNil(packet) {
		t.SkipNow()
	}

	if !assert.NoError(AVCodec_av_new_packet(packet, 1024)) {
		t.SkipNow()
	}
	if !assert.NoError(AVCodec_av_grow_packet(packet, 2048)) {
		t.SkipNow()
	}
	AVCodec_av_shrink_packet(packet, 1024)
	AVCodec_av_packet_unref(packet)
	AVCodec_av_packet_free(packet)
}
