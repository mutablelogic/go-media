package libheif

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: --static libheif
#include <stdlib.h>
#include <libheif/heif_image_handle.h>
#include <libheif/heif_aux_images.h>
#include <libheif/heif_metadata.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	ImageHandle C.heif_image_handle
)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - IMAGE HANDLE

func Libheif_image_handle_release(handle *ImageHandle) {
	if handle == nil {
		return
	}
	C.heif_image_handle_release((*C.heif_image_handle)(handle))
}

func Libheif_image_handle_is_primary_image(handle *ImageHandle) bool {
	return C.heif_image_handle_is_primary_image((*C.heif_image_handle)(handle)) != 0
}

func Libheif_image_handle_get_item_id(handle *ImageHandle) ItemID {
	return ItemID(C.heif_image_handle_get_item_id((*C.heif_image_handle)(handle)))
}

func Libheif_image_handle_get_width(handle *ImageHandle) int {
	return int(C.heif_image_handle_get_width((*C.heif_image_handle)(handle)))
}

func Libheif_image_handle_get_height(handle *ImageHandle) int {
	return int(C.heif_image_handle_get_height((*C.heif_image_handle)(handle)))
}

func Libheif_image_handle_has_alpha_channel(handle *ImageHandle) bool {
	return C.heif_image_handle_has_alpha_channel((*C.heif_image_handle)(handle)) != 0
}

func Libheif_image_handle_is_premultiplied_alpha(handle *ImageHandle) bool {
	return C.heif_image_handle_is_premultiplied_alpha((*C.heif_image_handle)(handle)) != 0
}

func Libheif_image_handle_get_luma_bits_per_pixel(handle *ImageHandle) int {
	return int(C.heif_image_handle_get_luma_bits_per_pixel((*C.heif_image_handle)(handle)))
}

func Libheif_image_handle_get_chroma_bits_per_pixel(handle *ImageHandle) int {
	return int(C.heif_image_handle_get_chroma_bits_per_pixel((*C.heif_image_handle)(handle)))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - THUMBNAILS

func Libheif_image_handle_get_number_of_thumbnails(handle *ImageHandle) int {
	return int(C.heif_image_handle_get_number_of_thumbnails((*C.heif_image_handle)(handle)))
}

func Libheif_image_handle_get_list_of_thumbnail_IDs(handle *ImageHandle, count int) []ItemID {
	if count <= 0 {
		return nil
	}

	ids := make([]C.heif_item_id, count)
	n := C.heif_image_handle_get_list_of_thumbnail_IDs((*C.heif_image_handle)(handle), &ids[0], C.int(count))
	if n <= 0 {
		return nil
	}
	if int(n) > len(ids) {
		n = C.int(len(ids))
	}

	result := make([]ItemID, int(n))
	for i := 0; i < int(n); i++ {
		result[i] = ItemID(ids[i])
	}
	return result
}

func Libheif_image_handle_get_thumbnail(handle *ImageHandle, thumbnailID ItemID) (*ImageHandle, error) {
	var thumb *C.heif_image_handle
	cerr := C.heif_image_handle_get_thumbnail(
		(*C.heif_image_handle)(handle),
		C.heif_item_id(thumbnailID),
		&thumb,
	)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return (*ImageHandle)(thumb), nil
	}
	return nil, err
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - IMAGE METADATA

func Libheif_image_handle_get_number_of_metadata_blocks(handle *ImageHandle, typeFilter string) int {
	var ctypeFilter *C.char
	if typeFilter != "" {
		ctypeFilter = C.CString(typeFilter)
		defer C.free(unsafe.Pointer(ctypeFilter))
	}
	return int(C.heif_image_handle_get_number_of_metadata_blocks((*C.heif_image_handle)(handle), ctypeFilter))
}

func Libheif_image_handle_get_list_of_metadata_block_IDs(handle *ImageHandle, typeFilter string, count int) []ItemID {
	if count <= 0 {
		return nil
	}

	var ctypeFilter *C.char
	if typeFilter != "" {
		ctypeFilter = C.CString(typeFilter)
		defer C.free(unsafe.Pointer(ctypeFilter))
	}

	ids := make([]C.heif_item_id, count)
	n := C.heif_image_handle_get_list_of_metadata_block_IDs((*C.heif_image_handle)(handle), ctypeFilter, &ids[0], C.int(count))
	if n <= 0 {
		return nil
	}
	if int(n) > len(ids) {
		n = C.int(len(ids))
	}

	result := make([]ItemID, int(n))
	for i := 0; i < int(n); i++ {
		result[i] = ItemID(ids[i])
	}
	return result
}

func Libheif_image_handle_get_metadata_type(handle *ImageHandle, metadataID ItemID) string {
	return C.GoString(C.heif_image_handle_get_metadata_type((*C.heif_image_handle)(handle), C.heif_item_id(metadataID)))
}

func Libheif_image_handle_get_metadata_content_type(handle *ImageHandle, metadataID ItemID) string {
	return C.GoString(C.heif_image_handle_get_metadata_content_type((*C.heif_image_handle)(handle), C.heif_item_id(metadataID)))
}

func Libheif_image_handle_get_metadata_size(handle *ImageHandle, metadataID ItemID) int {
	return int(C.heif_image_handle_get_metadata_size((*C.heif_image_handle)(handle), C.heif_item_id(metadataID)))
}

func Libheif_image_handle_get_metadata(handle *ImageHandle, metadataID ItemID) ([]byte, error) {
	size := Libheif_image_handle_get_metadata_size(handle, metadataID)
	if size <= 0 {
		return nil, nil
	}
	data := make([]byte, size)
	cerr := C.heif_image_handle_get_metadata(
		(*C.heif_image_handle)(handle),
		C.heif_item_id(metadataID),
		unsafe.Pointer(&data[0]),
	)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return data, nil
	}
	return nil, err
}
