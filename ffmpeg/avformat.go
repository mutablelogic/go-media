package ffmpeg

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>

int iformat_read_header(AVFormatContext* ctx) {
	return ctx->iformat->read_header(ctx);
}

*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVFormatContext   C.struct_AVFormatContext
	AVInputFormat     C.struct_AVInputFormat
	AVOutputFormat    C.struct_AVOutputFormat
	AVStream          C.struct_AVStream
	AVCodecParameters C.struct_AVCodecParameters
	AVPacket          C.struct_AVPacket
)

type (
	AVIOFlags     int
	AVDisposition uint32
	AVMediaType   int
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	AVIO_FLAG_NONE       AVIOFlags = 0
	AVIO_FLAG_READ       AVIOFlags = 1
	AVIO_FLAG_WRITE      AVIOFlags = 2
	AVIO_FLAG_READ_WRITE AVIOFlags = (AVIO_FLAG_READ | AVIO_FLAG_WRITE)
)

const (
	AV_DISPOSITION_DEFAULT          AVDisposition = 0x00001
	AV_DISPOSITION_DUB              AVDisposition = 0x00002
	AV_DISPOSITION_ORIGINAL         AVDisposition = 0x00004
	AV_DISPOSITION_COMMENT          AVDisposition = 0x00008
	AV_DISPOSITION_LYRICS           AVDisposition = 0x00010
	AV_DISPOSITION_KARAOKE          AVDisposition = 0x00020
	AV_DISPOSITION_FORCED           AVDisposition = 0x00040
	AV_DISPOSITION_HEARING_IMPAIRED AVDisposition = 0x00080
	AV_DISPOSITION_VISUAL_IMPAIRED  AVDisposition = 0x00100
	AV_DISPOSITION_CLEAN_EFFECTS    AVDisposition = 0x00200
	AV_DISPOSITION_ATTACHED_PIC     AVDisposition = 0x00400
	AV_DISPOSITION_TIMED_THUMBNAILS AVDisposition = 0x00800
	AV_DISPOSITION_CAPTIONS         AVDisposition = 0x10000
	AV_DISPOSITION_DESCRIPTIONS     AVDisposition = 0x20000
	AV_DISPOSITION_METADATA         AVDisposition = 0x40000
	AV_DISPOSITION_NONE             AVDisposition = 0x00000
	AV_DISPOSITION_MIN                            = AV_DISPOSITION_DEFAULT
	AV_DISPOSITION_MAX                            = AV_DISPOSITION_METADATA
)

const (
	AV_MEDIA_TYPE_UNKNOWN    AVMediaType = -1
	AV_MEDIA_TYPE_VIDEO      AVMediaType = 0
	AV_MEDIA_TYPE_AUDIO      AVMediaType = 1
	AV_MEDIA_TYPE_DATA       AVMediaType = 2
	AV_MEDIA_TYPE_SUBTITLE   AVMediaType = 3
	AV_MEDIA_TYPE_ATTACHMENT AVMediaType = 4
)

var (
	once_init, once_deinit sync.Once
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

// Read the format headers
func (this *AVFormatContext) ReadHeader() error {
	ctx := (*C.AVFormatContext)(unsafe.Pointer(this))
	if C.iformat_read_header(ctx) < 0 {
		return errors.New("read_header failed")
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

// Return number of streams
func (this *AVFormatContext) NumStreams() uint {
	ctx := (*C.AVFormatContext)(unsafe.Pointer(this))
	return uint(ctx.nb_streams)
}

// Return Streams
func (this *AVFormatContext) Streams() []*AVStream {
	var streams []*AVStream

	// Get context
	ctx := (*C.AVFormatContext)(unsafe.Pointer(this))

	// Make a fake slice
	if nb_streams := this.NumStreams(); nb_streams > 0 {
		// Make a fake slice
		sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&streams)))
		sliceHeader.Cap = int(nb_streams)
		sliceHeader.Len = int(nb_streams)
		sliceHeader.Data = uintptr(unsafe.Pointer(ctx.streams))
	}
	return streams
}

////////////////////////////////////////////////////////////////////////////////
// AVInputFormat and AVOutputFormat

// Return input formats
func EnumerateInputFormats() []*AVInputFormat {
	a := make([]*AVInputFormat, 0, 100)
	p := unsafe.Pointer(uintptr(0))
	for {
		if iformat := (*AVInputFormat)(C.av_demuxer_iterate(&p)); iformat == nil {
			break
		} else {
			a = append(a, iformat)
		}
	}
	return a
}

// Return output formats
func EnumerateOutputFormats() []*AVOutputFormat {
	a := make([]*AVOutputFormat, 0, 100)
	p := unsafe.Pointer(uintptr(0))
	for {
		if oformat := (*AVOutputFormat)(C.av_muxer_iterate(&p)); oformat == nil {
			break
		} else {
			a = append(a, oformat)
		}
	}
	return a
}

func (this *AVInputFormat) Name() string {
	return C.GoString(this.name)
}

func (this *AVInputFormat) Description() string {
	return C.GoString(this.long_name)
}

func (this *AVInputFormat) Ext() string {
	return C.GoString(this.extensions)
}

func (this *AVInputFormat) MimeType() string {
	return C.GoString(this.mime_type)
}

func (this *AVOutputFormat) Name() string {
	return C.GoString(this.name)
}

func (this *AVOutputFormat) Description() string {
	return C.GoString(this.long_name)
}

func (this *AVOutputFormat) Ext() string {
	return C.GoString(this.extensions)
}

func (this *AVOutputFormat) MimeType() string {
	return C.GoString(this.mime_type)
}

func (this *AVInputFormat) String() string {
	return fmt.Sprintf("<AVInputFormat>{ name=%v description=%v ext=%v mime_type=%v }", strconv.Quote(this.Name()), strconv.Quote(this.Description()), strconv.Quote(this.Ext()), strconv.Quote(this.MimeType()))
}

func (this *AVOutputFormat) String() string {
	return fmt.Sprintf("<AVOutputFormat>{ name=%v description=%v ext=%v mime_type=%v }", strconv.Quote(this.Name()), strconv.Quote(this.Description()), strconv.Quote(this.Ext()), strconv.Quote(this.MimeType()))
}

////////////////////////////////////////////////////////////////////////////////
// AVCodecParameters

func (this *AVCodecParameters) Type() AVMediaType {
	ctx := (*C.AVCodecParameters)(unsafe.Pointer(this))
	return AVMediaType(ctx.codec_type)
}

func (this *AVCodecParameters) BitRate() int64 {
	ctx := (*C.AVCodecParameters)(unsafe.Pointer(this))
	return int64(ctx.bit_rate)
}

func (this *AVCodecParameters) Width() int {
	ctx := (*C.AVCodecParameters)(unsafe.Pointer(this))
	return int(ctx.width)
}

func (this *AVCodecParameters) Height() int {
	ctx := (*C.AVCodecParameters)(unsafe.Pointer(this))
	return int(ctx.height)
}

func (this *AVCodecParameters) String() string {
	if this.Type() == AV_MEDIA_TYPE_VIDEO {
		return fmt.Sprintf("<AVCodecParameters>{ type=%v bit_rate=%v width=%v height=%v }", this.Type(), this.BitRate(), this.Width(), this.Height())
	} else {
		return fmt.Sprintf("<AVCodecParameters>{ type=%v bit_rate=%v }", this.Type(), this.BitRate())
	}
}

////////////////////////////////////////////////////////////////////////////////
// AVStream

func (this *AVStream) Index() int {
	ctx := (*C.AVStream)(unsafe.Pointer(this))
	return int(ctx.index)
}

func (this *AVStream) Id() int {
	ctx := (*C.AVStream)(unsafe.Pointer(this))
	return int(ctx.id)
}

func (this *AVStream) Metadata() *AVDictionary {
	return &AVDictionary{ctx: this.metadata}
}

func (this *AVStream) Duration() int64 {
	ctx := (*C.AVStream)(unsafe.Pointer(this))
	return int64(ctx.duration)
}

func (this *AVStream) NumFrames() int64 {
	ctx := (*C.AVStream)(unsafe.Pointer(this))
	return int64(ctx.nb_frames)
}

func (this *AVStream) StartTime() int64 {
	ctx := (*C.AVStream)(unsafe.Pointer(this))
	return int64(ctx.start_time)
}

func (this *AVStream) TimeBase() AVRational {
	ctx := (*C.AVStream)(unsafe.Pointer(this))
	return AVRational(ctx.time_base)
}

func (this *AVStream) MeanFrameRate() AVRational {
	ctx := (*C.AVStream)(unsafe.Pointer(this))
	return AVRational(ctx.avg_frame_rate)
}

func (this *AVStream) Disposition() AVDisposition {
	ctx := (*C.AVStream)(unsafe.Pointer(this))
	return AVDisposition(ctx.disposition)
}

func (this *AVStream) Codec() *AVCodecParameters {
	ctx := (*C.AVStream)(unsafe.Pointer(this))
	return (*AVCodecParameters)(ctx.codecpar)
}

func (this *AVStream) AttachedPic() AVPacket {
	ctx := (*C.AVStream)(unsafe.Pointer(this))
	return (AVPacket)(ctx.attached_pic)
}

func (this *AVStream) String() string {
	return fmt.Sprintf("<AVStream>{ index=%v id=%v disposition=%v codec=%v metadata=%v time_base=%v duration=%v mean_frame_rate=%v  }",
		this.Index(), this.Id(),
		this.Disposition(), this.Codec(), this.Metadata(), this.TimeBase(), this.Duration(), this.MeanFrameRate())
}

////////////////////////////////////////////////////////////////////////////////
// AVMediaType

func (t AVMediaType) String() string {
	switch t {
	case AV_MEDIA_TYPE_UNKNOWN:
		return "AV_MEDIA_TYPE_UNKNOWN"
	case AV_MEDIA_TYPE_VIDEO:
		return "AV_MEDIA_TYPE_VIDEO"
	case AV_MEDIA_TYPE_AUDIO:
		return "AV_MEDIA_TYPE_AUDIO"
	case AV_MEDIA_TYPE_DATA:
		return "AV_MEDIA_TYPE_DATA"
	case AV_MEDIA_TYPE_SUBTITLE:
		return "AV_MEDIA_TYPE_SUBTITLE"
	case AV_MEDIA_TYPE_ATTACHMENT:
		return "AV_MEDIA_TYPE_ATTACHMENT"
	default:
		return "[?? Invalid AVMediaType value]"
	}
}

////////////////////////////////////////////////////////////////////////////////
// AVDisposition

func (t AVDisposition) String() string {
	if t == AV_DISPOSITION_NONE {
		return "AV_DISPOSITION_NONE"
	}
	parts := ""
	for f := AV_DISPOSITION_MIN; f <= AV_DISPOSITION_MAX; f <<= 1 {
		if t&f == 0 {
			continue
		}
		switch f {
		case AV_DISPOSITION_DEFAULT:
			parts += "|" + "AV_DISPOSITION_DEFAULT"
		case AV_DISPOSITION_DUB:
			parts += "|" + "AV_DISPOSITION_DUB"
		case AV_DISPOSITION_ORIGINAL:
			parts += "|" + "AV_DISPOSITION_ORIGINAL"
		case AV_DISPOSITION_COMMENT:
			parts += "|" + "AV_DISPOSITION_COMMENT"
		case AV_DISPOSITION_LYRICS:
			parts += "|" + "AV_DISPOSITION_LYRICS"
		case AV_DISPOSITION_KARAOKE:
			parts += "|" + "AV_DISPOSITION_KARAOKE"
		case AV_DISPOSITION_FORCED:
			parts += "|" + "AV_DISPOSITION_FORCED"
		case AV_DISPOSITION_HEARING_IMPAIRED:
			parts += "|" + "AV_DISPOSITION_HEARING_IMPAIRED"
		case AV_DISPOSITION_VISUAL_IMPAIRED:
			parts += "|" + "AV_DISPOSITION_VISUAL_IMPAIRED"
		case AV_DISPOSITION_CLEAN_EFFECTS:
			parts += "|" + "AV_DISPOSITION_CLEAN_EFFECTS"
		case AV_DISPOSITION_ATTACHED_PIC:
			parts += "|" + "AV_DISPOSITION_ATTACHED_PIC"
		case AV_DISPOSITION_TIMED_THUMBNAILS:
			parts += "|" + "AV_DISPOSITION_TIMED_THUMBNAILS"
		case AV_DISPOSITION_CAPTIONS:
			parts += "|" + "AV_DISPOSITION_CAPTIONS"
		case AV_DISPOSITION_DESCRIPTIONS:
			parts += "|" + "AV_DISPOSITION_DESCRIPTIONS"
		case AV_DISPOSITION_METADATA:
			parts += "|" + "AV_DISPOSITION_METADATA"
		default:
			parts += "|" + "[?? Invalid AVDisposition value"
		}
	}
	return strings.TrimPrefix(parts, "|")
}

////////////////////////////////////////////////////////////////////////////////
// AVPacket

func (this AVPacket) Data() []byte {
	var data []byte
	// Make a fake slice
	ctx := (C.AVPacket)(this)
	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&data)))
	sliceHeader.Cap = int(ctx.size)
	sliceHeader.Len = int(ctx.size)
	sliceHeader.Data = uintptr(unsafe.Pointer(ctx.data))
	return data
}

func (this AVPacket) Size() int {
	ctx := (C.AVPacket)(this)
	return int(ctx.size)
}

func (this AVPacket) String() string {
	return fmt.Sprintf("<AVPacket>{ size=%v }", this.Size())
}
