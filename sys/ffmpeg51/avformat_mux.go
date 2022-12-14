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

// Allocate an AVFormatContext for an output format.
func AVFormat_alloc_output_context2(oformat *AVOutputFormat, format string, filename string) (*AVFormatContext, error) {
	var cFilename, cFormat *C.char
	if format != "" {
		cFormat = C.CString(format)
	}
	if filename != "" {
		cFilename = C.CString(filename)
	}
	defer C.free(unsafe.Pointer(cFilename))
	defer C.free(unsafe.Pointer(cFormat))
	var ctx *AVFormatContext
	if err := AVError(C.avformat_alloc_output_context2((**C.struct_AVFormatContext)(unsafe.Pointer(&ctx)), (*C.struct_AVOutputFormat)(oformat), cFormat, cFilename)); err != 0 {
		return nil, err
	} else {
		return ctx, nil
	}
}

// Allocate the stream private data and write the stream header to an output media file.
func AVFormat_write_header(ctx *AVFormatContext, options **AVDictionary) error {
	if err := AVError(C.avformat_write_header((*C.struct_AVFormatContext)(ctx), (**C.struct_AVDictionary)(unsafe.Pointer(options)))); err < 0 {
		return err
	} else {
		return nil
	}
	// TODO:
	// AVSTREAM_INIT_IN_WRITE_HEADER
	// AVSTREAM_INIT_IN_INIT_OUTPUT
}

// Write a packet to an output media file. Returns true if flushed and there is
// no more data to flush.
func AVFormat_write_frame(ctx *AVFormatContext, pkt *AVPacket) (bool, error) {
	if err := AVError(C.av_write_frame((*C.struct_AVFormatContext)(ctx), (*C.struct_AVPacket)(pkt))); err < 0 {
		return false, err
	} else if err == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

// Write a packet to an output media file ensuring correct interleaving.
func AVFormat_interleaved_write_frame(ctx *AVFormatContext, pkt *AVPacket) error {
	if err := AVError(C.av_interleaved_write_frame((*C.struct_AVFormatContext)(ctx), (*C.struct_AVPacket)(pkt))); err != 0 {
		return err
	} else {
		return nil
	}
}

// Write the stream trailer to an output media file and free the file private data.
func AVFormat_write_trailer(ctx *AVFormatContext) error {
	if err := AVError(C.av_write_trailer((*C.struct_AVFormatContext)(ctx))); err != 0 {
		return err
	} else {
		return nil
	}
}
