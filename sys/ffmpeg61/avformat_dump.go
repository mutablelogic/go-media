package ffmpeg

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func AVFormat_dump_format(ctx *AVFormatContext, stream_index int, filename string) {
	ctx_ := (*C.struct_AVFormatContext)(ctx)
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))
	C.av_dump_format((*C.AVFormatContext)(ctx), C.int(stream_index), cFilename, boolToInt(ctx_.oformat != nil))
}
