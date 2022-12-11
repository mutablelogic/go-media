package ffmpeg_test

import (
	"testing"

	// Package imports
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
	assert "github.com/stretchr/testify/assert"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_avcodec_000(t *testing.T) {
	t.Log("avcodec_version=", ffmpeg.AVCodec_version())
}

func Test_avcodec_001(t *testing.T) {
	t.Log("avcodec_configuration=", ffmpeg.AVCodec_configuration())
}

func Test_avcodec_002(t *testing.T) {
	t.Log("avcodec_license=", ffmpeg.AVCodec_license())
}

func Test_avcodec_003(t *testing.T) {
	var opaque uintptr
	for {
		codec := ffmpeg.AVCodec_iterate(&opaque)
		if codec == nil {
			break
		}
		t.Log("codec=", codec)
	}
}

func Test_avcodec_004(t *testing.T) {
	assert := assert.New(t)
	p := ffmpeg.AVCodec_av_packet_alloc()
	assert.NotNil(p)
	ffmpeg.AVCodec_av_grow_packet(p, 1024)
	assert.Equal(1024, p.Size())
	t.Log(p)
	buf := p.Bytes()
	assert.NotNil(buf)
	assert.Equal(1024, len(buf))
	ffmpeg.AVCodec_av_shrink_packet(p, 512)
	t.Log(p)
	assert.Equal(512, p.Size())
	buf = p.Bytes()
	assert.NotNil(buf)
	assert.Equal(512, len(buf))
	ffmpeg.AVCodec_av_packet_free(&p)
	assert.Nil(p)
}
