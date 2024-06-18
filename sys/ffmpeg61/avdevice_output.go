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

// List devices. Returns available device names and their parameters.
// device format may be nil if device name is set.
func AVDevice_list_output_sinks(device *AVOutputFormat, device_name string, device_options *AVDictionary) (*AVDeviceInfoList, error) {
	cName := C.CString(device_name)
	defer C.free(unsafe.Pointer(cName))

	var dict *C.struct_AVDictionary
	if device_options != nil {
		dict = device_options.ctx
	}
	var list *C.struct_AVDeviceInfoList
	if ret := int(C.avdevice_list_output_sinks((*C.struct_AVOutputFormat)(device), cName, dict, &list)); ret < 0 {
		return nil, AVError(ret)
	} else if ret == 0 {
		return nil, nil
	} else {
		return (*AVDeviceInfoList)(list), nil
	}
}
