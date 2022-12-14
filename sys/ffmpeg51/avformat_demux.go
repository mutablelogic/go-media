package ffmpeg

import (
	"io"
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
func AVFormat_open_input(url string, input_fmt *AVInputFormat, options **AVDictionary) (*AVFormatContext, error) {
	var ctx *AVFormatContext
	if err := AVError(C.avformat_open_input((**C.struct_AVFormatContext)(unsafe.Pointer(&ctx)), C.CString(url), (*C.struct_AVInputFormat)(input_fmt), (**C.struct_AVDictionary)(unsafe.Pointer(options)))); err != 0 {
		return nil, err
	} else {
		return ctx, nil
	}
}

// Close an opened input AVFormatContext. Free it and all its contents and set *s to NULL.
func AVFormat_close_input(ctx **AVFormatContext) {
	C.avformat_close_input((**C.struct_AVFormatContext)(unsafe.Pointer(ctx)))
}

// Close an opened input AVFormatContext.
func AVFormat_close_input_ptr(ctx *AVFormatContext) {
	C.avformat_close_input((**C.struct_AVFormatContext)(unsafe.Pointer(&ctx)))
}

// Read packets of a media file to get stream information.
func AVFormat_find_stream_info(ctx *AVFormatContext, options **AVDictionary) error {
	if err := AVError(C.avformat_find_stream_info((*C.struct_AVFormatContext)(ctx), (**C.struct_AVDictionary)(unsafe.Pointer(options)))); err < 0 {
		return err
	} else {
		return nil
	}
}

// Return the next frame of a stream. Returns io.EOF when no more frames are available.
func AVFormat_av_read_frame(ctx *AVFormatContext, pkt *AVPacket) error {
	if err := AVError(C.av_read_frame((*C.struct_AVFormatContext)(ctx), (*C.struct_AVPacket)(pkt))); err < 0 {
		if err == AVERROR_EOF {
			return io.EOF
		} else {
			return err
		}
	} else {
		return nil
	}
}

// Find the best stream based on a media type.
func AVFormat_av_find_best_stream(ctx *AVFormatContext, media_type AVMediaType, wanted_stream_nb int, related_stream int, decoder_ret **AVCodec, flags int) (int, error) {
	n := int(C.av_find_best_stream((*C.struct_AVFormatContext)(ctx), C.enum_AVMediaType(media_type), C.int(wanted_stream_nb), C.int(related_stream), (**C.struct_AVCodec)(unsafe.Pointer(decoder_ret)), C.int(flags)))
	if n < 0 {
		return 0, AVError(n)
	} else {
		return n, nil
	}
}
