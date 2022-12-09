package ffmpeg

import (
	"fmt"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavutil
#include <libavutil/channel_layout.h>
#include <stdlib.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	cBufSize = 1024
)

////////////////////////////////////////////////////////////////////////////////
// STRINFIGY

func (l *AVChannelLayout) String() string {
	str := "<AVChannelLayout"
	if description, err := AVUtil_av_channel_layout_describe(l); err == nil {
		str += fmt.Sprintf(" description=%q", description)
	}
	nb_channels := AVUtil_av_get_channel_layout_nb_channels(l)
	for i := 0; i < nb_channels; i++ {
		ch := AVUtil_av_channel_layout_channel_from_index(l, i)
		if str, err := AVUtil_av_channel_name(ch); err == nil {
			str += fmt.Sprintf(" ch_%d=%q", i, str)
		} else {
			str += fmt.Sprintf(" ch_%d=%v", i, ch)
		}
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Get the name of a given channel.
func AVUtil_av_channel_name(channel AVChannel) (string, error) {
	var buf [cBufSize]C.char
	if n := C.av_channel_name(&buf[0], cBufSize, C.enum_AVChannel(channel)); n < 0 {
		return "", AVError(n)
	} else {
		return C.GoString(&buf[0]), nil
	}
}

// Get a human readable string describing a given channel.
func AVUtil_av_channel_description(channel AVChannel) (string, error) {
	var buf [cBufSize]C.char
	if n := C.av_channel_description(&buf[0], cBufSize, C.enum_AVChannel(channel)); n < 0 {
		return "", AVError(n)
	} else {
		return C.GoString(&buf[0]), nil
	}
}

// This is the inverse function of av_channel_name.
func AVUtil_av_channel_from_string(name string) AVChannel {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return AVChannel(C.av_channel_from_string(cName))
}

// Iterate over all standard channel layouts.
func AVUtil_av_channel_layout_standard(iterator *uintptr) *AVChannelLayout {
	return (*AVChannelLayout)(C.av_channel_layout_standard((*unsafe.Pointer)(unsafe.Pointer(iterator))))
}

// Iterate over all standard channel layouts.
func AVUtil_av_channel_layout_describe(channel_layout *AVChannelLayout) (string, error) {
	var buf [cBufSize]C.char
	if n := C.av_channel_layout_describe((*C.struct_AVChannelLayout)(channel_layout), &buf[0], cBufSize); n < 0 {
		return "", AVError(n)
	} else {
		return C.GoString(&buf[0]), nil
	}
}

// Get the default channel layout for a given number of channels.
func AVUtil_av_channel_layout_default(ch_layout *AVChannelLayout, nb_channels int) {
	C.av_channel_layout_default((*C.struct_AVChannelLayout)(ch_layout), C.int(nb_channels))
}

// Return channel layout from a description
func AVUtil_av_channel_layout_from_string(ch_layout *AVChannelLayout, str string) error {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))
	if err := AVError(C.av_channel_layout_from_string((*C.struct_AVChannelLayout)(ch_layout), cStr)); err < 0 {
		return err
	} else {
		return nil
	}
}

// Free any allocated data in the channel layout and reset the channel count to 0.
func AVUtil_av_channel_layout_uninit(ch_layout *AVChannelLayout) {
	C.av_channel_layout_uninit((*C.struct_AVChannelLayout)(ch_layout))
}

// Get the channel with the given index in a channel layout.
func AVUtil_av_channel_layout_channel_from_index(ch_layout *AVChannelLayout, index int) AVChannel {
	return AVChannel(C.av_channel_layout_channel_from_index((*C.struct_AVChannelLayout)(ch_layout), C.uint(index)))
}

// Get the index of a given channel in a channel layout.
func AVUtil_av_channel_layout_index_from_channel(ch_layout *AVChannelLayout, channel AVChannel) int {
	return int(C.av_channel_layout_index_from_channel((*C.struct_AVChannelLayout)(ch_layout), C.enum_AVChannel(channel)))
}

// Return number of channels
func AVUtil_av_get_channel_layout_nb_channels(ch_layout *AVChannelLayout) int {
	return int((*C.struct_AVChannelLayout)(ch_layout).nb_channels)
}

// Check whether a channel layout is valid
func AVUtil_av_channel_layout_check(ch_layout *AVChannelLayout) bool {
	return C.av_channel_layout_check((*C.struct_AVChannelLayout)(ch_layout)) != 0
}
