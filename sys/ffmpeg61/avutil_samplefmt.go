package ffmpeg

import (
	"encoding/json"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/avutil.h>
#include <libavutil/samplefmt.h>
*/
import "C"

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

func (v AVSampleFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Enumerate sample formats
func AVUtil_next_sample_fmt(iterator *uintptr) AVSampleFormat {
	if iterator == nil {
		return AV_SAMPLE_FMT_NONE
	}

	// Increment the iterator
	*iterator += 1

	// Check for end of enumeration
	if AVSampleFormat(*iterator) == AV_SAMPLE_FMT_NB {
		return AV_SAMPLE_FMT_NONE
	}

	// Return the sample format
	return AVSampleFormat(*iterator)
}

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

// Get the packed alternative form of the given sample format.
// If the passed sample_fmt is already in packed format, the format returned is the same as the input.
func AVUtil_get_packed_sample_fmt(sample_fmt AVSampleFormat) AVSampleFormat {
	return AVSampleFormat(C.av_get_packed_sample_fmt(C.enum_AVSampleFormat(sample_fmt)))
}

// Get the planar alternative form of the given sample format.
// If the passed sample_fmt is already in planar format, the format returned is the same as the input.
func AVUtil_get_planar_sample_fmt(sample_fmt AVSampleFormat) AVSampleFormat {
	return AVSampleFormat(C.av_get_planar_sample_fmt(C.enum_AVSampleFormat(sample_fmt)))
}
