package ffmpeg_test

import (
	"testing"

	// Package imports
	"github.com/stretchr/testify/assert"

	// Namespace imports
	. "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

func Test_avutil_log_000(t *testing.T) {
	assert := assert.New(t)

	// Set log level
	AVUtil_log_set_level(AV_LOG_TRACE)
	assert.Equal(AV_LOG_TRACE, AVUtil_log_get_level())

	// Log a message
	AVUtil_log(nil, AV_LOG_TRACE, "This is a trace message")
	AVUtil_log(nil, AV_LOG_DEBUG, "This is a debug message")
	AVUtil_log(nil, AV_LOG_VERBOSE, "This is a verbose message")
	AVUtil_log(nil, AV_LOG_INFO, "This is a info message")
	AVUtil_log(nil, AV_LOG_WARNING, "This is a warning message")
	AVUtil_log(nil, AV_LOG_ERROR, "This is a error message")
	AVUtil_log(nil, AV_LOG_FATAL, "This is a fatal message")
	AVUtil_log(nil, AV_LOG_PANIC, "This is a panic message")
}

func Test_avutil_log_001(t *testing.T) {
	assert := assert.New(t)

	// Set log level
	AVUtil_log_set_level(AV_LOG_ERROR)
	assert.Equal(AV_LOG_ERROR, AVUtil_log_get_level())

	// Set log callback
	AVUtil_log_set_callback(func(level AVLog, message string, userInfo any) {
		t.Logf("Level=%v, Message=%v userInfo=%v", level, message, userInfo)
	})

	// Log a message
	AVUtil_log(nil, AV_LOG_TRACE, "This is a trace message")
	AVUtil_log(nil, AV_LOG_DEBUG, "This is a debug message")
	AVUtil_log(nil, AV_LOG_VERBOSE, "This is a verbose message")
	AVUtil_log(nil, AV_LOG_INFO, "This is a info message")
	AVUtil_log(nil, AV_LOG_WARNING, "This is a warning message")
	AVUtil_log(nil, AV_LOG_ERROR, "This is a error message")
	AVUtil_log(nil, AV_LOG_FATAL, "This is a fatal message")
	AVUtil_log(nil, AV_LOG_PANIC, "This is a panic message")
}
