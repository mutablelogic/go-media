package libheif

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: --static libheif
#include <stdlib.h>
#include <libheif/heif_context.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Context C.heif_context
	ItemID  C.heif_item_id
)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - CONTEXT LIFECYCLE

func Libheif_context_alloc() *Context {
	ctx := C.heif_context_alloc()
	if ctx == nil {
		return nil
	}
	return (*Context)(ctx)
}

func Libheif_context_free(ctx *Context) {
	if ctx == nil {
		return
	}
	C.heif_context_free((*C.heif_context)(ctx))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - CONTEXT READ

func Libheif_context_read_from_file(ctx *Context, filename string) error {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	cerr := C.heif_context_read_from_file(
		(*C.heif_context)(ctx),
		(*C.char)(cfilename),
		nil,
	)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return nil
	}
	return err
}

func Libheif_context_read_from_memory(ctx *Context, data []byte) error {
	if len(data) == 0 {
		cerr := C.heif_context_read_from_memory(
			(*C.heif_context)(ctx),
			nil,
			0,
			nil,
		)
		err := fromCError(cerr)
		if err.Code == HEIF_ERROR_OK {
			return nil
		}
		return err
	}

	cerr := C.heif_context_read_from_memory(
		(*C.heif_context)(ctx),
		unsafe.Pointer(&data[0]),
		C.size_t(len(data)),
		nil,
	)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return nil
	}
	return err
}

func Libheif_context_read_from_memory_without_copy(ctx *Context, data []byte) error {
	if len(data) == 0 {
		cerr := C.heif_context_read_from_memory_without_copy(
			(*C.heif_context)(ctx),
			nil,
			0,
			nil,
		)
		err := fromCError(cerr)
		if err.Code == HEIF_ERROR_OK {
			return nil
		}
		return err
	}

	cerr := C.heif_context_read_from_memory_without_copy(
		(*C.heif_context)(ctx),
		unsafe.Pointer(&data[0]),
		C.size_t(len(data)),
		nil,
	)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return nil
	}
	return err
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - CONTEXT INFO

func Libheif_context_get_number_of_top_level_images(ctx *Context) int {
	return int(C.heif_context_get_number_of_top_level_images((*C.heif_context)(ctx)))
}

func Libheif_context_is_top_level_image_ID(ctx *Context, id ItemID) bool {
	return C.heif_context_is_top_level_image_ID((*C.heif_context)(ctx), C.heif_item_id(id)) != 0
}

func Libheif_context_get_primary_image_ID(ctx *Context) (ItemID, error) {
	var id C.heif_item_id
	cerr := C.heif_context_get_primary_image_ID((*C.heif_context)(ctx), &id)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return ItemID(id), nil
	}
	return ItemID(id), err
}

func Libheif_context_get_primary_image_handle(ctx *Context) (*ImageHandle, error) {
	var handle *C.heif_image_handle
	cerr := C.heif_context_get_primary_image_handle((*C.heif_context)(ctx), &handle)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return (*ImageHandle)(handle), nil
	}
	return nil, err
}

func Libheif_context_get_image_handle(ctx *Context, id ItemID) (*ImageHandle, error) {
	var handle *C.heif_image_handle
	cerr := C.heif_context_get_image_handle((*C.heif_context)(ctx), C.heif_item_id(id), &handle)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return (*ImageHandle)(handle), nil
	}
	return nil, err
}

func Libheif_context_get_list_of_top_level_image_IDs(ctx *Context, count int) []ItemID {
	if count <= 0 {
		return nil
	}

	ids := make([]C.heif_item_id, count)
	n := C.heif_context_get_list_of_top_level_image_IDs((*C.heif_context)(ctx), &ids[0], C.int(count))
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
