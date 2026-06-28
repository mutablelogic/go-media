package libexif

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libexif
#include <stdlib.h>
#include <libexif/exif-data.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	Data       C.ExifData
	DataOption C.ExifDataOption
)

////////////////////////////////////////////////////////////////////////////////
// GLBOALS

const (
	EXIF_DATA_OPTION_NONE                   DataOption = 0
	EXIF_DATA_OPTION_IGNORE_UNKNOWN_TAGS    DataOption = C.EXIF_DATA_OPTION_IGNORE_UNKNOWN_TAGS    // Act as though unknown tags are not present.
	EXIF_DATA_OPTION_FOLLOW_SPECIFICATION   DataOption = C.EXIF_DATA_OPTION_FOLLOW_SPECIFICATION   // Fix the EXIF tags to follow the spec.
	EXIF_DATA_OPTION_DONT_CHANGE_MAKER_NOTE DataOption = C.EXIF_DATA_OPTION_DONT_CHANGE_MAKER_NOTE // Leave the MakerNote alone, which could cause it to be corrupted.
)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - LIFECYCLE

func Exif_data_new() *Data {
	return (*Data)(C.exif_data_new())
}

func Exif_data_new_from_data(data []byte) *Data {
	return (*Data)(C.exif_data_new_from_data((*C.uchar)(unsafe.Pointer(&data[0])), C.uint(len(data))))
}

func Exif_data_new_from_file(filename string) *Data {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))
	return (*Data)(C.exif_data_new_from_file(cfilename))
}

func Exif_data_unref(data *Data) {
	C.exif_data_unref((*C.ExifData)(data))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - LOAD

func Exif_data_load_data(data *Data, input []byte) {
	C.exif_data_load_data((*C.ExifData)(data), (*C.uchar)(unsafe.Pointer(&input[0])), C.uint(len(input)))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - SAVE

func Exif_data_save_data(data *Data) []byte {
	var ptr *C.uchar
	var size C.uint
	C.exif_data_save_data((*C.ExifData)(data), &ptr, &size)
	if ptr == nil || size == 0 {
		return nil
	}
	defer C.free(unsafe.Pointer(ptr))
	return C.GoBytes(unsafe.Pointer(ptr), C.int(size))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - DATA PROPERTIES

func Exif_data_get_byte_order(data *Data) ByteOrder {
	return (ByteOrder)(C.exif_data_get_byte_order((*C.ExifData)(data)))
}

func Exif_data_set_byte_order(data *Data, order ByteOrder) {
	C.exif_data_set_byte_order((*C.ExifData)(data), C.ExifByteOrder(order))
}

func Exif_data_get_data_type(data *Data) DataType {
	return (DataType)(C.exif_data_get_data_type((*C.ExifData)(data)))
}

func Exif_data_set_data_type(data *Data, dtype DataType) {
	C.exif_data_set_data_type((*C.ExifData)(data), C.ExifDataType(dtype))
}

func Exif_data_get_maker_note_data(data *Data) *MakerNoteData {
	return (*MakerNoteData)(C.exif_data_get_mnote_data((*C.ExifData)(data)))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - OPTIONS

func Exif_data_option_get_description(data *Data, option DataOption) string {
	return C.GoString(C.exif_data_option_get_description(C.ExifDataOption(option)))
}

func Exif_data_option_get_name(data *Data, option DataOption) string {
	return C.GoString(C.exif_data_option_get_name(C.ExifDataOption(option)))
}

func Exif_data_set_option(data *Data, option DataOption) {
	C.exif_data_set_option((*C.ExifData)(data), C.ExifDataOption(option))
}

func Exif_data_unset_option(data *Data, option DataOption) {
	C.exif_data_unset_option((*C.ExifData)(data), C.ExifDataOption(option))
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS - OTHER

func Exif_data_fix(data *Data) {
	C.exif_data_fix((*C.ExifData)(data))
}

func Exif_data_dump(data *Data) {
	C.exif_data_dump((*C.ExifData)(data))
}
