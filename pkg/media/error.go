package media

import (
	"fmt"
	"strings"

	ffmpeg "github.com/djthorpe/go-media/sys/ffmpeg"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type MediaError struct {
	Level   LogLevel
	Message string
}

type LogLevel ffmpeg.AVLogLevel

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	AV_LOG_PANIC   = LogLevel(ffmpeg.AV_LOG_PANIC)
	AV_LOG_FATAL   = LogLevel(ffmpeg.AV_LOG_FATAL)
	AV_LOG_ERROR   = LogLevel(ffmpeg.AV_LOG_ERROR)
	AV_LOG_WARNING = LogLevel(ffmpeg.AV_LOG_WARNING)
	AV_LOG_INFO    = LogLevel(ffmpeg.AV_LOG_INFO)
	AV_LOG_VERBOSE = LogLevel(ffmpeg.AV_LOG_VERBOSE)
	AV_LOG_DEBUG   = LogLevel(ffmpeg.AV_LOG_DEBUG)
	AV_LOG_TRACE   = LogLevel(ffmpeg.AV_LOG_TRACE)
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewMediaError(level ffmpeg.AVLogLevel, message string) error {
	return MediaError{LogLevel(level), strings.TrimSpace(message)}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (e MediaError) Error() string {
	return fmt.Sprint(ffmpeg.AVLogLevel(e.Level), " ", e.Message)
}
