package ffmpeg_test

import (
	"testing"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

func Test_logging_001(t *testing.T) {
	// Set logging
	ffmpeg.SetLogging(true, func(v string) {
		t.Log(v)
	})

	ff.AVUtil_log(nil, ff.AV_LOG_INFO, "INFO test")
	ff.AVUtil_log(nil, ff.AV_LOG_WARNING, "WARN test")
	ff.AVUtil_log(nil, ff.AV_LOG_ERROR, "ERROR test")
}

func Test_logging_002(t *testing.T) {
	// Set logging
	ffmpeg.SetLogging(false, func(v string) {
		t.Log(v)
	})

	ff.AVUtil_log(nil, ff.AV_LOG_INFO, "INFO test")
	ff.AVUtil_log(nil, ff.AV_LOG_WARNING, "WARN test")
	ff.AVUtil_log(nil, ff.AV_LOG_ERROR, "ERROR test")
}
