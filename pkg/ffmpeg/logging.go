package ffmpeg

import (
	"fmt"
	"strings"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Logging function
type LogFn func(text string)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set logging options, including a callback function
func SetLogging(verbose bool, fn LogFn) {
	ff.AVUtil_log_set_level(ff.AV_LOG_INFO)
	if !verbose {
		ff.AVUtil_log_set_level(ff.AV_LOG_ERROR)
	}
	if fn != nil {
		ff.AVUtil_log_set_callback(func(level ff.AVLog, message string, userInfo any) {
			fn(fmt.Sprintf("[%v] %v", level, strings.TrimSpace(message)))
		})
	} else {
		ff.AVUtil_log_set_callback(nil)
	}
}
