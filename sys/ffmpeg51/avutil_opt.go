package ffmpeg

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/opt.h>
#include <stdlib.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (ctx *SWRContext) AVUtil_av_opt_set(name string, value string) error {
	if err := AVError(C.av_opt_set(unsafe.Pointer(ctx), C.CString(name), C.CString(value), 0)); err != 0 {
		return nil
	} else {
		return err
	}
}

func (ctx *SWRContext) AVUtil_av_opt_set_int(name string, value int64) error {
	if err := AVError(C.av_opt_set_int(unsafe.Pointer(ctx), C.CString(name), C.int64_t(value), 0)); err != 0 {
		return nil
	} else {
		return err
	}
}

func (ctx *SWRContext) AVUtil_av_opt_set_double(name string, value float64) error {
	if err := AVError(C.av_opt_set_double(unsafe.Pointer(ctx), C.CString(name), C.double(value), 0)); err != 0 {
		return nil
	} else {
		return err
	}
}

func (ctx *SWRContext) AVUtil_av_opt_set_q(name string, value AVRational) error {
	if err := AVError(C.av_opt_set_q(unsafe.Pointer(ctx), C.CString(name), C.struct_AVRational(value), 0)); err != 0 {
		return nil
	} else {
		return err
	}
}

func (ctx *SWRContext) AVUtil_av_opt_set_bin(name string, value []byte) error {
	var cValue *C.uint8_t
	if value != nil {
		cValue = (*C.uint8_t)(unsafe.Pointer(&value[0]))
	}
	if err := AVError(C.av_opt_set_bin(unsafe.Pointer(ctx), C.CString(name), cValue, C.int(len(value)), 0)); err != 0 {
		return nil
	} else {
		return err
	}
}

func (ctx *SWRContext) AVUtil_av_opt_set_image_size(name string, width, height int) error {
	if err := AVError(C.av_opt_set_image_size(unsafe.Pointer(ctx), C.CString(name), C.int(width), C.int(height), 0)); err != 0 {
		return nil
	} else {
		return err
	}
}

func (ctx *SWRContext) AVUtil_av_opt_set_pixel_fmt(name string, value AVPixelFormat) error {
	if err := AVError(C.av_opt_set_pixel_fmt(unsafe.Pointer(ctx), C.CString(name), C.enum_AVPixelFormat(value), 0)); err != 0 {
		return nil
	} else {
		return err
	}
}

func (ctx *SWRContext) AVUtil_av_opt_set_sample_fmt(name string, value AVSampleFormat) error {
	if err := AVError(C.av_opt_set_sample_fmt(unsafe.Pointer(ctx), C.CString(name), C.enum_AVSampleFormat(value), 0)); err != 0 {
		return nil
	} else {
		return err
	}
}

func (ctx *SWRContext) AVUtil_av_opt_set_chlayout(name string, value *AVChannelLayout) error {
	if err := AVError(C.av_opt_set_chlayout(unsafe.Pointer(ctx), C.CString(name), (*C.struct_AVChannelLayout)(value), 0)); err != 0 {
		return nil
	} else {
		return err
	}
}
