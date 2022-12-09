package ffmpeg

import (
	"fmt"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/log.h>
#include <stdarg.h>
#include <stdlib.h>
#include <stdio.h>
#define MAX_LOG_BUFFER 1024

extern void av_log_cb_(int level,char* message,void* userInfo);

static void av_log_cb(void* userInfo,int level,const char* fmt,va_list args) {
	static char buf[MAX_LOG_BUFFER];
	vsnprintf(buf, MAX_LOG_BUFFER, fmt, args);
	av_log_cb_(level, buf, userInfo);
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
// GLOBALS

var (
	log_callback AVLogCallback
)

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

// Send the specified message to the log if the level is less than or equal to the current av_log_level.
func AVUtil_av_log(class *AVClass, level AVLogLevel, v string, args ...any) {
	cStr := C.CString(fmt.Sprintf(v, args...))
	defer C.free(unsafe.Pointer(cStr))
	C.av_log_(unsafe.Pointer(class), C.int(level), cStr)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

//export av_log_cb_
func av_log_cb_(level C.int, message *C.char, userInfo unsafe.Pointer) {
	if log_callback != nil && message != nil {
		level_ := AVLogLevel(level)
		if level_ <= AVUtil_av_log_get_level() {
			log_callback(level_, C.GoString(message), uintptr(userInfo))
		}
	}
}
