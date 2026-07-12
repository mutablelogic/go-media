package libheif

import (
	"errors"
	"fmt"
	"runtime/cgo"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: --static libheif
#include <libheif/heif_context.h>

extern heif_error heif_go_writer_write_callback(heif_context* ctx, void* data, size_t size, void* userdata);

static heif_error go_heif_writer_write_callback_c(heif_context* ctx, const void* data, size_t size, void* userdata) {
	return heif_go_writer_write_callback(ctx, (void*)data, size, userdata);
}

static heif_error go_heif_context_write(heif_context* ctx, void* userdata) {
	heif_writer writer;
	writer.writer_api_version = 1;
	writer.write = go_heif_writer_write_callback_c;
	return heif_context_write(ctx, &writer, userdata);
}

static heif_error go_heif_success(void) {
	return heif_error_success;
}

static heif_error go_heif_make_error(int code, int subcode) {
	heif_error err;
	err.code = (heif_error_code)code;
	err.subcode = (heif_suberror_code)subcode;
	err.message = "go writer callback error";
	return err;
}
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type WriterFunc func([]byte) error

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - CONTEXT WRITE

func Libheif_context_write(ctx *Context, writer WriterFunc) error {
	if writer == nil {
		return HeifError{Code: HEIF_ERROR_USAGE_ERROR, Subcode: HEIF_SUBERROR_UNSPECIFIED, Message: "writer is nil"}
	}

	handle := cgo.NewHandle(writer)
	defer handle.Delete()

	cerr := C.go_heif_context_write((*C.heif_context)(ctx), unsafe.Pointer(handle))
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return nil
	}
	return err
}

////////////////////////////////////////////////////////////////////////////////
// CALLBACKS

//export heif_go_writer_write_callback
func heif_go_writer_write_callback(ctx *C.heif_context, data unsafe.Pointer, size C.size_t, userdata unsafe.Pointer) (ret C.heif_error) {
	_ = ctx

	defer func() {
		if r := recover(); r != nil {
			ret = C.go_heif_make_error(C.int(HEIF_ERROR_USAGE_ERROR), C.int(HEIF_SUBERROR_UNSPECIFIED))
		}
	}()

	handle := cgo.Handle(uintptr(userdata))
	writer, ok := handle.Value().(WriterFunc)
	if !ok {
		return C.go_heif_make_error(C.int(HEIF_ERROR_USAGE_ERROR), C.int(HEIF_SUBERROR_UNSPECIFIED))
	}

	var buf []byte
	if size > 0 && data != nil {
		buf = unsafe.Slice((*byte)(data), int(size))
	}

	if err := writer(buf); err != nil {
		var heifErr HeifError
		if errors.As(err, &heifErr) {
			return C.go_heif_make_error(C.int(heifErr.Code), C.int(heifErr.Subcode))
		}
		return C.go_heif_make_error(C.int(HEIF_ERROR_USAGE_ERROR), C.int(HEIF_SUBERROR_UNSPECIFIED))
	}

	return C.go_heif_success()
}

////////////////////////////////////////////////////////////////////////////////
// HELPERS

func Libheif_context_write_to_writer(ctx *Context, writer WriterFunc) error {
	return Libheif_context_write(ctx, writer)
}

func Libheif_writer_write_all(w WriterFunc, data []byte) error {
	if w == nil {
		return fmt.Errorf("writer is nil")
	}
	return w(data)
}
