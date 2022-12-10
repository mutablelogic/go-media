package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
#include <stdlib.h>
*/
import "C"
import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the LIBAVCODEC_VERSION_INT constant.
func AVCodec_version() uint {
	return uint(C.avcodec_version())
}

// Return the libavcodec build-time configuration.
func AVCodec_configuration() string {
	return C.GoString(C.avcodec_configuration())
}

// Return the libavcodec license.
func AVCodec_license() string {
	return C.GoString(C.avcodec_license())
}

// Allocate an AVCodecContext and set its fields to default values.
func AVCodec_alloc_context3(codec *AVCodec) *AVCodecContext {
	return (*AVCodecContext)(C.avcodec_alloc_context3((*C.struct_AVCodec)(codec)))
}

// Free the codec context and everything associated with it and write NULL to the provided pointer.
func AVCodec_free_context(ctx **AVCodecContext) {
	C.avcodec_free_context((**C.struct_AVCodecContext)(unsafe.Pointer(ctx)))
}

// Get the AVClass for AVCodecContext.
func AVCodec_get_class() *AVClass {
	return (*AVClass)(C.avcodec_get_class())
}

// Get the AVClass for AVSubtitleRect.
func AVCodec_get_subtitle_rect_class() *AVClass {
	return (*AVClass)(C.avcodec_get_subtitle_rect_class())
}

// Iterate over all registered codecs.
func AVCodec_iterate(opaque *uintptr) *AVCodec {
	return (*AVCodec)(C.av_codec_iterate((*unsafe.Pointer)(unsafe.Pointer(opaque))))
}

// Find a registered decoder with a matching codec ID.
func AVCodec_find_decoder(id AVCodecID) *AVCodec {
	return (*AVCodec)(C.avcodec_find_decoder((C.enum_AVCodecID)(id)))
}

// Find a registered decoder with the specified name.
func AVCodec_find_decoder_by_name(name string) *AVCodec {
	cStr := C.CString(name)
	defer C.free(unsafe.Pointer(cStr))
	return (*AVCodec)(C.avcodec_find_decoder_by_name(cStr))
}

// Find a registered encoder with a matching codec ID.
func AVCodec_find_encoder(id AVCodecID) *AVCodec {
	return (*AVCodec)(C.avcodec_find_encoder((C.enum_AVCodecID)(id)))
}

// Find a registered encoder with the specified name.
func AVCodec_find_encoder_by_name(name string) *AVCodec {
	cStr := C.CString(name)
	defer C.free(unsafe.Pointer(cStr))
	return (*AVCodec)(C.avcodec_find_encoder_by_name(cStr))
}

// Return true if codec is an encoder, false otherwise.
func (codec *AVCodec) AVCodec_is_encoder() bool {
	return intToBool(int(C.av_codec_is_encoder((*C.struct_AVCodec)(codec))))
}

// Return true if codec is a decoder, false otherwise.
func (codec *AVCodec) AVCodec_is_decoder() bool {
	return intToBool(int(C.av_codec_is_decoder((*C.struct_AVCodec)(codec))))
}
