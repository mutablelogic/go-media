package ffmpeg_test

import (
	"testing"

	// Package imports
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
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
