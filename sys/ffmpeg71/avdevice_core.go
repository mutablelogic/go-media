package ffmpeg

import (
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavdevice
#include <libavdevice/avdevice.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

func AVDevice_list_devices(ctx *AVFormatContext) (*AVDeviceInfoList, error) {
	var list *C.struct_AVDeviceInfoList
	if ret := int(C.avdevice_list_devices((*C.struct_AVFormatContext)(unsafe.Pointer(ctx)), &list)); ret < 0 {
		return nil, AVError(ret)
	} else if ret == 0 {
		return nil, nil
	} else {
		return (*AVDeviceInfoList)(list), nil
	}
}

func AVDevice_free_list_devices(device_list *AVDeviceInfoList) {
	C.avdevice_free_list_devices((**C.struct_AVDeviceInfoList)(unsafe.Pointer(&device_list)))
}
