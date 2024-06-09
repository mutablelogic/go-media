package ffmpeg

import (
	"syscall"
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

// Open an input stream and read the header.
func AVFormat_open_reader(reader *AVIOContextEx, format *AVInputFormat, options **AVDictionary) (*AVFormatContext, error) {
	ctx := AVFormat_alloc_context()
	if ctx == nil {
		return nil, AVError(syscall.ENOMEM)
	} else {
		ctx.pb = (*C.struct_AVIOContext)(unsafe.Pointer(reader.AVIOContext))
	}
	if err := AVError(C.avformat_open_input((**C.struct_AVFormatContext)(unsafe.Pointer(&ctx)), nil, (*C.struct_AVInputFormat)(format), (**C.struct_AVDictionary)(unsafe.Pointer(options)))); err != 0 {
		return nil, err
	} else {
		return ctx, nil
	}
}
