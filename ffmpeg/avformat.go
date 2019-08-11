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
// TYPES

type (
	AVFormatContext C.struct_AVFormatContext
	AVInputFormat   C.struct_AVInputFormat
	//AVIOContext     C.struct_AVIOContext
)

type (
	AVIOFlags int
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AVIO_FLAG_NONE       AVIOFlags = 0
	AVIO_FLAG_READ       AVIOFlags = 1
	AVIO_FLAG_WRITE      AVIOFlags = 2
	AVIO_FLAG_READ_WRITE AVIOFlags = (AVIO_FLAG_READ | AVIO_FLAG_WRITE)
)

////////////////////////////////////////////////////////////////////////////////
// LIBRARY FUNCTIONS

// NewAVFormatContext creates a new format context
func NewAVFormatContext() *AVFormatContext {
	return (*AVFormatContext)(C.avformat_alloc_context())
}

// Close for AVFormatContext
func (this *AVFormatContext) Close() {
	ctx := (*C.AVFormatContext)(unsafe.Pointer(this))
	C.avformat_free_context(ctx)
}

// Open Input
func (this *AVFormatContext) OpenInput(filename string, input_format *AVInputFormat) error {
	filename_ := C.CString(filename)
	defer C.free(unsafe.Pointer(filename_))
	ctx := (*C.AVFormatContext)(unsafe.Pointer(this))
	dict := new(AVDictionary)
	if err := AVError(C.avformat_open_input(
		&ctx,
		filename_,
		(*C.struct_AVInputFormat)(input_format),
		(**C.struct_AVDictionary)(unsafe.Pointer(dict)),
	)); err != 0 {
		return err
	} else {
		return nil
	}
}

// Return Metadata Dictionary
func (this *AVFormatContext) Metadata() *AVDictionary {
	return &AVDictionary{ctx: this.metadata}
}

// Return Filename
func (this *AVFormatContext) Filename() string {
	return C.GoString(&this.filename[0])
}

/*
func NewAVIOContext(url string, flags AVIOFlags) (*AVIOContext, error) {
	ctx := new(AVIOContext)
	url_ := C.CString(url)
	defer C.free(unsafe.Pointer(url_))
	if err := AVError(C.avio_open((**C.AVIOContext)(unsafe.Pointer(ctx)), url_, C.int(flags))); err != 0 {
		return nil, err
	} else {
		return ctx, nil
	}
}

func (this *AVIOContext) Close() error {
	ctx := (*C.AVIOContext)(unsafe.Pointer(this))
	if err := AVError(C.avio_close(ctx)); err != 0 {
		return err
	} else {
		return nil
	}
}
*/
