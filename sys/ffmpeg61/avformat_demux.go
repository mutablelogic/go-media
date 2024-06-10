package ffmpeg

import (
	"io"
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

// Open an input stream from a URL and read the header.
func AVFormat_open_url(url string, format *AVInputFormat, options **AVDictionary) (*AVFormatContext, error) {
	// Create a C string for the URL
	cUrl := C.CString(url)
	defer C.free(unsafe.Pointer(cUrl))

	// Allocate a context
	ctx := AVFormat_alloc_context()
	if ctx == nil {
		return nil, AVError(syscall.ENOMEM)
	}

	// Open the URL
	if err := AVError(C.avformat_open_input((**C.struct_AVFormatContext)(unsafe.Pointer(&ctx)), cUrl, (*C.struct_AVInputFormat)(format), (**C.struct_AVDictionary)(unsafe.Pointer(options)))); err != 0 {
		return nil, err
	}

	// Return success
	return ctx, nil
}

// Open an input stream from a device.
func AVFormat_open_device(format *AVInputFormat, options **AVDictionary) (*AVFormatContext, error) {
	return AVFormat_open_url("", format, options)
}

// Read a frame from the input stream.
func AVFormat_av_read_frame(ctx *AVFormatContext, packet *AVPacket) error {
	if err := AVError(C.av_read_frame((*C.struct_AVFormatContext)(ctx), (*C.struct_AVPacket)(packet))); err < 0 {
		if err == AVERROR_EOF {
			return io.EOF
		} else {
			return err
		}
	}
	// Return success
	return nil
}
