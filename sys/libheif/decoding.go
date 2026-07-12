package libheif

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: --static libheif
#include <libheif/heif_decoding.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	DecodingOptions C.heif_decoding_options
)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - CONTEXT DECODING CONTROL

func Libheif_have_decoder_for_format(format CompressionFormat) bool {
	return C.heif_have_decoder_for_format(C.heif_compression_format(format)) != 0
}

func Libheif_context_set_max_decoding_threads(ctx *Context, maxThreads int) {
	C.heif_context_set_max_decoding_threads((*C.heif_context)(ctx), C.int(maxThreads))
}

func Libheif_context_get_max_decoding_threads(ctx *Context) int {
	return int(C.heif_context_get_max_decoding_threads((*C.heif_context)(ctx)))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - DECODING OPTIONS

func Libheif_decoding_options_alloc() *DecodingOptions {
	opts := C.heif_decoding_options_alloc()
	if opts == nil {
		return nil
	}
	return (*DecodingOptions)(opts)
}

func Libheif_decoding_options_copy(dst, src *DecodingOptions) {
	if dst == nil || src == nil {
		return
	}
	C.heif_decoding_options_copy((*C.heif_decoding_options)(dst), (*C.heif_decoding_options)(src))
}

func Libheif_decoding_options_free(opts *DecodingOptions) {
	if opts == nil {
		return
	}
	C.heif_decoding_options_free((*C.heif_decoding_options)(opts))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - DECODING

func Libheif_decode_image(handle *ImageHandle, colorspace HeifColorspace, chroma HeifChroma) (*Image, error) {
	return Libheif_decode_image_with_options(handle, colorspace, chroma, nil)
}

func Libheif_decode_image_with_options(handle *ImageHandle, colorspace HeifColorspace, chroma HeifChroma, options *DecodingOptions) (*Image, error) {
	var img *C.heif_image
	var coptions *C.heif_decoding_options
	if options != nil {
		coptions = (*C.heif_decoding_options)(options)
	}
	cerr := C.heif_decode_image(
		(*C.heif_image_handle)(handle),
		&img,
		C.heif_colorspace(colorspace),
		C.heif_chroma(chroma),
		coptions,
	)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return (*Image)(img), nil
	}
	return nil, err
}

func Libheif_decode_primary_image(ctx *Context, colorspace HeifColorspace, chroma HeifChroma) (*Image, error) {
	handle, err := Libheif_context_get_primary_image_handle(ctx)
	if err != nil {
		return nil, err
	}
	if handle == nil {
		return nil, HeifError{Code: HEIF_ERROR_INVALID_INPUT, Message: "primary image handle is nil"}
	}
	defer Libheif_image_handle_release(handle)

	return Libheif_decode_image(handle, colorspace, chroma)
}

func Libheif_decode_primary_image_rgb(ctx *Context) (*Image, error) {
	return Libheif_decode_primary_image(ctx, HEIF_COLORSPACE_RGB, HEIF_CHROMA_INTERLEAVED_RGB)
}

func Libheif_decode_image_by_item_id(ctx *Context, id ItemID, colorspace HeifColorspace, chroma HeifChroma) (*Image, error) {
	handle, err := Libheif_context_get_image_handle(ctx, id)
	if err != nil {
		return nil, err
	}
	if handle == nil {
		return nil, HeifError{Code: HEIF_ERROR_INVALID_INPUT, Message: "image handle is nil"}
	}
	defer Libheif_image_handle_release(handle)

	return Libheif_decode_image(handle, colorspace, chroma)
}
