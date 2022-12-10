package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
*/
import "C"
import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Find AVInputFormat based on the short name of the input format.
func AVFormat_av_find_input_format(short_name string) *AVInputFormat {
	return (*AVInputFormat)(C.av_find_input_format(C.CString(short_name)))
}

// Open an input stream and read the header.
func AVFormat_open_input(ctx **AVFormatContext, url string, fmt *AVInputFormat, options **AVDictionary) error {
	if err := AVError(C.avformat_open_input((**C.struct_AVFormatContext)(unsafe.Pointer(ctx)), C.CString(url), (*C.struct_AVInputFormat)(fmt), (**C.struct_AVDictionary)(unsafe.Pointer(options)))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Close an opened input AVFormatContext. Free it and all its contents and set *s to NULL.
func AVFormat_close_input(ctx **AVFormatContext) {
	C.avformat_close_input((**C.struct_AVFormatContext)(unsafe.Pointer(ctx)))
}
