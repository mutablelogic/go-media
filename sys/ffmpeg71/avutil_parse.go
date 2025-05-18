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
// PUBLIC METHODS

// Parse size and return the width and height of the detected values.
func AVUtil_parse_video_size(size string) (int, int, error) {
	var width, height C.int
	var cStr = C.CString(size)
	defer C.free(unsafe.Pointer(cStr))
	if ret := AVError(C.av_parse_video_size(&width, &height, cStr)); ret < 0 {
		return 0, 0, ret
	}
	return int(width), int(height), nil
}
