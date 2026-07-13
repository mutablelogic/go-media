package libheif

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: --static libheif
#include <libheif/heif_library.h>
#include <libheif/heif_version.h>

static const char* go_libheif_plugin_directory(void) {
	return LIBHEIF_PLUGIN_DIRECTORY;
}
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - LIBRARY

func Libheif_get_version() string {
	return C.GoString(C.heif_get_version())
}

func Libheif_get_version_number() uint32 {
	return uint32(C.heif_get_version_number())
}

func Libheif_get_version_number_major() int {
	return int(C.heif_get_version_number_major())
}

func Libheif_get_version_number_minor() int {
	return int(C.heif_get_version_number_minor())
}

func Libheif_get_version_number_maintenance() int {
	return int(C.heif_get_version_number_maintenance())
}

func Libheif_get_plugin_directory() string {
	return C.GoString(C.go_libheif_plugin_directory())
}

func Libheif_init() error {
	cerr := C.heif_init(nil)
	err := fromCError(cerr)
	if err.Code == HEIF_ERROR_OK {
		return nil
	}
	return err
}

func Libheif_deinit() {
	C.heif_deinit()
}
