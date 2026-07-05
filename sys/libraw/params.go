package libraw

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libraw
#include <libraw/libraw.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - OUTPUT PARAM SETTERS

func Libraw_set_demosaic(data *Data, value int) {
	C.libraw_set_demosaic((*C.libraw_data_t)(data), C.int(value))
}

func Libraw_set_output_color(data *Data, value int) {
	C.libraw_set_output_color((*C.libraw_data_t)(data), C.int(value))
}

func Libraw_set_output_bps(data *Data, value int) {
	C.libraw_set_output_bps((*C.libraw_data_t)(data), C.int(value))
}

func Libraw_set_output_tif(data *Data, value int) {
	C.libraw_set_output_tif((*C.libraw_data_t)(data), C.int(value))
}

func Libraw_set_no_auto_bright(data *Data, value int) {
	C.libraw_set_no_auto_bright((*C.libraw_data_t)(data), C.int(value))
}

func Libraw_set_bright(data *Data, value float32) {
	C.libraw_set_bright((*C.libraw_data_t)(data), C.float(value))
}

func Libraw_set_highlight(data *Data, value int) {
	C.libraw_set_highlight((*C.libraw_data_t)(data), C.int(value))
}

func Libraw_set_gamma(data *Data, index int, value float32) {
	C.libraw_set_gamma((*C.libraw_data_t)(data), C.int(index), C.float(value))
}

func Libraw_set_user_mul(data *Data, index int, value float32) {
	C.libraw_set_user_mul((*C.libraw_data_t)(data), C.int(index), C.float(value))
}

func Libraw_set_adjust_maximum_thr(data *Data, value float32) {
	C.libraw_set_adjust_maximum_thr((*C.libraw_data_t)(data), C.float(value))
}

func Libraw_set_fbdd_noiserd(data *Data, value int) {
	C.libraw_set_fbdd_noiserd((*C.libraw_data_t)(data), C.int(value))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - COLOR GETTERS

func Libraw_get_cam_mul(data *Data, index int) float32 {
	return float32(C.libraw_get_cam_mul((*C.libraw_data_t)(data), C.int(index)))
}

func Libraw_get_pre_mul(data *Data, index int) float32 {
	return float32(C.libraw_get_pre_mul((*C.libraw_data_t)(data), C.int(index)))
}

func Libraw_get_rgb_cam(data *Data, index1, index2 int) float32 {
	return float32(C.libraw_get_rgb_cam((*C.libraw_data_t)(data), C.int(index1), C.int(index2)))
}

func Libraw_get_color_maximum(data *Data) int {
	return int(C.libraw_get_color_maximum((*C.libraw_data_t)(data)))
}
