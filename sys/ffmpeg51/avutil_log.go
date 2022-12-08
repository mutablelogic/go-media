package ffmpeg

import (
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
//#include <libavutil/dict.h>
//#include <libavutil/mem.h>
//#include <libavutil/frame.h>
//#include <libavutil/error.h>
#include <stdlib.h>
#define MAX_LOG_BUFFER 1024

extern void av_log_cb_(int level,char* message,void* userInfo);

static void av_log_cb(void* userInfo,int level,const char* fmt,va_list args) {
	static char buf[MAX_LOG_BUFFER];
	vsnprintf(buf,MAX_LOG_BUFFER,fmt,args);
	av_log_cb_(level,buf,userInfo);
}
static void av_log_set_callback_(int def) {
	// true if the default callback should be set
	if (def) {
		av_log_set_callback(av_log_default_callback);
	} else {
		av_log_set_callback(av_log_cb);
	}
}
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type AVLogLevel C.int
type AVLogCallback func(level AVLogLevel, message string, userInfo uintptr)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	log_callback AVLogCallback
)

const (
	AV_LOG_QUIET   AVLogLevel = C.AV_LOG_QUIET
	AV_LOG_PANIC   AVLogLevel = C.AV_LOG_PANIC
	AV_LOG_FATAL   AVLogLevel = C.AV_LOG_FATAL
	AV_LOG_ERROR   AVLogLevel = C.AV_LOG_ERROR
	AV_LOG_WARNING AVLogLevel = C.AV_LOG_WARNING
	AV_LOG_INFO    AVLogLevel = C.AV_LOG_INFO
	AV_LOG_VERBOSE AVLogLevel = C.AV_LOG_VERBOSE
	AV_LOG_DEBUG   AVLogLevel = C.AV_LOG_DEBUG
	AV_LOG_TRACE   AVLogLevel = C.AV_LOG_TRACE
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v AVLogLevel) String() string {
	switch v {
	case AV_LOG_QUIET:
		return "AV_LOG_QUIET"
	case AV_LOG_PANIC:
		return "AV_LOG_PANIC"
	case AV_LOG_FATAL:
		return "AV_LOG_FATAL"
	case AV_LOG_ERROR:
		return "AV_LOG_ERROR"
	case AV_LOG_WARNING:
		return "AV_LOG_WARNING"
	case AV_LOG_INFO:
		return "AV_LOG_INFO"
	case AV_LOG_VERBOSE:
		return "AV_LOG_VERBOSE"
	case AV_LOG_DEBUG:
		return "AV_LOG_DEBUG"
	case AV_LOG_TRACE:
		return "AV_LOG_TRACE"
	default:
		return "[?? Invalid AVLogLevel value]"
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// sets both the callback function and the level of output
// for logging. Where the callback is nil, the default ffmpeg logging is used.
func AVUtil_av_log_set_level(level AVLogLevel, cb AVLogCallback) {
	log_callback = cb
	if cb == nil {
		C.av_log_set_callback_(1)
	} else {
		C.av_log_set_callback_(0)
	}
	C.av_log_set_level(C.int(level))
}

// Get the current log level.
func AVUtil_av_log_get_level() AVLogLevel {
	return AVLogLevel(C.av_log_get_level())
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

//export av_log_cb_
func av_log_cb_(level C.int, message *C.char, userInfo unsafe.Pointer) {
	if log_callback != nil && message != nil {
		level_ := AVLogLevel(level)
		if level_ <= AVGetLogLevel() {
			log_callback(level_, C.GoString(message), uintptr(userInfo))
		}
	}
}
