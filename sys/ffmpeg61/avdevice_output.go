package ffmpeg

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavdevice
#include <libavdevice/avdevice.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

// Return the first registered audio output format, or NULL if there are none.
func AVDevice_output_audio_device_first() *AVOutputFormat {
	return (*AVOutputFormat)(C.av_output_audio_device_next((*C.struct_AVOutputFormat)(nil)))
}

// Return the next registered audio output device.
func AVDevice_output_audio_device_next(d *AVOutputFormat) *AVOutputFormat {
	return (*AVOutputFormat)(C.av_output_audio_device_next((*C.struct_AVOutputFormat)(d)))
}

// Return the first registered video output format, or NULL if there are none.
func AVDevice_output_video_device_first() *AVOutputFormat {
	return (*AVOutputFormat)(C.av_output_video_device_next((*C.struct_AVOutputFormat)(nil)))
}

// Return the next registered video output device.
func AVDevice_output_video_device_next(d *AVOutputFormat) *AVOutputFormat {
	return (*AVOutputFormat)(C.av_output_video_device_next((*C.struct_AVOutputFormat)(d)))
}

// List devices. Returns available device names and their parameters, or nil if the
// enumeration of devices is not supported.
// Device format may be nil if device name is set. Call AVDevice_free_list_devices
// to free resources afterwards.
func AVDevice_list_output_sinks(device *AVOutputFormat, device_name string, device_options *AVDictionary) (*AVDeviceInfoList, error) {
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
	if ret := int(C.avdevice_list_output_sinks((*C.struct_AVOutputFormat)(device), cName, dict, &list)); ret < 0 {
		return nil, AVError(ret)
	}

	// Return success
	return (*AVDeviceInfoList)(list), nil
}
