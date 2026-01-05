package ffmpeg

import (
	"encoding/json"
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
// TYPES

type (
	AVAppToDevMessageType C.enum_AVAppToDevMessageType
	AVDevToAppMessageType C.enum_AVDevToAppMessageType
	AVDeviceInfoList      C.struct_AVDeviceInfoList
	AVDeviceInfo          C.struct_AVDeviceInfo
)

////////////////////////////////////////////////////////////////////////////////
// JSON OUTPUT

func (ctx *AVDeviceInfoList) MarshalJSON() ([]byte, error) {
	type jsonAVDeviceInfoList struct {
		Devices       []*AVDeviceInfo `json:"devices"`
		DefaultDevice int             `json:"default_device"`
	}
	return json.Marshal(jsonAVDeviceInfoList{
		Devices:       ctx.Devices(),
		DefaultDevice: int(ctx.default_device),
	})
}

func (ctx *AVDeviceInfo) MarshalJSON() ([]byte, error) {
	type jsonAVDeviceInfo struct {
		Name        string        `json:"device_name"`
		Description string        `json:"device_description"`
		MediaTypes  []AVMediaType `json:"media_types"`
	}
	return json.Marshal(jsonAVDeviceInfo{
		Name:        C.GoString(ctx.device_name),
		Description: C.GoString(ctx.device_description),
		MediaTypes:  ctx.MediaTypes(),
	})
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVDeviceInfoList) String() string {
	return marshalToString(ctx)
}

func (ctx *AVDeviceInfo) String() string {
	return marshalToString(ctx)
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

// list of autodetected devices
func (ctx *AVDeviceInfoList) Devices() []*AVDeviceInfo {
	if ctx == nil || ctx.nb_devices == 0 || ctx.devices == nil {
		return nil
	}
	return cAVDeviceInfoSlice(unsafe.Pointer(ctx.devices), ctx.nb_devices)
}

// number of autodetected devices
func (ctx *AVDeviceInfoList) NumDevices() int {
	if ctx == nil {
		return 0
	}
	return int(ctx.nb_devices)
}

// index of default device or -1 if no default
func (ctx *AVDeviceInfoList) Default() int {
	if ctx == nil {
		return -1
	}
	return int(ctx.default_device)
}

func (ctx *AVDeviceInfo) Name() string {
	return C.GoString(ctx.device_name)
}

func (ctx *AVDeviceInfo) Description() string {
	return C.GoString(ctx.device_description)
}

func (ctx *AVDeviceInfo) MediaTypes() []AVMediaType {
	return cAVMediaTypeSlice(unsafe.Pointer(ctx.media_types), ctx.nb_media_types)
}
