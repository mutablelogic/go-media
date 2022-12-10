package ffmpeg

import (
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Find AVInputFormat based on the short name of the input format.
func AVFormat_av_find_input_format(short_name string) *AVInputFormat {
	return (*AVInputFormat)(C.av_find_input_format(C.CString(short_name)))
}

// Open an input stream and read the header.
func AVFormat_open_input(ctx **AVFormatContext, url string, input_fmt *AVInputFormat, options **AVDictionary) error {
	if err := AVError(C.avformat_open_input((**C.struct_AVFormatContext)(unsafe.Pointer(ctx)), C.CString(url), (*C.struct_AVInputFormat)(input_fmt), (**C.struct_AVDictionary)(unsafe.Pointer(options)))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Close an opened input AVFormatContext. Free it and all its contents and set *s to NULL.
func AVFormat_close_input(ctx **AVFormatContext) {
	C.avformat_close_input((**C.struct_AVFormatContext)(unsafe.Pointer(ctx)))
}

// Read packets of a media file to get stream information.
func AVFormat_find_stream_info(ctx *AVFormatContext, options **AVDictionary) error {
	if err := AVError(C.avformat_find_stream_info((*C.struct_AVFormatContext)(ctx), (**C.struct_AVDictionary)(unsafe.Pointer(options)))); err < 0 {
		return err
	} else {
		return nil
	}
}
