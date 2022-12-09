package ffmpeg

import (
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libswresample
#include <libswresample/swresample.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	SWRContext C.struct_SwrContext
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *SWRContext) String() string {
	str := "<SWRContext"
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - VERSION

// Return the LIBSWRESAMPLE_VERSION_INT constant.
func SWR_version() uint {
	return uint(C.swresample_version())
}

// Return the swr build-time configuration.
func SWR_configuration() string {
	return C.GoString(C.swresample_configuration())
}

// Return the swr license.
func SWR_license() string {
	return C.GoString(C.swresample_license())
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - INIT

// Allocate SwrContext.
func SWR_alloc() *SWRContext {
	return (*SWRContext)(C.swr_alloc())
}

// Free the given SwrContext.
func (ctx *SWRContext) SWR_free() {
	C.swr_free((**C.struct_SwrContext)(unsafe.Pointer(&ctx)))
}

// Initialize context after user parameters have been set.
func (ctx *SWRContext) SWR_init() error {
	if err := AVError(C.swr_init((*C.struct_SwrContext)(ctx))); err == 0 {
		return nil
	} else {
		return err
	}
}

// Closes the context so that swr_is_initialized() returns 0
func (ctx *SWRContext) SWR_close() {
	C.swr_close((*C.struct_SwrContext)(ctx))
}

// Check whether an swr context has been initialized or not.
func (ctx *SWRContext) SWR_is_initialized() bool {
	return C.swr_is_initialized((*C.struct_SwrContext)(ctx)) != 0
}

// Set/reset common parameters.
func (ctx *SWRContext) SWR_alloc_set_opts2(out_ch_layout *AVChannelLayout, out_sample_fmt AVSampleFormat, out_sample_rate int, in_ch_layout *AVChannelLayout, in_sample_fmt AVSampleFormat, in_sample_rate int, log_offset AVLogLevel, log_context *AVClass) error {
	ctx_ := (*C.struct_SwrContext)(ctx)
	if err := AVError(C.swr_alloc_set_opts2(&ctx_, (*C.struct_AVChannelLayout)(out_ch_layout), C.enum_AVSampleFormat(out_sample_fmt), C.int(out_sample_rate), (*C.struct_AVChannelLayout)(in_ch_layout), C.enum_AVSampleFormat(in_sample_fmt), C.int(in_sample_rate), C.int(log_offset), unsafe.Pointer(log_context))); err == 0 {
		return nil
	} else {
		return err
	}
}

// Core conversion functions. Returns number of samples output per channel.
// in and in_count can be set to 0 to flush the last few samples out at the end.
func (ctx *SWRContext) SWR_convert(out **byte, out_count int, in **byte, in_count int) (int, error) {
	n := int(C.swr_convert((*C.struct_SwrContext)(ctx), (**C.uint8_t)(unsafe.Pointer(out)), C.int(out_count), (**C.uint8_t)(unsafe.Pointer(in)), C.int(in_count)))
	if n < 0 {
		return n, AVError(AVERROR_INVALIDDATA)
	} else {
		return n, nil
	}
}

// Core conversion functions. Returns number of samples output per channel.
// in can be set to nil to flush the last few samples out at the end.
func (ctx *SWRContext) SWR_convert_bytes(out, in []byte) (int, error) {
	var pOut *byte
	var pIn *byte
	if out != nil {
		pOut = &out[0]
	}
	if in != nil {
		pIn = &in[0]
	}
	return ctx.SWR_convert(&pOut, len(out), &pIn, len(in))
}

// Convert the next timestamp from input to output timestamps are in 1/(in_sample_rate * out_sample_rate) units.
func (ctx *SWRContext) SWR_next_pts(pts int64) int64 {
	return int64(C.swr_next_pts((*C.struct_SwrContext)(ctx), C.int64_t(pts)))
}

// Drops the specified number of output samples.
func (ctx *SWRContext) SWR_drop_output(count int) error {
	if err := AVError(C.swr_drop_output((*C.struct_SwrContext)(ctx), C.int(count))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Inject the specified number of silence samples.
func (ctx *SWRContext) SWR_inject_silence(count int) error {
	if err := AVError(C.swr_inject_silence((*C.struct_SwrContext)(ctx), C.int(count))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Gets the delay the next input sample will experience relative to the next output sample.
func (ctx *SWRContext) SWR_get_delay(base int64) int64 {
	return int64(C.swr_get_delay((*C.struct_SwrContext)(ctx), C.int64_t(base)))
}

// Find an upper bound on the number of samples that the next swr_convert
func (ctx *SWRContext) SWR_get_out_samples(in_samples int) (int, error) {
	n := int(C.swr_get_out_samples((*C.struct_SwrContext)(ctx), C.int(in_samples)))
	if n < 0 {
		return n, AVError(n)
	} else {
		return n, nil
	}
}
