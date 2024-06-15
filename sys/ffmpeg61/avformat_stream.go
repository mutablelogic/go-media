package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func AVFormat_new_stream(ctx *AVFormatContext, c *AVCodec) *AVStream {
	return (*AVStream)(C.avformat_new_stream((*C.struct_AVFormatContext)(ctx), (*C.struct_AVCodec)(c)))
}

// Find the best stream given the media type, wanted stream number, and related stream number.
func AVFormat_find_best_stream(ctx *AVFormatContext, t AVMediaType, wanted int, related int) (int, *AVCodec, error) {
	var codec *C.struct_AVCodec
	ret := int(C.av_find_best_stream((*C.struct_AVFormatContext)(ctx), (C.enum_AVMediaType)(t), C.int(wanted), C.int(related), (**C.struct_AVCodec)(&codec), 0))
	if ret < 0 {
		return 0, nil, AVError(ret)
	} else {
		return ret, (*AVCodec)(codec), nil
	}
}
