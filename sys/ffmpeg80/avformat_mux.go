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

// Open an output stream with custom I/O (using AVIOContextEx).
// Useful for writing to memory, network, or custom destinations.
// The format parameter can be nil to auto-detect from the filename.
func AVFormat_open_writer(writer *AVIOContextEx, format *AVOutputFormat, filename string) (*AVFormatContext, error) {
	var ctx *AVFormatContext

	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	// Allocate output context
	if err := AVError(C.avformat_alloc_output_context2((**C.struct_AVFormatContext)(unsafe.Pointer(&ctx)), (*C.struct_AVOutputFormat)(format), nil, cFilename)); err != 0 {
		return nil, err
	}

	// Set custom I/O
	ctx.SetPb(writer)
	ctx.SetFlags(ctx.Flags() | AVFMT_FLAG_CUSTOM_IO)

	return ctx, nil
}

// Create an output file with automatic I/O management.
// Opens the file for writing and sets up the AVIOContext.
// The format parameter can be nil to auto-detect from the filename.
func AVFormat_create_file(filename string, format *AVOutputFormat) (*AVFormatContext, error) {
	var ctx *AVFormatContext

	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	// Allocate output context
	if err := AVError(C.avformat_alloc_output_context2((**C.struct_AVFormatContext)(unsafe.Pointer(&ctx)), (*C.struct_AVOutputFormat)(format), nil, cFilename)); err != 0 {
		return nil, err
	}

	// Open I/O if the format needs a file
	if !ctx.Output().Flags().Is(AVFMT_NOFILE) {
		if ioctx, err := AVFormat_avio_open(filename, AVIO_FLAG_WRITE); err != nil {
			AVFormat_free_context(ctx)
			return nil, err
		} else {
			ctx.SetPb(ioctx)
		}
	}

	return ctx, nil
}

// Close an output writer and free resources.
// Closes the I/O context if it was opened by AVFormat_create_file.
// Does not close custom I/O contexts created by AVFormat_open_writer.
func AVFormat_close_writer(ctx *AVFormatContext) error {
	if ctx == nil {
		return nil
	}

	var result error
	octx := (*C.struct_AVFormatContext)(ctx)

	// Close I/O if it's not a custom I/O and not NOFILE
	if octx.oformat.flags&C.int(AVFMT_NOFILE) == 0 && octx.flags&C.int(AVFMT_FLAG_CUSTOM_IO) == 0 {
		if octx.pb != nil {
			if err := AVError(C.avio_closep(&octx.pb)); err != 0 {
				result = errors.Join(result, err)
			}
		}
	}

	// Free the context
	C.avformat_free_context(octx)

	return result
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
	}
	return nil
}

// Write the stream trailer to an output media file and free the file private data.
func AVFormat_write_trailer(ctx *AVFormatContext) error {
	if err := AVError(C.av_write_trailer((*C.struct_AVFormatContext)(ctx))); err != 0 {
		return err
	}
	return nil
}
