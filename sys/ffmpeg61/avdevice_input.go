package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavdevice
#include <libavdevice/avdevice.h>
*/
import "C"
import "unsafe"

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

// List devices. Returns available device names and their parameters.
// device format may be nil if device name is set.
func AVDevice_list_input_sources(device *AVInputFormat, device_name string, device_options *AVDictionary) (*AVDeviceInfoList, error) {
	cName := C.CString(device_name)
	defer C.free(unsafe.Pointer(cName))

	var dict *C.struct_AVDictionary
	if device_options != nil {
		dict = device_options.ctx
	}
	var list *C.struct_AVDeviceInfoList
	if ret := int(C.avdevice_list_input_sources((*C.struct_AVInputFormat)(device), cName, dict, &list)); ret < 0 {
		return nil, AVError(ret)
	} else if ret == 0 {
		return nil, nil
	} else {
		return (*AVDeviceInfoList)(list), nil
	}
}
