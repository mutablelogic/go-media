package ffmpeg

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Allocate an AVCodecContext and set its fields to default values.
func AVCodec_alloc_context(codec *AVCodec) *AVCodecContext {
	return (*AVCodecContext)(C.avcodec_alloc_context3((*C.struct_AVCodec)(codec)))
}

// Free the codec context and everything associated with it.
func AVCodec_free_context(ctx *AVCodecContext) {
	C.avcodec_free_context((**C.struct_AVCodecContext)(unsafe.Pointer(&ctx)))
}

// From fill the parameters based on the values from the supplied codec parameters
func AVCodec_parameters_copy(ctx *AVCodecParameters, codecpar *AVCodecParameters) error {
	if err := AVError(C.avcodec_parameters_copy((*C.AVCodecParameters)(ctx), (*C.AVCodecParameters)(codecpar))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Fill the parameters struct based on the values from the supplied codec context (encoding)
func AVCodec_parameters_from_context(codecpar *AVCodecParameters, ctx *AVCodecContext) error {
	if err := AVError(C.avcodec_parameters_from_context((*C.AVCodecParameters)(codecpar), (*C.struct_AVCodecContext)(ctx))); err < 0 {
		return err
	}
	return nil
}

// Fill the codec context based on the values from the supplied codec parameters (decoding)
func AVCodec_parameters_to_context(ctx *AVCodecContext, codecpar *AVCodecParameters) error {
	if err := AVError(C.avcodec_parameters_to_context((*C.struct_AVCodecContext)(ctx), (*C.AVCodecParameters)(codecpar))); err < 0 {
		return err
	}
	return nil
}

// Initialize the AVCodecContext to use the given AVCodec.
func AVCodec_open(ctx *AVCodecContext, codec *AVCodec, options *AVDictionary) error {
	var opts **C.struct_AVDictionary
	if options != nil {
		opts = &options.ctx
	}
	if err := AVError(C.avcodec_open2((*C.struct_AVCodecContext)(ctx), (*C.struct_AVCodec)(codec), opts)); err != 0 {
		return err
	}
	return nil
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
func AVCodec_is_encoder(codec *AVCodec) bool {
	return C.av_codec_is_encoder((*C.struct_AVCodec)(codec)) != 0
}

// Return true if codec is a decoder, false otherwise.
func AVCodec_is_decoder(codec *AVCodec) bool {
	return C.av_codec_is_decoder((*C.struct_AVCodec)(codec)) != 0
}
