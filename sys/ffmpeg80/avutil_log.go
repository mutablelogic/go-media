package ffmpeg

import (
	"encoding/json"
	"fmt"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#define MAX_LOG_BUFFER 1024

extern void av_log_cb_(int level,char* message,void* userInfo);

static inline void av_log_cb(void* userInfo,int level,const char* fmt,va_list args) {
	static char buf[MAX_LOG_BUFFER];
	if (level <= av_log_get_level()) {
		vsnprintf(buf, MAX_LOG_BUFFER, fmt, args);
		av_log_cb_(level, buf, userInfo);
	}
}

static void av_log_set_callback_(int def) {
	// true if the default callback should be set
	if (def) {
		av_log_set_callback(av_log_default_callback);
	} else {
		av_log_set_callback(av_log_cb);
	}
}

static void av_log_(void* class, int level, const char* fmt) {
	av_log(class, level, "%s", fmt);
}
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type AVLogFunc func(level AVLog, message string, userInfo any)

type (
	AVLog   C.int
	AVClass C.struct_AVClass
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_LOG_QUIET   AVLog = -8 // C.AV_LOG_QUIET
	AV_LOG_PANIC   AVLog = 0  // C.AV_LOG_PANIC
	AV_LOG_FATAL   AVLog = 8  // C.AV_LOG_FATAL
	AV_LOG_ERROR   AVLog = 16 // C.AV_LOG_ERROR
	AV_LOG_WARNING AVLog = 24 // C.AV_LOG_WARNING
	AV_LOG_INFO    AVLog = 32 // C.AV_LOG_INFO
	AV_LOG_VERBOSE AVLog = 40 // C.AV_LOG_VERBOSE
	AV_LOG_DEBUG   AVLog = 48 // C.AV_LOG_DEBUG
	AV_LOG_TRACE   AVLog = 56 // C.AV_LOG_TRACE
)

var cbLog AVLogFunc

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v AVLog) String() string {
	switch v {
	case AV_LOG_QUIET:
		return "QUIET"
	case AV_LOG_PANIC:
		return "PANIC"
	case AV_LOG_FATAL:
		return "FATAL"
	case AV_LOG_ERROR:
		return "ERROR"
	case AV_LOG_WARNING:
		return "WARN"
	case AV_LOG_INFO:
		return "INFO"
	case AV_LOG_VERBOSE:
		return "VERBOSE"
	case AV_LOG_DEBUG:
		return "DEBUG"
	case AV_LOG_TRACE:
		return "TRACE"
	default:
		return fmt.Sprintf("[AVLog:%d]", v)
	}
}

// MarshalJSON implements the json.Marshaler interface
func (v AVLog) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

func AVUtil_log_set_level(level AVLog) {
	C.av_log_set_level(C.int(level))
}

func AVUtil_log_get_level() AVLog {
	return AVLog(C.av_log_get_level())
}

// Send the specified message to the log if the level is less than or equal to the
// current av_log_level.
func AVUtil_log(class *AVClass, level AVLog, v string, args ...any) {
	cStr := C.CString(fmt.Sprintf(v, args...))
	defer C.free(unsafe.Pointer(cStr))
	C.av_log_(unsafe.Pointer(class), C.int(level), cStr)
}

// Set callback for logging. If cb is nil, the default callback will be set.
func AVUtil_log_set_callback(cb AVLogFunc) {
	if cb == nil {
		C.av_log_set_callback_(1)
		cbLog = nil
	} else {
		C.av_log_set_callback_(0)
		cbLog = cb
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

//export av_log_cb_
func av_log_cb_(level C.int, message *C.char, userInfo unsafe.Pointer) {
	if cbLog != nil {
		cbLog(AVLog(level), C.GoString(message), userInfo)
	}
}
