package ffmpeg

import (
	"errors"
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

// Open an output stream without managing a file.
func AVFormat_open_writer(writer *AVIOContextEx, format *AVOutputFormat, filename string) (*AVFormatContext, error) {
	var ctx *AVFormatContext

	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))
	if err := AVError(C.avformat_alloc_output_context2((**C.struct_AVFormatContext)(unsafe.Pointer(&ctx)), (*C.struct_AVOutputFormat)(format), nil, cFilename)); err != 0 {
		return nil, err
	} else {
		ctx.SetPb(writer)
	}

	// TODO: Mark AVFMT_NOFILE

	// Return success
	return ctx, nil
}

// Open an output file.
func AVFormat_create_file(filename string, format *AVOutputFormat) (*AVFormatContext, error) {
	var ctx *AVFormatContext

	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))
	if err := AVError(C.avformat_alloc_output_context2((**C.struct_AVFormatContext)(unsafe.Pointer(&ctx)), (*C.struct_AVOutputFormat)(format), nil, cFilename)); err != 0 {
		return nil, err
	} else if !ctx.Output().Flags().Is(AVFMT_NOFILE) {
		if ioctx, err := AVFormat_avio_open(filename, AVIO_FLAG_WRITE); err != nil {
			return nil, err
		} else {
			ctx.SetPb(ioctx)
		}
	}

	// Return success
	return ctx, nil
}

func AVFormat_close_writer(ctx *AVFormatContext) error {
	var result error

	octx := (*C.struct_AVFormatContext)(ctx)
	if octx.oformat.flags&C.int(AVFMT_NOFILE) == 0 && octx.pb != nil {
		if err := AVError(C.avio_closep(&octx.pb)); err != 0 {
			result = errors.Join(result, err)
		}
	}
	C.avformat_free_context(octx)

	// Return any errors
	return result
}

// Allocate an AVFormatContext for an output format.
func AVFormat_alloc_output_context2(ctx **AVFormatContext, format *AVOutputFormat, filename string) error {
	var cFilename *C.char
	if filename != "" {
		cFilename = C.CString(filename)
	}
	defer C.free(unsafe.Pointer(cFilename))
	if err := AVError(C.avformat_alloc_output_context2((**C.struct_AVFormatContext)(unsafe.Pointer(&ctx)), (*C.struct_AVOutputFormat)(format), nil, cFilename)); err != 0 {
		return err
	}

	// Return success
	return nil
}

// Allocate the stream private data and initialize the codec, but do not write the header.
// May optionally be used before avformat_write_header() to initialize stream parameters before actually writing the header.
func AVFormat_init_output(ctx *AVFormatContext, options *AVDictionary) error {
	var opts **C.struct_AVDictionary
	if options != nil {
		opts = &options.ctx
	}
	if err := AVError(C.avformat_init_output((*C.struct_AVFormatContext)(ctx), opts)); err != 0 {
		return err
	} else {
		return nil
	}
}

// Allocate the stream private data and write the stream header to an output media file.
func AVFormat_write_header(ctx *AVFormatContext, options *AVDictionary) error {
	var opts **C.struct_AVDictionary
	if options != nil {
		opts = &options.ctx
	}
	if err := AVError(C.avformat_write_header((*C.struct_AVFormatContext)(ctx), opts)); err != 0 {
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
