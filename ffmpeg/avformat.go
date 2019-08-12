package ffmpeg

import (
	"unsafe"
	"sync"
	"fmt"
	"strconv"
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
	AVOutputFormat   C.struct_AVOutputFormat
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

var (
	once_init,once_deinit sync.Once
)

////////////////////////////////////////////////////////////////////////////////
// INIT AND DEINIT

// Register and Deregister
func AVFormatInit() {
	once_init.Do(func() {
		C.avformat_network_init()		
	})
}

func AVFormatDeinit() {
	once_deinit.Do(func() {
		C.avformat_network_deinit()
	})
}

////////////////////////////////////////////////////////////////////////////////
// AVFORMATCONTEXT

// NewAVFormatContext creates a new format context
func NewAVFormatContext() *AVFormatContext {
	return (*AVFormatContext)(C.avformat_alloc_context())
}

// Free AVFormatContext
func (this *AVFormatContext) Free() {
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

// Close Input
func (this *AVFormatContext) CloseInput() {
	ctx := (*C.AVFormatContext)(unsafe.Pointer(this))
	C.avformat_close_input(&ctx)
}

// Return Metadata Dictionary
func (this *AVFormatContext) Metadata() *AVDictionary {
	return &AVDictionary{ctx: this.metadata}
}

// Return Filename
func (this *AVFormatContext) Filename() string {
	return C.GoString(&this.filename[0])
}

////////////////////////////////////////////////////////////////////////////////
// AVInputFormat and AVOutputFormat

// Return input formats
func EnumerateInputFormats() []*AVInputFormat {
	a := make([]*AVInputFormat,0,100)
	p := unsafe.Pointer(uintptr(0))
	for {
		if iformat := (*AVInputFormat)(C.av_demuxer_iterate(&p)); iformat == nil {
			break
		} else {
			a = append(a,iformat)
		}
	}
	return a
}


// Return output formats
func EnumerateOutputFormats() []*AVOutputFormat {
	a := make([]*AVOutputFormat,0,100)
	p := unsafe.Pointer(uintptr(0))
	for {
		if oformat := (*AVOutputFormat)(C.av_muxer_iterate(&p)); oformat == nil {
			break
		} else {
			a = append(a,oformat)
		}
	}
	return a
}

func  (this *AVInputFormat) Name() string {
	return C.GoString(this.name)
}

func  (this *AVInputFormat) Description() string {
	return C.GoString(this.long_name)
}

func (this *AVInputFormat) Ext() string {
	return C.GoString(this.extensions)
}

func (this *AVInputFormat) MimeType() string {
	return C.GoString(this.mime_type)
}


func  (this *AVOutputFormat) Name() string {
	return C.GoString(this.name)
}

func  (this *AVOutputFormat) Description() string {
	return C.GoString(this.long_name)
}

func (this *AVOutputFormat) Ext() string {
	return C.GoString(this.extensions)
}

func (this *AVOutputFormat) MimeType() string {
	return C.GoString(this.mime_type)
}

func (this *AVInputFormat) Id() int {
	return int(this.raw_codec_id)
}

func (this *AVInputFormat) String() string {
	return fmt.Sprintf("<AVInputFormat>{ id=0x%08X name=%v description=%v ext=%v mime_type=%v }",this.Id(),strconv.Quote(this.Name()),strconv.Quote(this.Description()),strconv.Quote(this.Ext()),strconv.Quote(this.MimeType()))
}

func (this *AVOutputFormat) String() string {
	return fmt.Sprintf("<AVOutputFormat>{ name=%v description=%v ext=%v mime_type=%v }",strconv.Quote(this.Name()),strconv.Quote(this.Description()),strconv.Quote(this.Ext()),strconv.Quote(this.MimeType()))
}
