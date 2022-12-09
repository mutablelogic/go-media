package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/samplefmt.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the name of sample_fmt, or empty string if v is not recognized.
func AVUtil_av_get_sample_fmt_name(v AVSampleFormat) string {
	return C.GoString(C.av_get_sample_fmt_name(C.enum_AVSampleFormat(v)))
}

// Return a sample format corresponding to name, or AV_SAMPLE_FMT_NONE on error.
func AVUtil_av_get_sample_fmt(s string) AVSampleFormat {
	cStr := C.CString(s)
	defer C.free(unsafe.Pointer(cStr))
	return AVSampleFormat(C.av_get_sample_fmt(cStr))
}

// Return the planar or packed audio corresponding to f. Returns AV_SAMPLE_FMT_NONE on error.
func AVUtil_av_get_alt_sample_fmt(f AVSampleFormat, planar bool) AVSampleFormat {
	return AVSampleFormat(C.av_get_alt_sample_fmt(C.enum_AVSampleFormat(f), C.int(boolToInt(planar))))
}

// Return the packed audio corresponding to f. Returns AV_SAMPLE_FMT_NONE on error.
func AVUtil_av_get_packed_sample_fmt(f AVSampleFormat) AVSampleFormat {
	return AVSampleFormat(C.av_get_packed_sample_fmt(C.enum_AVSampleFormat(f)))
}

// Return the planar audio corresponding to f. Returns AV_SAMPLE_FMT_NONE on error.
func AVUtil_av_get_planar_sample_fmt(f AVSampleFormat) AVSampleFormat {
	return AVSampleFormat(C.av_get_planar_sample_fmt(C.enum_AVSampleFormat(f)))
}

// Return number of bytes per sample or zero if unknown for the given sample format
func AVUtil_av_get_bytes_per_sample(f AVSampleFormat) int {
	return int(C.av_get_bytes_per_sample(C.enum_AVSampleFormat(f)))
}

// Check if the sample format is planar.
func AVUtil_av_sample_fmt_is_planar(f AVSampleFormat) bool {
	return intToBool(int(C.av_sample_fmt_is_planar(C.enum_AVSampleFormat(f))))
}

// Get the required buffer size for the given audio parameters.
// When align is 1 no alignment is done
func AVUtil_av_samples_get_buffer_size(linesize *int, nb_channels int, nb_samples int, sample_fmt AVSampleFormat, align int) int {
	return int(C.av_samples_get_buffer_size((*C.int)(unsafe.Pointer(linesize)), C.int(nb_channels), C.int(nb_samples), C.enum_AVSampleFormat(sample_fmt), C.int(align)))
}

// Allocate a samples buffer for nb_samples samples, and fill data pointers and linesize accordingly.
// Allocated data will be initialized to silence.
// When align is 1 no alignment is done
func AVUtil_av_samples_alloc(data **uint8, linesize *int, nb_channels int, nb_samples int, sample_fmt AVSampleFormat, align int) int {
	return int(C.av_samples_alloc((**C.uint8_t)(unsafe.Pointer(data)), (*C.int)(unsafe.Pointer(linesize)), C.int(nb_channels), C.int(nb_samples), C.enum_AVSampleFormat(sample_fmt), C.int(align)))
}
