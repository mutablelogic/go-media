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
// INIT

// Device-specific demuxers/muxers (avfoundation, v4l2, dshow, alsa, ...)
// live in libavdevice, a separate library from the demuxers/muxers
// libavformat auto-registers on load - per avdevice.h's own doc comment,
// "[t]o use libavdevice, simply call avdevice_register_all()". Without this,
// AVFormat_find_input_format/AVFormat_find_output_format can't find them by
// name, even though this package is already always linked against
// libavdevice (see the other avdevice_*.go files) and even though the
// device *iterators* (AVDevice_input_audio_device_first et al, used by
// ListDevices) work regardless, since those walk libavdevice's own
// compiled-in list rather than the shared registered-format one.
func init() {
	C.avdevice_register_all()
}

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
