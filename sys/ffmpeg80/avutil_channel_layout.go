package ffmpeg

import (
	"encoding/json"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/channel_layout.h>

AVChannelLayout _AV_CHANNEL_LAYOUT_MONO = AV_CHANNEL_LAYOUT_MONO;
AVChannelLayout _AV_CHANNEL_LAYOUT_STEREO = AV_CHANNEL_LAYOUT_STEREO;
AVChannelLayout _AV_CHANNEL_LAYOUT_2POINT1 = AV_CHANNEL_LAYOUT_2POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_2_1 = AV_CHANNEL_LAYOUT_2_1;
AVChannelLayout _AV_CHANNEL_LAYOUT_SURROUND = AV_CHANNEL_LAYOUT_SURROUND;
AVChannelLayout _AV_CHANNEL_LAYOUT_3POINT1 = AV_CHANNEL_LAYOUT_3POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_4POINT0 = AV_CHANNEL_LAYOUT_4POINT0;
AVChannelLayout _AV_CHANNEL_LAYOUT_4POINT1 = AV_CHANNEL_LAYOUT_4POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_2_2 = AV_CHANNEL_LAYOUT_2_2;
AVChannelLayout _AV_CHANNEL_LAYOUT_QUAD = AV_CHANNEL_LAYOUT_QUAD;
AVChannelLayout _AV_CHANNEL_LAYOUT_5POINT0 = AV_CHANNEL_LAYOUT_5POINT0;
AVChannelLayout _AV_CHANNEL_LAYOUT_5POINT1 = AV_CHANNEL_LAYOUT_5POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_5POINT0_BACK = AV_CHANNEL_LAYOUT_5POINT0_BACK;
AVChannelLayout _AV_CHANNEL_LAYOUT_5POINT1_BACK = AV_CHANNEL_LAYOUT_5POINT1_BACK;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT0 = AV_CHANNEL_LAYOUT_6POINT0;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT0_FRONT = AV_CHANNEL_LAYOUT_6POINT0_FRONT;
AVChannelLayout _AV_CHANNEL_LAYOUT_HEXAGONAL = AV_CHANNEL_LAYOUT_HEXAGONAL;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT1 = AV_CHANNEL_LAYOUT_6POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT1_BACK = AV_CHANNEL_LAYOUT_6POINT1_BACK;
AVChannelLayout _AV_CHANNEL_LAYOUT_6POINT1_FRONT = AV_CHANNEL_LAYOUT_6POINT1_FRONT;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT0 = AV_CHANNEL_LAYOUT_7POINT0;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT0_FRONT = AV_CHANNEL_LAYOUT_7POINT0_FRONT;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT1 = AV_CHANNEL_LAYOUT_7POINT1;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT1_WIDE = AV_CHANNEL_LAYOUT_7POINT1_WIDE;
AVChannelLayout _AV_CHANNEL_LAYOUT_7POINT1_WIDE_BACK = AV_CHANNEL_LAYOUT_7POINT1_WIDE_BACK;
AVChannelLayout _AV_CHANNEL_LAYOUT_OCTAGONAL = AV_CHANNEL_LAYOUT_OCTAGONAL;
AVChannelLayout _AV_CHANNEL_LAYOUT_HEXADECAGONAL = AV_CHANNEL_LAYOUT_HEXADECAGONAL;
AVChannelLayout _AV_CHANNEL_LAYOUT_STEREO_DOWNMIX = AV_CHANNEL_LAYOUT_STEREO_DOWNMIX;
AVChannelLayout _AV_CHANNEL_LAYOUT_22POINT2 = AV_CHANNEL_LAYOUT_22POINT2;
AVChannelLayout _AV_CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER = AV_CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER;
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVChannel       C.enum_AVChannel
	AVChannelLayout C.AVChannelLayout
	AVChannelOrder  C.enum_AVChannelOrder
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	cBufSize = 32
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ch AVChannelLayout) MarshalJSON() ([]byte, error) {
	if ch.NumChannels() == 0 {
		return json.Marshal(nil)
	} else if str, err := AVUtil_channel_layout_describe(&ch); err != nil {
		return nil, err
	} else {
		return json.Marshal(str)
	}
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

// Get the name of a given channel.
func AVUtil_channel_name(channel AVChannel) (string, error) {
	var buf [cBufSize]C.char
	if n := C.av_channel_name(&buf[0], cBufSize, C.enum_AVChannel(channel)); n < 0 {
		return "", AVError(n)
	} else {
		return C.GoString(&buf[0]), nil
	}
}

// Get a human readable string describing a given channel.
func AVUtil_channel_description(channel AVChannel) (string, error) {
	var buf [cBufSize]C.char
	if n := C.av_channel_description(&buf[0], cBufSize, C.enum_AVChannel(channel)); n < 0 {
		return "", AVError(n)
	} else {
		return C.GoString(&buf[0]), nil
	}
}

// This is the inverse function of av_channel_name.
func AVUtil_channel_from_string(name string) AVChannel {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return AVChannel(C.av_channel_from_string(cName))
}

// Iterate over all standard channel layouts.
func AVUtil_channel_layout_standard(iterator *uintptr) *AVChannelLayout {
	return (*AVChannelLayout)(C.av_channel_layout_standard((*unsafe.Pointer)(unsafe.Pointer(iterator))))
}

// Get a human-readable string describing the channel layout properties.
func AVUtil_channel_layout_describe(channel_layout *AVChannelLayout) (string, error) {
	var buf [cBufSize]C.char
	if n := C.av_channel_layout_describe((*C.struct_AVChannelLayout)(channel_layout), &buf[0], cBufSize); n < 0 {
		return "", AVError(n)
	} else {
		return C.GoString(&buf[0]), nil
	}
}

// Get the default channel layout for a given number of channels.
func AVUtil_channel_layout_default(ch_layout *AVChannelLayout, nb_channels int) {
	C.av_channel_layout_default((*C.struct_AVChannelLayout)(ch_layout), C.int(nb_channels))
}

// Return channel layout from a description
func AVUtil_channel_layout_from_string(ch_layout *AVChannelLayout, str string) error {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))
	if err := AVError(C.av_channel_layout_from_string((*C.struct_AVChannelLayout)(ch_layout), cStr)); err < 0 {
		return err
	} else {
		return nil
	}
}

// Free any allocated data in the channel layout and reset the channel count to 0.
func AVUtil_channel_layout_uninit(ch_layout *AVChannelLayout) {
	C.av_channel_layout_uninit((*C.struct_AVChannelLayout)(ch_layout))
}

// Get the channel with the given index in a channel layout.
func AVUtil_channel_layout_channel_from_index(ch_layout *AVChannelLayout, index int) AVChannel {
	return AVChannel(C.av_channel_layout_channel_from_index((*C.struct_AVChannelLayout)(ch_layout), C.uint(index)))
}

// Get the index of a given channel in a channel layout.
func AVUtil_channel_layout_index_from_channel(ch_layout *AVChannelLayout, channel AVChannel) int {
	return int(C.av_channel_layout_index_from_channel((*C.struct_AVChannelLayout)(ch_layout), C.enum_AVChannel(channel)))
}

// Return number of channels
func AVUtil_get_channel_layout_nb_channels(ch_layout *AVChannelLayout) int {
	return int((*C.struct_AVChannelLayout)(ch_layout).nb_channels)
}

// Check whether a channel layout is valid
func AVUtil_channel_layout_check(ch_layout *AVChannelLayout) bool {
	return C.av_channel_layout_check((*C.struct_AVChannelLayout)(ch_layout)) != 0
}

// Check whether two channel layouts are semantically the same
func AVUtil_channel_layout_compare(a *AVChannelLayout, b *AVChannelLayout) bool {
	ret := C.av_channel_layout_compare((*C.struct_AVChannelLayout)(a), (*C.struct_AVChannelLayout)(b))
	return ret == 0
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (ctx AVChannelLayout) NumChannels() int {
	return int(ctx.nb_channels)
}

func (ctx AVChannelLayout) Order() AVChannelOrder {
	return AVChannelOrder(ctx.order)
}

func (o AVChannelOrder) String() string {
	switch o {
	case 0:
		return "unspec"
	case 1:
		return "native"
	case 2:
		return "custom"
	case 3:
		return "ambisonic"
	default:
		return "unknown"
	}
}

////////////////////////////////////////////////////////////////////////////////
// CHANNEL LAYOUT CONSTANTS

func AV_CHANNEL_LAYOUT_MONO() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_MONO)
}

func AV_CHANNEL_LAYOUT_STEREO() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_STEREO)
}

func AV_CHANNEL_LAYOUT_2POINT1() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_2POINT1)
}

func AV_CHANNEL_LAYOUT_2_1() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_2_1)
}

func AV_CHANNEL_LAYOUT_SURROUND() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_SURROUND)
}

func AV_CHANNEL_LAYOUT_3POINT1() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_3POINT1)
}

func AV_CHANNEL_LAYOUT_4POINT0() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_4POINT0)
}

func AV_CHANNEL_LAYOUT_4POINT1() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_4POINT1)
}

func AV_CHANNEL_LAYOUT_2_2() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_2_2)
}

func AV_CHANNEL_LAYOUT_QUAD() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_QUAD)
}

func AV_CHANNEL_LAYOUT_5POINT0() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_5POINT0)
}

func AV_CHANNEL_LAYOUT_5POINT1() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_5POINT1)
}

func AV_CHANNEL_LAYOUT_5POINT0_BACK() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_5POINT0_BACK)
}

func AV_CHANNEL_LAYOUT_5POINT1_BACK() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_5POINT1_BACK)
}

func AV_CHANNEL_LAYOUT_6POINT0() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT0)
}

func AV_CHANNEL_LAYOUT_6POINT0_FRONT() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT0_FRONT)
}

func AV_CHANNEL_LAYOUT_HEXAGONAL() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_HEXAGONAL)
}

func AV_CHANNEL_LAYOUT_6POINT1() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT1)
}

func AV_CHANNEL_LAYOUT_6POINT1_BACK() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT1_BACK)
}

func AV_CHANNEL_LAYOUT_6POINT1_FRONT() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_6POINT1_FRONT)
}

func AV_CHANNEL_LAYOUT_7POINT0() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT0)
}

func AV_CHANNEL_LAYOUT_7POINT0_FRONT() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT0_FRONT)
}

func AV_CHANNEL_LAYOUT_7POINT1() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT1)
}

func AV_CHANNEL_LAYOUT_7POINT1_WIDE() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT1_WIDE)
}

func AV_CHANNEL_LAYOUT_7POINT1_WIDE_BACK() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_7POINT1_WIDE_BACK)
}

func AV_CHANNEL_LAYOUT_OCTAGONAL() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_OCTAGONAL)
}

func AV_CHANNEL_LAYOUT_HEXADECAGONAL() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_HEXADECAGONAL)
}

func AV_CHANNEL_LAYOUT_STEREO_DOWNMIX() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_STEREO_DOWNMIX)
}

func AV_CHANNEL_LAYOUT_22POINT2() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_22POINT2)
}

func AV_CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER() AVChannelLayout {
	return AVChannelLayout(C._AV_CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER)
}
