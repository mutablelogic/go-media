package ffmpeg

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#include <libavutil/parseutils.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - VIDEO SIZE

// AVUtil_parse_video_size parses a string representing a video size and returns
// the width and height. The string can be in the format "WIDTHxHEIGHT" (e.g., "1920x1080")
// or a named resolution (e.g., "vga", "hd720", "hd1080", "4k").
//
// Returns width, height, and error if parsing fails.
func AVUtil_parse_video_size(size string) (int, int, error) {
	var width, height C.int
	var cStr = C.CString(size)
	defer C.free(unsafe.Pointer(cStr))
	if ret := AVError(C.av_parse_video_size(&width, &height, cStr)); ret < 0 {
		return 0, 0, ret
	}
	return int(width), int(height), nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - VIDEO RATE

// AVUtil_parse_video_rate parses a string representing a video frame rate and returns
// an AVRational. The string can be in formats like "25", "30000/1001", "29.97", etc.
//
// Returns AVRational and error if parsing fails.
func AVUtil_parse_video_rate(rate string) (AVRational, error) {
	var avrate C.AVRational
	var cStr = C.CString(rate)
	defer C.free(unsafe.Pointer(cStr))
	if ret := AVError(C.av_parse_video_rate(&avrate, cStr)); ret < 0 {
		return AVRational{}, ret
	}
	return AVRational(avrate), nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - TIME

// AVUtil_parse_time parses a time string and returns the time in microseconds.
// The time string can be in various formats:
//   - [-][HH:]MM:SS[.m...]
//   - [-]S+[.m...]
//
// If duration is true, the parsing is for a duration (can be negative).
//
// Returns time in microseconds and error if parsing fails.
func AVUtil_parse_time(timestr string, duration bool) (int64, error) {
	var timeval C.int64_t
	var cStr = C.CString(timestr)
	defer C.free(unsafe.Pointer(cStr))
	var durFlag C.int
	if duration {
		durFlag = 1
	}
	if ret := AVError(C.av_parse_time(&timeval, cStr, durFlag)); ret < 0 {
		return 0, ret
	}
	return int64(timeval), nil
}
