package ffmpeg

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libswresample libavutil
#include <libswresample/swresample.h>
#include <libavutil/avutil.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Set common parameters for resampling.
func SWResample_set_opts(ctx *SWRContext, out_ch_layout AVChannelLayout, out_sample_fmt AVSampleFormat, out_sample_rate int, in_ch_layout AVChannelLayout, in_sample_fmt AVSampleFormat, in_sample_rate int) error {
	if err := AVError(C.swr_alloc_set_opts2((**C.struct_SwrContext)(unsafe.Pointer(&ctx)),
		(*C.struct_AVChannelLayout)(&out_ch_layout),
		C.enum_AVSampleFormat(out_sample_fmt),
		C.int(out_sample_rate),
		(*C.struct_AVChannelLayout)(&in_ch_layout),
		C.enum_AVSampleFormat(in_sample_fmt),
		C.int(in_sample_rate),
		0, nil)); err < 0 {
		return err
	}

	// Return success
	return nil
}
