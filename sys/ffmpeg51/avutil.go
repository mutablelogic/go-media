package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/log.h>
#include <libavutil/dict.h>
#include <libavutil/rational.h>
*/
import "C"

type (
	AVError           C.int
	AVClass           C.struct_AVClass
	AVLogLevel        C.int
	AVLogCallback     func(AVLogLevel, string, uintptr)
	AVDictionaryEntry C.struct_AVDictionaryEntry
	AVDictionaryFlag  int
	AVDictionary      struct {
		ctx *C.struct_AVDictionary
	}
	AVRational C.struct_AVRational
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

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

const (
	AV_DICT_MATCH_CASE      AVDictionaryFlag = C.AV_DICT_MATCH_CASE
	AV_DICT_IGNORE_SUFFIX   AVDictionaryFlag = C.AV_DICT_IGNORE_SUFFIX
	AV_DICT_DONT_STRDUP_KEY AVDictionaryFlag = C.AV_DICT_DONT_STRDUP_KEY
	AV_DICT_DONT_STRDUP_VAL AVDictionaryFlag = C.AV_DICT_DONT_STRDUP_VAL
	AV_DICT_DONT_OVERWRITE  AVDictionaryFlag = C.AV_DICT_DONT_OVERWRITE
	AV_DICT_APPEND          AVDictionaryFlag = C.AV_DICT_APPEND
	AV_DICT_MULTIKEY        AVDictionaryFlag = C.AV_DICT_MULTIKEY
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
