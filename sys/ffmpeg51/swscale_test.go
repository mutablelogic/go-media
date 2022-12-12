package ffmpeg_test

import (
	"testing"

	// Package imports
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"
)

////////////////////////////////////////////////////////////////////////////////
// TESTS

func Test_sws_000(t *testing.T) {
	t.Log("SWS_version=", ffmpeg.SWS_version())
}

func Test_sws_001(t *testing.T) {
	t.Log("SWS_configuration=", ffmpeg.SWS_configuration())
}

func Test_sws_002(t *testing.T) {
	t.Log("SWS_license=", ffmpeg.SWS_license())
}
