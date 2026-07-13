package libheif

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: --static libheif
#include <libheif/heif_image.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Image          C.heif_image
	HeifColorspace C.heif_colorspace
	HeifChroma     C.heif_chroma
	HeifChannel    C.heif_channel
)

////////////////////////////////////////////////////////////////////////////////
// CONSTS

const (
	HEIF_COLORSPACE_UNDEFINED  HeifColorspace = C.heif_colorspace_undefined
	HEIF_COLORSPACE_YCBCR      HeifColorspace = C.heif_colorspace_YCbCr
	HEIF_COLORSPACE_RGB        HeifColorspace = C.heif_colorspace_RGB
	HEIF_COLORSPACE_MONOCHROME HeifColorspace = C.heif_colorspace_monochrome
	HEIF_COLORSPACE_CUSTOM     HeifColorspace = C.heif_colorspace_custom
)

const (
	HEIF_CHROMA_UNDEFINED               HeifChroma = C.heif_chroma_undefined
	HEIF_CHROMA_PLANAR                  HeifChroma = C.heif_chroma_planar
	HEIF_CHROMA_420                     HeifChroma = C.heif_chroma_420
	HEIF_CHROMA_422                     HeifChroma = C.heif_chroma_422
	HEIF_CHROMA_444                     HeifChroma = C.heif_chroma_444
	HEIF_CHROMA_INTERLEAVED_RGB         HeifChroma = C.heif_chroma_interleaved_RGB
	HEIF_CHROMA_INTERLEAVED_RGBA        HeifChroma = C.heif_chroma_interleaved_RGBA
	HEIF_CHROMA_INTERLEAVED_RRGGBB_BE   HeifChroma = C.heif_chroma_interleaved_RRGGBB_BE
	HEIF_CHROMA_INTERLEAVED_RRGGBBAA_BE HeifChroma = C.heif_chroma_interleaved_RRGGBBAA_BE
	HEIF_CHROMA_INTERLEAVED_RRGGBB_LE   HeifChroma = C.heif_chroma_interleaved_RRGGBB_LE
	HEIF_CHROMA_INTERLEAVED_RRGGBBAA_LE HeifChroma = C.heif_chroma_interleaved_RRGGBBAA_LE
)

const (
	HEIF_CHANNEL_Y           HeifChannel = C.heif_channel_Y
	HEIF_CHANNEL_CB          HeifChannel = C.heif_channel_Cb
	HEIF_CHANNEL_CR          HeifChannel = C.heif_channel_Cr
	HEIF_CHANNEL_R           HeifChannel = C.heif_channel_R
	HEIF_CHANNEL_G           HeifChannel = C.heif_channel_G
	HEIF_CHANNEL_B           HeifChannel = C.heif_channel_B
	HEIF_CHANNEL_ALPHA       HeifChannel = C.heif_channel_Alpha
	HEIF_CHANNEL_INTERLEAVED HeifChannel = C.heif_channel_interleaved
)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - IMAGE

func Libheif_image_release(img *Image) {
	if img == nil {
		return
	}
	C.heif_image_release((*C.heif_image)(img))
}

func Libheif_image_create(width, height int, colorspace HeifColorspace, chroma HeifChroma) (*Image, error) {
	var img *C.heif_image
	cerr := C.heif_image_create(
		C.int(width),
		C.int(height),
		C.heif_colorspace(colorspace),
		C.heif_chroma(chroma),
		&img,
	)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return (*Image)(img), nil
	}
	return nil, err
}

func Libheif_image_add_plane(img *Image, channel HeifChannel, width, height, bitDepth int) error {
	cerr := C.heif_image_add_plane(
		(*C.heif_image)(img),
		C.heif_channel(channel),
		C.int(width),
		C.int(height),
		C.int(bitDepth),
	)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return nil
	}
	return err
}

func Libheif_image_get_colorspace(img *Image) HeifColorspace {
	return HeifColorspace(C.heif_image_get_colorspace((*C.heif_image)(img)))
}

func Libheif_image_get_chroma_format(img *Image) HeifChroma {
	return HeifChroma(C.heif_image_get_chroma_format((*C.heif_image)(img)))
}

func Libheif_image_get_width(img *Image, channel HeifChannel) int {
	return int(C.heif_image_get_width((*C.heif_image)(img), C.heif_channel(channel)))
}

func Libheif_image_get_height(img *Image, channel HeifChannel) int {
	return int(C.heif_image_get_height((*C.heif_image)(img), C.heif_channel(channel)))
}

func Libheif_image_get_primary_width(img *Image) int {
	return int(C.heif_image_get_primary_width((*C.heif_image)(img)))
}

func Libheif_image_get_primary_height(img *Image) int {
	return int(C.heif_image_get_primary_height((*C.heif_image)(img)))
}

func Libheif_image_has_channel(img *Image, channel HeifChannel) bool {
	return C.heif_image_has_channel((*C.heif_image)(img), C.heif_channel(channel)) != 0
}

func Libheif_image_get_bits_per_pixel_range(img *Image, channel HeifChannel) int {
	return int(C.heif_image_get_bits_per_pixel_range((*C.heif_image)(img), C.heif_channel(channel)))
}

func Libheif_image_get_pixel_aspect_ratio(img *Image) (aspectH, aspectV uint32) {
	C.heif_image_get_pixel_aspect_ratio(
		(*C.heif_image)(img),
		(*C.uint32_t)(&aspectH),
		(*C.uint32_t)(&aspectV),
	)
	return aspectH, aspectV
}

func Libheif_image_set_pixel_aspect_ratio(img *Image, aspectH, aspectV uint32) {
	C.heif_image_set_pixel_aspect_ratio((*C.heif_image)(img), C.uint32_t(aspectH), C.uint32_t(aspectV))
}

func Libheif_image_get_decoding_warnings(img *Image, firstWarningIdx, maxOutputBufferEntries int) ([]HeifError, int) {
	if maxOutputBufferEntries < 0 {
		return nil, 0
	}
	if maxOutputBufferEntries == 0 {
		count := int(C.heif_image_get_decoding_warnings((*C.heif_image)(img), C.int(firstWarningIdx), nil, 0))
		return nil, count
	}

	warnings := make([]C.heif_error, maxOutputBufferEntries)
	count := int(C.heif_image_get_decoding_warnings(
		(*C.heif_image)(img),
		C.int(firstWarningIdx),
		&warnings[0],
		C.int(maxOutputBufferEntries),
	))
	if count <= 0 {
		return nil, count
	}
	if count > len(warnings) {
		count = len(warnings)
	}

	result := make([]HeifError, count)
	for i := 0; i < count; i++ {
		result[i] = fromCError(warnings[i])
	}
	return result, count
}

func Libheif_image_get_plane_readonly(img *Image, channel HeifChannel) ([]byte, int) {
	var stride C.size_t
	ptr := C.heif_image_get_plane_readonly2((*C.heif_image)(img), C.heif_channel(channel), &stride)
	if ptr == nil {
		return nil, 0
	}
	h := Libheif_image_get_height(img, channel)
	if h <= 0 || stride == 0 {
		return nil, int(stride)
	}
	size := int(stride) * h
	if size <= 0 {
		return nil, int(stride)
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size), int(stride)
}

func Libheif_image_get_plane(img *Image, channel HeifChannel) ([]byte, int) {
	var stride C.size_t
	ptr := C.heif_image_get_plane2((*C.heif_image)(img), C.heif_channel(channel), &stride)
	if ptr == nil {
		return nil, 0
	}
	h := Libheif_image_get_height(img, channel)
	if h <= 0 || stride == 0 {
		return nil, int(stride)
	}
	size := int(stride) * h
	if size <= 0 {
		return nil, int(stride)
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(ptr)), size), int(stride)
}
