package ffmpeg

import (
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - VERSION

// Return the LIBAVFORMAT_VERSION_INT constant.
func AVFormat_version() uint {
	return uint(C.avformat_version())
}

// Return the libavformat build-time configuration.
func AVFormat_configuration() string {
	return C.GoString(C.avformat_configuration())
}

// Return the libavformat license.
func AVFormat_license() string {
	return C.GoString(C.avformat_license())
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MUXER/DEMUXER ITERATORS

func AVFormat_av_muxer_iterate(opaque *uintptr) *AVOutputFormat {
	return (*AVOutputFormat)(C.av_muxer_iterate((*unsafe.Pointer)(unsafe.Pointer(opaque))))
}

func AVFormat_av_demuxer_iterate(opaque *uintptr) *AVInputFormat {
	return (*AVInputFormat)(C.av_demuxer_iterate((*unsafe.Pointer)(unsafe.Pointer(opaque))))
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - CONTEXT

func AVFormat_alloc_context() *AVFormatContext {
	return (*AVFormatContext)(C.avformat_alloc_context())
}

func AVFormat_free_context(ctx *AVFormatContext) {
	C.avformat_free_context((*C.struct_AVFormatContext)(ctx))
}

func AVFormat_avformat_get_class() *AVClass {
	return (*AVClass)(C.avformat_get_class())
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - AVSTREAM

func AVFormat_avstream_getclass() *AVClass {
	return (*AVClass)(C.av_stream_get_class())
}

func AVFormat_avformat_new_stream(oc *AVFormatContext, c *AVCodec) *AVStream {
	return (*AVStream)(C.avformat_new_stream((*C.struct_AVFormatContext)(oc), (*C.struct_AVCodec)(c)))
}
