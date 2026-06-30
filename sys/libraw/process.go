package libraw

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libraw
#include <libraw/libraw.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - UNPACK

func Libraw_unpack(data *Data) int {
	return int(C.libraw_unpack((*C.libraw_data_t)(data)))
}

func Libraw_unpack_thumb(data *Data) int {
	return int(C.libraw_unpack_thumb((*C.libraw_data_t)(data)))
}

func Libraw_unpack_thumb_ex(data *Data, i int) int {
	return int(C.libraw_unpack_thumb_ex((*C.libraw_data_t)(data), C.int(i)))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - PROCESS

func Libraw_dcraw_process(data *Data) int {
	return int(C.libraw_dcraw_process((*C.libraw_data_t)(data)))
}

func Libraw_raw2image(data *Data) int {
	return int(C.libraw_raw2image((*C.libraw_data_t)(data)))
}

func Libraw_free_image(data *Data) {
	C.libraw_free_image((*C.libraw_data_t)(data))
}

func Libraw_adjust_sizes_info_only(data *Data) int {
	return int(C.libraw_adjust_sizes_info_only((*C.libraw_data_t)(data)))
}

func Libraw_subtract_black(data *Data) {
	C.libraw_subtract_black((*C.libraw_data_t)(data))
}
