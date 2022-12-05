package ffmpeg_test

import (
	"testing"

	// Pacakge imports
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
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
