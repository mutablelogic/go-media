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
func AVFormat_open_reader(reader *AVIOContextEx, format *AVInputFormat, options *AVDictionary) (*AVFormatContext, error) {
	var opts **C.struct_AVDictionary
	if options != nil {
		opts = &options.ctx
	}

	// Allocate a context
	ctx := AVFormat_alloc_context()
	if ctx == nil {
		return nil, AVError(syscall.ENOMEM)
	} else {
		ctx.pb = (*C.struct_AVIOContext)(unsafe.Pointer(reader.AVIOContext))
	}

	// Open the stream
	if err := AVError(C.avformat_open_input((**C.struct_AVFormatContext)(unsafe.Pointer(&ctx)), nil, (*C.struct_AVInputFormat)(format), opts)); err != 0 {
		return nil, err
	} else {
		return ctx, nil
	}
}

// Open an input stream from a URL and read the header.
func AVFormat_open_url(url string, format *AVInputFormat, options *AVDictionary) (*AVFormatContext, error) {
	var opts **C.struct_AVDictionary
	if options != nil {
		opts = &options.ctx
	}

	// Create a C string for the URL
	cUrl := C.CString(url)
	defer C.free(unsafe.Pointer(cUrl))

	// Allocate a context
	ctx := AVFormat_alloc_context()
	if ctx == nil {
		return nil, AVError(syscall.ENOMEM)
	}

	// Open the URL
	if err := AVError(C.avformat_open_input((**C.struct_AVFormatContext)(unsafe.Pointer(&ctx)), cUrl, (*C.struct_AVInputFormat)(format), opts)); err != 0 {
		return nil, err
	}

	// Return success
	return ctx, nil
}

// Open an input stream from a device.
func AVFormat_open_device(format *AVInputFormat, options *AVDictionary) (*AVFormatContext, error) {
	var opts **C.struct_AVDictionary
	if options != nil {
		opts = &options.ctx
	}

	// Allocate a context
	ctx := AVFormat_alloc_context()
	if ctx == nil {
		return nil, AVError(syscall.ENOMEM)
	}

	// Open the device
	if err := AVError(C.avformat_open_input((**C.struct_AVFormatContext)(unsafe.Pointer(&ctx)), nil, (*C.struct_AVInputFormat)(format), opts)); err != 0 {
		return nil, err
	}

	// Return success
	return ctx, nil
}

// Close an opened input AVFormatContext, free it and all its contents.
func AVFormat_close_input(ctx *AVFormatContext) {
	C.avformat_close_input((**C.struct_AVFormatContext)(unsafe.Pointer(&ctx)))
}

// Read packets of a media file to get stream information.
func AVFormat_find_stream_info(ctx *AVFormatContext, options *AVDictionary) error {
	var opts **C.struct_AVDictionary
	if options != nil {
		opts = &options.ctx
	}
	if err := AVError(C.avformat_find_stream_info((*C.struct_AVFormatContext)(ctx), opts)); err != 0 {
		return err
	}
	// Return success
	return nil
}

// Read a frame from the input stream. Return io.EOF if the end of the stream is reached.
func AVFormat_read_frame(ctx *AVFormatContext, packet *AVPacket) error {
	if err := AVError(C.av_read_frame((*C.struct_AVFormatContext)(ctx), (*C.struct_AVPacket)(packet))); err < 0 {
		if err == AVERROR_EOF {
			return io.EOF
		} else if err.IsErrno(syscall.EAGAIN) {
			return syscall.EAGAIN
		} else {
			return err
		}
	}
	// Return success
	return nil
}

// Discard all internally buffered data.
func AVFormat_flush(ctx *AVFormatContext) error {
	if err := AVError(C.avformat_flush((*C.struct_AVFormatContext)(ctx))); err != 0 {
		return err
	}
	// Return success
	return nil
}

// Start playing a network-based stream (e.g. RTSP stream) at the current position.
func AVFormat_read_play(ctx *AVFormatContext) error {
	if err := AVError(C.av_read_play((*C.struct_AVFormatContext)(ctx))); err != 0 {
		return err
	}
	// Return success
	return nil
}

// Pause a network-based stream (e.g. RTSP stream).
func AVFormat_read_pause(ctx *AVFormatContext) error {
	if err := AVError(C.av_read_pause((*C.struct_AVFormatContext)(ctx))); err != 0 {
		return err
	}
	// Return success
	return nil
}
