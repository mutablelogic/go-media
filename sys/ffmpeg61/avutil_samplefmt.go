package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#include <libavutil/samplefmt.h>
*/
import "C"
import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AV_SAMPLE_FMT_NONE AVSampleFormat = C.AV_SAMPLE_FMT_NONE
	AV_SAMPLE_FMT_U8   AVSampleFormat = C.AV_SAMPLE_FMT_U8
	AV_SAMPLE_FMT_S16  AVSampleFormat = C.AV_SAMPLE_FMT_S16
	AV_SAMPLE_FMT_S32  AVSampleFormat = C.AV_SAMPLE_FMT_S32
	AV_SAMPLE_FMT_FLT  AVSampleFormat = C.AV_SAMPLE_FMT_FLT
	AV_SAMPLE_FMT_DBL  AVSampleFormat = C.AV_SAMPLE_FMT_DBL
	AV_SAMPLE_FMT_U8P  AVSampleFormat = C.AV_SAMPLE_FMT_U8P
	AV_SAMPLE_FMT_S16P AVSampleFormat = C.AV_SAMPLE_FMT_S16P
	AV_SAMPLE_FMT_S32P AVSampleFormat = C.AV_SAMPLE_FMT_S32P
	AV_SAMPLE_FMT_FLTP AVSampleFormat = C.AV_SAMPLE_FMT_FLTP
	AV_SAMPLE_FMT_DBLP AVSampleFormat = C.AV_SAMPLE_FMT_DBLP
	AV_SAMPLE_FMT_S64  AVSampleFormat = C.AV_SAMPLE_FMT_S64
	AV_SAMPLE_FMT_S64P AVSampleFormat = C.AV_SAMPLE_FMT_S64P
	AV_SAMPLE_FMT_NB   AVSampleFormat = C.AV_SAMPLE_FMT_NB
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v AVSampleFormat) String() string {
	return AVUtil_get_sample_fmt_name(v)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Return the name of sample_fmt, or empty string if sample_fmt is not recognized
func AVUtil_get_sample_fmt_name(sample_fmt AVSampleFormat) string {
	return C.GoString(C.av_get_sample_fmt_name(C.enum_AVSampleFormat(sample_fmt)))
}

// Return a sample format corresponding to name, or AV_SAMPLE_FMT_NONE on error.
func AVUtil_get_sample_fmt(name string) AVSampleFormat {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return AVSampleFormat(C.av_get_sample_fmt(cName))
}

// Return number of bytes per sample.
func AVUtil_get_bytes_per_sample(sample_fmt AVSampleFormat) int {
	return int(C.av_get_bytes_per_sample(C.enum_AVSampleFormat(sample_fmt)))
}

// Check if the sample format is planar.
func AVUtil_sample_fmt_is_planar(sample_fmt AVSampleFormat) bool {
	return C.av_sample_fmt_is_planar(C.enum_AVSampleFormat(sample_fmt)) != 0
}

// Get the required buffer size for the given audio parameters,
// include the number of samples in a single channel.
func AVUtil_samples_get_buffer_size(sample_fmt AVSampleFormat, nb_channels int, nb_samples int, align bool) int {
	return int(C.av_samples_get_buffer_size(nil, C.int(nb_channels), C.int(nb_samples), C.enum_AVSampleFormat(sample_fmt), boolToInt(align)))
}
