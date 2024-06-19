package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavdevice
#include <libavdevice/avdevice.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

// Return the first registered audio input format, or NULL if there are none.
func AVDevice_input_audio_device_first() *AVInputFormat {
	return (*AVInputFormat)(C.av_input_audio_device_next((*C.struct_AVInputFormat)(nil)))
}

// Return the next registered audio input device.
func AVDevice_input_audio_device_next(d *AVInputFormat) *AVInputFormat {
	return (*AVInputFormat)(C.av_input_audio_device_next((*C.struct_AVInputFormat)(d)))
}

// Return the first registered video input format, or NULL if there are none.
func AVDevice_input_video_device_first() *AVInputFormat {
	return (*AVInputFormat)(C.av_input_video_device_next((*C.struct_AVInputFormat)(nil)))
}

// Return the next registered video input device.
func AVDevice_input_video_device_next(d *AVInputFormat) *AVInputFormat {
	return (*AVInputFormat)(C.av_input_video_device_next((*C.struct_AVInputFormat)(d)))
}

// List devices. Returns available device names and their parameters, or nil if the
// enumeration of devices is not supported.
// Device format may be nil if device name is set. Call AVDevice_free_list_devices
// to free resources afterwards.
func AVDevice_list_input_sources(device *AVInputFormat, device_name string, device_options *AVDictionary) (*AVDeviceInfoList, error) {
	// Return nil if the devices does not implement the get_device_list function
	if device != nil && device.get_device_list == nil {
		return nil, nil
	}

	// Prepare name
	cName := C.CString(device_name)
	defer C.free(unsafe.Pointer(cName))

	// Prepare dictionary
	var dict *C.struct_AVDictionary
	if device_options != nil {
		dict = device_options.ctx
	}

	// Get list
	var list *C.struct_AVDeviceInfoList
	if ret := int(C.avdevice_list_input_sources((*C.struct_AVInputFormat)(device), cName, dict, &list)); ret < 0 {
		fmt.Println("C list", device, cName, dict, list, ret)
		return nil, AVError(ret)
	}

	// Return success
	return (*AVDeviceInfoList)(list), nil
}
