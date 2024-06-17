package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libswresample
#include <libswresample/swresample.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Core conversion function. Returns number of samples output per channel.
// in and in_count can be set to 0 to flush the last few samples out at the end.
func SWResample_convert(ctx *SWRContext, dst *AVSamples, dst_nb_samples int, src *AVSamples, src_nb_samples int) (int, error) {
	n := int(C.swr_convert(
		(*C.struct_SwrContext)(ctx),
		&dst.planes[0],
		C.int(dst_nb_samples),
		&src.planes[0],
		C.int(src_nb_samples),
	))
	if n < 0 {
		return 0, AVError(n)
	} else {
		return n, nil
	}
}

// Convert the next timestamp from input to output timestamps are in 1/(in_sample_rate * out_sample_rate) units.
func SWResample_next_pts(ctx *SWRContext, pts int64) int64 {
	return int64(C.swr_next_pts((*C.struct_SwrContext)(ctx), C.int64_t(pts)))
}

// Drops the specified number of output samples.
func SWResample_drop_output(ctx *SWRContext, count int) error {
	if err := AVError(C.swr_drop_output((*C.struct_SwrContext)(ctx), C.int(count))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Inject the specified number of silence samples.
func SWResample_inject_silence(ctx *SWRContext, count int) error {
	if err := AVError(C.swr_inject_silence((*C.struct_SwrContext)(ctx), C.int(count))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Gets the delay the next input sample will experience relative to the next output sample.
func SWResample_get_delay(ctx *SWRContext, base int64) int64 {
	return int64(C.swr_get_delay((*C.struct_SwrContext)(ctx), C.int64_t(base)))
}

// Find an upper bound on the number of samples that the next swr_convert call will output, if called with in_samples of input samples.
func SWResample_get_out_samples(ctx *SWRContext, in_samples int) (int, error) {
	n := int(C.swr_get_out_samples((*C.struct_SwrContext)(ctx), C.int(in_samples)))
	if n < 0 {
		return n, AVError(n)
	} else {
		return n, nil
	}
}

// Convert the samples in the input AVFrame and write them to the output AVFrame.
func SWResample_convert_frame(ctx *SWRContext, src, dest *AVFrame) error {
	if err := AVError(C.swr_convert_frame((*C.struct_SwrContext)(ctx), (*C.struct_AVFrame)(dest), (*C.struct_AVFrame)(src))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Configure or reconfigure the SwrContext using the information provided by the AVFrames.
func SWResample_config_frame(ctx *SWRContext, src, dest *AVFrame) error {
	if err := AVError(C.swr_config_frame((*C.struct_SwrContext)(ctx), (*C.struct_AVFrame)(dest), (*C.struct_AVFrame)(src))); err != 0 {
		return err
	} else {
		return nil
	}
}
