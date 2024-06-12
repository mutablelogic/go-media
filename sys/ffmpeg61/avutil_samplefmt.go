package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#include <libavutil/samplefmt.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

func AVUtil_get_sample_fmt_name(samplefmt AVSampleFormat) string {
	return C.GoString(C.av_get_sample_fmt_name(C.enum_AVSampleFormat(samplefmt)))
}
