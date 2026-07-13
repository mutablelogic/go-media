package libheif

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: --static libheif
#include <stdlib.h>
#include <libheif/heif_encoding.h>
#include <libheif/heif_context.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Encoder              C.heif_encoder
	EncoderDescriptor    C.heif_encoder_descriptor
	EncoderParameter     C.heif_encoder_parameter
	EncodingOptions      C.heif_encoding_options
	CompressionFormat    C.heif_compression_format
	Brand2               C.heif_brand2
	EncoderParameterType C.heif_encoder_parameter_type
)

////////////////////////////////////////////////////////////////////////////////
// CONSTS

const (
	HEIF_COMPRESSION_UNDEFINED    CompressionFormat = C.heif_compression_undefined
	HEIF_COMPRESSION_HEVC         CompressionFormat = C.heif_compression_HEVC
	HEIF_COMPRESSION_AVC          CompressionFormat = C.heif_compression_AVC
	HEIF_COMPRESSION_JPEG         CompressionFormat = C.heif_compression_JPEG
	HEIF_COMPRESSION_AV1          CompressionFormat = C.heif_compression_AV1
	HEIF_COMPRESSION_VVC          CompressionFormat = C.heif_compression_VVC
	HEIF_COMPRESSION_EVC          CompressionFormat = C.heif_compression_EVC
	HEIF_COMPRESSION_JPEG2000     CompressionFormat = C.heif_compression_JPEG2000
	HEIF_COMPRESSION_UNCOMPRESSED CompressionFormat = C.heif_compression_uncompressed
	HEIF_COMPRESSION_MASK         CompressionFormat = C.heif_compression_mask
	HEIF_COMPRESSION_HTJ2K        CompressionFormat = C.heif_compression_HTJ2K
)

const (
	HEIF_BRAND2_HEIC Brand2 = C.heif_brand2_heic
	HEIF_BRAND2_HEIX Brand2 = C.heif_brand2_heix
	HEIF_BRAND2_HEVC Brand2 = C.heif_brand2_hevc
	HEIF_BRAND2_HEVX Brand2 = C.heif_brand2_hevx
	HEIF_BRAND2_AVIF Brand2 = C.heif_brand2_avif
	HEIF_BRAND2_AVIS Brand2 = C.heif_brand2_avis
	HEIF_BRAND2_MIF1 Brand2 = C.heif_brand2_mif1
	HEIF_BRAND2_MIF2 Brand2 = C.heif_brand2_mif2
	HEIF_BRAND2_MIF3 Brand2 = C.heif_brand2_mif3
	HEIF_BRAND2_MSF1 Brand2 = C.heif_brand2_msf1
	HEIF_BRAND2_VVIC Brand2 = C.heif_brand2_vvic
	HEIF_BRAND2_VVIS Brand2 = C.heif_brand2_vvis
	HEIF_BRAND2_JPEG Brand2 = C.heif_brand2_jpeg
	HEIF_BRAND2_JPGS Brand2 = C.heif_brand2_jpgs
	HEIF_BRAND2_J2KI Brand2 = C.heif_brand2_j2ki
	HEIF_BRAND2_J2IS Brand2 = C.heif_brand2_j2is
	HEIF_BRAND2_MIAF Brand2 = C.heif_brand2_miaf
	HEIF_BRAND2_1PIC Brand2 = C.heif_brand2_1pic
	HEIF_BRAND2_AVCI Brand2 = C.heif_brand2_avci
	HEIF_BRAND2_AVCS Brand2 = C.heif_brand2_avcs
	HEIF_BRAND2_UNIF Brand2 = C.heif_brand2_unif
	HEIF_BRAND2_ISO8 Brand2 = C.heif_brand2_iso8
	HEIF_BRAND2_ISOM Brand2 = C.heif_brand2_isom
	HEIF_BRAND2_MP41 Brand2 = C.heif_brand2_mp41
	HEIF_BRAND2_MP42 Brand2 = C.heif_brand2_mp42
)

const (
	HEIF_ENCODER_PARAMETER_INTEGER EncoderParameterType = C.heif_encoder_parameter_type_integer
	HEIF_ENCODER_PARAMETER_BOOLEAN EncoderParameterType = C.heif_encoder_parameter_type_boolean
	HEIF_ENCODER_PARAMETER_STRING  EncoderParameterType = C.heif_encoder_parameter_type_string
)

func cStringOrNil(s string) *C.char {
	if s == "" {
		return nil
	}
	return C.CString(s)
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - ENCODER

func Libheif_have_encoder_for_format(format CompressionFormat) bool {
	return C.heif_have_encoder_for_format(C.heif_compression_format(format)) != 0
}

func Libheif_get_encoder_descriptors_count(format CompressionFormat, nameFilter string) int {
	cname := cStringOrNil(nameFilter)
	if cname != nil {
		defer C.free(unsafe.Pointer(cname))
	}
	return int(C.heif_get_encoder_descriptors(C.heif_compression_format(format), cname, nil, 0))
}

func Libheif_get_encoder_descriptors(format CompressionFormat, nameFilter string, count int) []*EncoderDescriptor {
	if count <= 0 {
		return nil
	}

	cname := cStringOrNil(nameFilter)
	if cname != nil {
		defer C.free(unsafe.Pointer(cname))
	}

	descriptors := make([]*C.heif_encoder_descriptor, count)
	n := C.heif_get_encoder_descriptors(
		C.heif_compression_format(format),
		cname,
		(**C.heif_encoder_descriptor)(unsafe.Pointer(&descriptors[0])),
		C.int(count),
	)
	if n <= 0 {
		return nil
	}
	if int(n) > len(descriptors) {
		n = C.int(len(descriptors))
	}

	result := make([]*EncoderDescriptor, int(n))
	for i := 0; i < int(n); i++ {
		result[i] = (*EncoderDescriptor)(descriptors[i])
	}
	return result
}

func Libheif_context_get_encoder(ctx *Context, descriptor *EncoderDescriptor) (*Encoder, error) {
	var enc *C.heif_encoder
	cerr := C.heif_context_get_encoder(
		(*C.heif_context)(ctx),
		(*C.heif_encoder_descriptor)(descriptor),
		&enc,
	)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return (*Encoder)(enc), nil
	}
	return nil, err
}

func Libheif_encoder_descriptor_get_name(descriptor *EncoderDescriptor) string {
	return C.GoString(C.heif_encoder_descriptor_get_name((*C.heif_encoder_descriptor)(descriptor)))
}

func Libheif_encoder_descriptor_get_id_name(descriptor *EncoderDescriptor) string {
	return C.GoString(C.heif_encoder_descriptor_get_id_name((*C.heif_encoder_descriptor)(descriptor)))
}

func Libheif_encoder_descriptor_get_compression_format(descriptor *EncoderDescriptor) CompressionFormat {
	return CompressionFormat(C.heif_encoder_descriptor_get_compression_format((*C.heif_encoder_descriptor)(descriptor)))
}

func Libheif_encoder_descriptor_supports_lossy_compression(descriptor *EncoderDescriptor) bool {
	return C.heif_encoder_descriptor_supports_lossy_compression((*C.heif_encoder_descriptor)(descriptor)) != 0
}

func Libheif_encoder_descriptor_supports_lossless_compression(descriptor *EncoderDescriptor) bool {
	return C.heif_encoder_descriptor_supports_lossless_compression((*C.heif_encoder_descriptor)(descriptor)) != 0
}

func Libheif_context_get_encoder_for_format(ctx *Context, format CompressionFormat) (*Encoder, error) {
	var enc *C.heif_encoder
	cerr := C.heif_context_get_encoder_for_format(
		(*C.heif_context)(ctx),
		C.heif_compression_format(format),
		&enc,
	)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return (*Encoder)(enc), nil
	}
	return nil, err
}

func Libheif_context_set_primary_image(ctx *Context, handle *ImageHandle) error {
	cerr := C.heif_context_set_primary_image((*C.heif_context)(ctx), (*C.heif_image_handle)(handle))
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return nil
	}
	return err
}

func Libheif_context_set_major_brand(ctx *Context, brand Brand2) {
	C.heif_context_set_major_brand((*C.heif_context)(ctx), C.heif_brand2(brand))
}

func Libheif_context_add_compatible_brand(ctx *Context, brand Brand2) {
	C.heif_context_add_compatible_brand((*C.heif_context)(ctx), C.heif_brand2(brand))
}

func Libheif_context_set_write_mini_format(ctx *Context, enable bool) {
	value := 0
	if enable {
		value = 1
	}
	C.heif_context_set_write_mini_format((*C.heif_context)(ctx), C.int(value))
}

func Libheif_encoder_release(encoder *Encoder) {
	if encoder == nil {
		return
	}
	C.heif_encoder_release((*C.heif_encoder)(encoder))
}

func Libheif_encoder_get_name(encoder *Encoder) string {
	return C.GoString(C.heif_encoder_get_name((*C.heif_encoder)(encoder)))
}

func Libheif_encoder_list_parameters(encoder *Encoder) []*EncoderParameter {
	params := C.heif_encoder_list_parameters((*C.heif_encoder)(encoder))
	if params == nil {
		return nil
	}

	result := make([]*EncoderParameter, 0, 16)
	entries := (*[1 << 20]*C.heif_encoder_parameter)(unsafe.Pointer(params))
	for i := 0; ; i++ {
		ptr := entries[i]
		if ptr == nil {
			break
		}
		result = append(result, (*EncoderParameter)(ptr))
	}
	return result
}

func Libheif_encoder_parameter_get_name(parameter *EncoderParameter) string {
	return C.GoString(C.heif_encoder_parameter_get_name((*C.heif_encoder_parameter)(parameter)))
}

func Libheif_encoder_parameter_get_type(parameter *EncoderParameter) EncoderParameterType {
	return EncoderParameterType(C.heif_encoder_parameter_get_type((*C.heif_encoder_parameter)(parameter)))
}

func Libheif_encoder_parameter_get_valid_integer_range(parameter *EncoderParameter) (haveMinimum, haveMaximum bool, minimum, maximum int, err error) {
	var cHaveMinimumMaximum C.int
	var cMinimum, cMaximum C.int
	cerr := C.heif_encoder_parameter_get_valid_integer_range(
		(*C.heif_encoder_parameter)(parameter),
		&cHaveMinimumMaximum,
		&cMinimum,
		&cMaximum,
	)
	heifErr := fromCError(cerr)
	if heifErr.Code != HEIF_ERROR_OK {
		return false, false, 0, 0, heifErr
	}
	haveMinimum = cHaveMinimumMaximum != 0
	haveMaximum = cHaveMinimumMaximum != 0
	return haveMinimum, haveMaximum, int(cMinimum), int(cMaximum), nil
}

func Libheif_encoder_parameter_get_valid_integer_values(parameter *EncoderParameter) (haveMinimum, haveMaximum bool, minimum, maximum int, values []int, err error) {
	var cHaveMinimum, cHaveMaximum C.int
	var cMinimum, cMaximum C.int
	var cNumValidValues C.int
	var cValues *C.int
	cerr := C.heif_encoder_parameter_get_valid_integer_values(
		(*C.heif_encoder_parameter)(parameter),
		&cHaveMinimum,
		&cHaveMaximum,
		&cMinimum,
		&cMaximum,
		&cNumValidValues,
		&cValues,
	)
	heifErr := fromCError(cerr)
	if heifErr.Code != HEIF_ERROR_OK {
		return false, false, 0, 0, nil, heifErr
	}
	haveMinimum = cHaveMinimum != 0
	haveMaximum = cHaveMaximum != 0
	minimum = int(cMinimum)
	maximum = int(cMaximum)
	if cNumValidValues > 0 && cValues != nil {
		values = make([]int, int(cNumValidValues))
		entries := (*[1 << 20]C.int)(unsafe.Pointer(cValues))
		for i := 0; i < int(cNumValidValues); i++ {
			values[i] = int(entries[i])
		}
	}
	return haveMinimum, haveMaximum, minimum, maximum, values, nil
}

func Libheif_encoder_parameter_get_valid_string_values(parameter *EncoderParameter) ([]string, error) {
	var cValues **C.char
	cerr := C.heif_encoder_parameter_get_valid_string_values((*C.heif_encoder_parameter)(parameter), &cValues)
	err := fromCError(cerr)
	if err.Code != HEIF_ERROR_OK {
		return nil, err
	}
	if cValues == nil {
		return nil, nil
	}

	result := make([]string, 0, 8)
	entries := (*[1 << 20]*C.char)(unsafe.Pointer(cValues))
	for i := 0; ; i++ {
		ptr := entries[i]
		if ptr == nil {
			break
		}
		result = append(result, C.GoString(ptr))
	}
	return result, nil
}

func Libheif_encoder_set_parameter_integer(encoder *Encoder, parameterName string, value int) error {
	cname := C.CString(parameterName)
	defer C.free(unsafe.Pointer(cname))

	cerr := C.heif_encoder_set_parameter_integer((*C.heif_encoder)(encoder), cname, C.int(value))
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return nil
	}
	return err
}

func Libheif_encoder_get_parameter_integer(encoder *Encoder, parameterName string) (int, error) {
	cname := C.CString(parameterName)
	defer C.free(unsafe.Pointer(cname))

	var value C.int
	cerr := C.heif_encoder_get_parameter_integer((*C.heif_encoder)(encoder), cname, &value)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return int(value), nil
	}
	return 0, err
}

func Libheif_encoder_set_parameter_boolean(encoder *Encoder, parameterName string, value bool) error {
	cname := C.CString(parameterName)
	defer C.free(unsafe.Pointer(cname))

	intValue := 0
	if value {
		intValue = 1
	}
	cerr := C.heif_encoder_set_parameter_boolean((*C.heif_encoder)(encoder), cname, C.int(intValue))
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return nil
	}
	return err
}

func Libheif_encoder_get_parameter_boolean(encoder *Encoder, parameterName string) (bool, error) {
	cname := C.CString(parameterName)
	defer C.free(unsafe.Pointer(cname))

	var value C.int
	cerr := C.heif_encoder_get_parameter_boolean((*C.heif_encoder)(encoder), cname, &value)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return value != 0, nil
	}
	return false, err
}

func Libheif_encoder_set_parameter_string(encoder *Encoder, parameterName, value string) error {
	cname := C.CString(parameterName)
	defer C.free(unsafe.Pointer(cname))
	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	cerr := C.heif_encoder_set_parameter_string((*C.heif_encoder)(encoder), cname, cvalue)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return nil
	}
	return err
}

func Libheif_encoder_get_parameter_string(encoder *Encoder, parameterName string, valueSize int) (string, error) {
	if valueSize <= 0 {
		valueSize = 1024
	}
	cname := C.CString(parameterName)
	defer C.free(unsafe.Pointer(cname))

	buf := C.malloc(C.size_t(valueSize))
	if buf == nil {
		return "", HeifError{Code: HEIF_ERROR_MEMORY_ALLOCATION_ERROR, Message: "unable to allocate parameter buffer"}
	}
	defer C.free(buf)

	cerr := C.heif_encoder_get_parameter_string((*C.heif_encoder)(encoder), cname, (*C.char)(buf), C.int(valueSize))
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return C.GoString((*C.char)(buf)), nil
	}
	return "", err
}

func Libheif_encoder_set_parameter(encoder *Encoder, parameterName, value string) error {
	cname := C.CString(parameterName)
	defer C.free(unsafe.Pointer(cname))
	cvalue := C.CString(value)
	defer C.free(unsafe.Pointer(cvalue))

	cerr := C.heif_encoder_set_parameter((*C.heif_encoder)(encoder), cname, cvalue)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return nil
	}
	return err
}

func Libheif_encoder_get_parameter(encoder *Encoder, parameterName string, valueSize int) (string, error) {
	if valueSize <= 0 {
		valueSize = 1024
	}
	cname := C.CString(parameterName)
	defer C.free(unsafe.Pointer(cname))

	buf := C.malloc(C.size_t(valueSize))
	if buf == nil {
		return "", HeifError{Code: HEIF_ERROR_MEMORY_ALLOCATION_ERROR, Message: "unable to allocate parameter buffer"}
	}
	defer C.free(buf)

	cerr := C.heif_encoder_get_parameter((*C.heif_encoder)(encoder), cname, (*C.char)(buf), C.int(valueSize))
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return C.GoString((*C.char)(buf)), nil
	}
	return "", err
}

func Libheif_encoder_has_default(encoder *Encoder, parameterName string) bool {
	cname := C.CString(parameterName)
	defer C.free(unsafe.Pointer(cname))
	return C.heif_encoder_has_default((*C.heif_encoder)(encoder), cname) != 0
}

func Libheif_encoder_set_lossy_quality(encoder *Encoder, quality int) error {
	cerr := C.heif_encoder_set_lossy_quality((*C.heif_encoder)(encoder), C.int(quality))
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return nil
	}
	return err
}

func Libheif_encoder_set_lossless(encoder *Encoder, enable bool) error {
	value := 0
	if enable {
		value = 1
	}
	cerr := C.heif_encoder_set_lossless((*C.heif_encoder)(encoder), C.int(value))
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return nil
	}
	return err
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - ENCODING OPTIONS

func Libheif_encoding_options_alloc() *EncodingOptions {
	opts := C.heif_encoding_options_alloc()
	if opts == nil {
		return nil
	}
	return (*EncodingOptions)(opts)
}

func Libheif_encoding_options_copy(dst, src *EncodingOptions) {
	if dst == nil || src == nil {
		return
	}
	C.heif_encoding_options_copy((*C.heif_encoding_options)(dst), (*C.heif_encoding_options)(src))
}

func Libheif_encoding_options_free(opts *EncodingOptions) {
	if opts == nil {
		return
	}
	C.heif_encoding_options_free((*C.heif_encoding_options)(opts))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - CONTEXT ENCODING/WRITE

func Libheif_context_encode_image(ctx *Context, img *Image, encoder *Encoder, options *EncodingOptions) (*ImageHandle, error) {
	var handle *C.heif_image_handle
	var coptions *C.heif_encoding_options
	if options != nil {
		coptions = (*C.heif_encoding_options)(options)
	}
	cerr := C.heif_context_encode_image(
		(*C.heif_context)(ctx),
		(*C.heif_image)(img),
		(*C.heif_encoder)(encoder),
		coptions,
		&handle,
	)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return (*ImageHandle)(handle), nil
	}
	return nil, err
}

func Libheif_context_write_to_file(ctx *Context, filename string) error {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	cerr := C.heif_context_write_to_file((*C.heif_context)(ctx), cfilename)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return nil
	}
	return err
}
