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
// TYPES

type (
	AVAppToDevMessageType C.enum_AVAppToDevMessageType
	AVDevToAppMessageType C.enum_AVDevToAppMessageType
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	// Dummy message.
	AV_APP_TO_DEV_NONE AVAppToDevMessageType = C.AV_APP_TO_DEV_NONE
	// Window size change message.
	// data: AVDeviceRect: new window size.
	AV_APP_TO_DEV_WINDOW_SIZE AVAppToDevMessageType = C.AV_APP_TO_DEV_WINDOW_SIZE
	// Repaint request message.
	// data: AVDeviceRect: area required to be repainted, or NULL if whole area is required to be repainted.
	AV_APP_TO_DEV_WINDOW_REPAINT AVAppToDevMessageType = C.AV_APP_TO_DEV_WINDOW_REPAINT
	// Message sent when device is paused.
	// data: NULL
	AV_APP_TO_DEV_PAUSE AVAppToDevMessageType = C.AV_APP_TO_DEV_PAUSE
	// Message sent when device is unpaused.
	// data: NULL
	AV_APP_TO_DEV_PLAY AVAppToDevMessageType = C.AV_APP_TO_DEV_PLAY
	// Message sent when device play/pause is toggled.
	// data: NULL
	AV_APP_TO_DEV_TOGGLE_PAUSE AVAppToDevMessageType = C.AV_APP_TO_DEV_TOGGLE_PAUSE
	// Volume control message.
	// data: double: new volume with range of 0.0 - 1.0.
	AV_APP_TO_DEV_SET_VOLUME AVAppToDevMessageType = C.AV_APP_TO_DEV_SET_VOLUME
	// Mute control messages.
	// data: NULL.
	AV_APP_TO_DEV_MUTE        AVAppToDevMessageType = C.AV_APP_TO_DEV_MUTE
	AV_APP_TO_DEV_UNMUTE      AVAppToDevMessageType = C.AV_APP_TO_DEV_UNMUTE
	AV_APP_TO_DEV_TOGGLE_MUTE AVAppToDevMessageType = C.AV_APP_TO_DEV_TOGGLE_MUTE
	// Get volume/mute messages.
	// Force the device to send AV_DEV_TO_APP_VOLUME_LEVEL_CHANGED or AV_DEV_TO_APP_MUTE_STATE_CHANGED command respectively.
	// data: NULL.
	AV_APP_TO_DEV_GET_VOLUME AVAppToDevMessageType = C.AV_APP_TO_DEV_GET_VOLUME
	AV_APP_TO_DEV_GET_MUTE   AVAppToDevMessageType = C.AV_APP_TO_DEV_GET_MUTE
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the LIBAVDEVICE_VERSION_INT constant.
func AVDevice_version() uint {
	return uint(C.avdevice_version())
}

// Return the libavdevice build-time configuration.
func AVDevice_configuration() string {
	return C.GoString(C.avdevice_configuration())
}

// Return the libavdevice license.
func AVDevice_license() string {
	return C.GoString(C.avdevice_license())
}

// Initialize libavdevice and register all the input and output devices.
func AVDevice_register_all() {
	C.avdevice_register_all()
}

// Return the first registered audio input format, or NULL if there are none.
func AVDevice_av_input_audio_device_first() *AVInputFormat {
	return (*AVInputFormat)(C.av_input_audio_device_next((*C.struct_AVInputFormat)(nil)))
}

// Return the next registered audio input device.
func (ctx *AVInputFormat) AVDevice_av_input_audio_device_next() *AVInputFormat {
	return (*AVInputFormat)(C.av_input_audio_device_next((*C.struct_AVInputFormat)(ctx)))
}

// Return the first registered video input format, or NULL if there are none.
func AVDevice_av_input_video_device_first() *AVInputFormat {
	return (*AVInputFormat)(C.av_input_video_device_next((*C.struct_AVInputFormat)(nil)))
}

// Return the next registered video input device.
func (ctx *AVInputFormat) AVDevice_av_input_video_device_next() *AVInputFormat {
	return (*AVInputFormat)(C.av_input_video_device_next((*C.struct_AVInputFormat)(ctx)))
}

// Return the first registered audio output format, or NULL if there are none.
func AVDevice_av_output_audio_device_first() *AVOutputFormat {
	return (*AVOutputFormat)(C.av_output_audio_device_next((*C.struct_AVOutputFormat)(nil)))
}

// Return the next registered audio output device.
func (ctx *AVOutputFormat) AVDevice_av_output_audio_device_next() *AVOutputFormat {
	return (*AVOutputFormat)(C.av_output_audio_device_next((*C.struct_AVOutputFormat)(ctx)))
}

// Return the first registered video output format, or NULL if there are none.
func AVDevice_av_output_video_device_first() *AVOutputFormat {
	return (*AVOutputFormat)(C.av_output_video_device_next((*C.struct_AVOutputFormat)(nil)))
}

// Return the next registered video output device.
func (ctx *AVOutputFormat) AVDevice_av_output_video_device_next() *AVOutputFormat {
	return (*AVOutputFormat)(C.av_output_video_device_next((*C.struct_AVOutputFormat)(ctx)))
}

// Send control message from application to device.
func (ctx *AVFormatContext) AVDevice_app_to_dev_control_message(typ AVAppToDevMessageType, data []byte) int {
	return int(C.avdevice_app_to_dev_control_message((*C.struct_AVFormatContext)(ctx), C.enum_AVAppToDevMessageType(typ), unsafe.Pointer(&data[0]), C.size_t(len(data))))
}

// Send control message from device to application.
func (ctx *AVFormatContext) AVDevice_dev_to_app_control_message(typ AVDevToAppMessageType, data []byte) int {
	return int(C.avdevice_dev_to_app_control_message((*C.struct_AVFormatContext)(ctx), C.enum_AVDevToAppMessageType(typ), unsafe.Pointer(&data[0]), C.size_t(len(data))))
}

/*
int 	avdevice_list_devices (struct AVFormatContext *s, AVDeviceInfoList **device_list)
 	List devices. More...

void 	avdevice_free_list_devices (AVDeviceInfoList **device_list)
 	Convenient function to free result of avdevice_list_devices(). More...

int 	avdevice_list_input_sources (const AVInputFormat *device, const char *device_name, AVDictionary *device_options, AVDeviceInfoList **device_list)
 	List devices. More...

int 	avdevice_list_output_sinks (const AVOutputFormat *device, const char *device_name, AVDictionary *device_options, AVDeviceInfoList **device_list)
*/
