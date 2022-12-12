package ffmpeg_test

import (
	"testing"

	// Pacakge imports
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
	"github.com/stretchr/testify/assert"
)

const (
	SAMPLE_MP4 = "../../etc/sample.mp4"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_avformat_000(t *testing.T) {
	t.Log("avformat_version=", ffmpeg.AVFormat_version())
}

func Test_avformat_001(t *testing.T) {
	t.Log("avformat_configuration=", ffmpeg.AVFormat_configuration())
}

func Test_avformat_002(t *testing.T) {
	t.Log("avformat_license=", ffmpeg.AVFormat_license())
}

func Test_avformat_003(t *testing.T) {
	var opaque uintptr
	for {
		format := ffmpeg.AVFormat_av_muxer_iterate(&opaque)
		if format == nil {
			break
		}
		t.Log("muxer=", format)
	}
}

func Test_avformat_004(t *testing.T) {
	var opaque uintptr
	for {
		format := ffmpeg.AVFormat_av_demuxer_iterate(&opaque)
		if format == nil {
			break
		}
		t.Log("demuxer=", format)
	}
}

func Test_avformat_005(t *testing.T) {
	assert := assert.New(t)
	var ctx *ffmpeg.AVFormatContext
	var dict *ffmpeg.AVDictionary
	err := ffmpeg.AVFormat_open_input(&ctx, SAMPLE_MP4, nil, &dict)
	assert.NoError(err)
	assert.NotNil(ctx)
	t.Log(ctx, dict)
	ffmpeg.AVFormat_close_input(&ctx)
	assert.Nil(ctx)
}
