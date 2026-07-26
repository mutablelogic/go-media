package ffmpeg

import (
	"encoding/json"
	"errors"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
#include <libavutil/mem.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVChapter C.struct_AVChapter
)

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (ctx *AVChapter) Id() int64 {
	return int64(ctx.id)
}

func (ctx *AVChapter) TimeBase() AVRational {
	return AVRational(ctx.time_base)
}

func (ctx *AVChapter) Start() int64 {
	return int64(ctx.start)
}

func (ctx *AVChapter) End() int64 {
	return int64(ctx.end)
}

func (ctx *AVChapter) Metadata() *AVDictionary {
	return &AVDictionary{ctx.metadata}
}

func (ctx *AVChapter) SetId(id int64) {
	ctx.id = C.int64_t(id)
}

func (ctx *AVChapter) SetTimeBase(tb AVRational) {
	ctx.time_base = C.AVRational(tb)
}

func (ctx *AVChapter) SetStart(start int64) {
	ctx.start = C.int64_t(start)
}

func (ctx *AVChapter) SetEnd(end int64) {
	ctx.end = C.int64_t(end)
}

// SetMetadata takes ownership of dict, the same way AVFormatContext's and
// AVStream's SetMetadata do - the caller must not free it afterwards.
func (ctx *AVChapter) SetMetadata(dict *AVDictionary) {
	if dict == nil {
		ctx.metadata = nil
	} else {
		ctx.metadata = dict.ctx
	}
}

////////////////////////////////////////////////////////////////////////////////
// MUXING

// AVFormat_new_chapters allocates ctx.chapters/nb_chapters to hold n
// chapters and returns them for the caller to populate via AVChapter's
// setters. Per avformat.h's own doc comment on AVFormatContext.chapters,
// chapters are normally "set by user" when muxing (libavformat has no public
// constructor for them, unlike AVFormat_new_stream for streams) - so this
// allocates the pointer array and each chapter by hand, using FFmpeg's own
// allocator (av_malloc/av_mallocz) so avformat_free_context can free them
// exactly as it frees demuxed chapters. Must be called before
// AVFormat_write_header for muxers that write chapters in the header (most
// do - mov/mp4, mkv, ...).
func AVFormat_new_chapters(ctx *AVFormatContext, n int) ([]*AVChapter, error) {
	if n <= 0 {
		return nil, nil
	}

	arr := C.av_malloc(C.size_t(n) * C.size_t(unsafe.Sizeof(uintptr(0))))
	if arr == nil {
		return nil, errors.New("failed to allocate chapters array")
	}

	result := make([]*AVChapter, n)
	ptr := uintptr(arr)
	for i := 0; i < n; i++ {
		chapter := (*C.struct_AVChapter)(C.av_mallocz(C.size_t(unsafe.Sizeof(C.struct_AVChapter{}))))
		if chapter == nil {
			for j := 0; j < i; j++ {
				C.av_free(unsafe.Pointer((*C.struct_AVChapter)(result[j])))
			}
			C.av_free(arr)
			return nil, errors.New("failed to allocate chapter")
		}
		*(**C.struct_AVChapter)(unsafe.Pointer(ptr)) = chapter
		result[i] = (*AVChapter)(chapter)
		ptr += unsafe.Sizeof(uintptr(0))
	}

	ctx.chapters = (**C.struct_AVChapter)(arr)
	ctx.nb_chapters = C.uint(n)
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVChapter) MarshalJSON() ([]byte, error) {
	type jsonAVChapter struct {
		Id       int64      `json:"id"`
		TimeBase AVRational `json:"time_base,omitempty"`
		Start    int64      `json:"start"`
		End      int64      `json:"end"`
	}
	return json.Marshal(jsonAVChapter{
		Id:       ctx.Id(),
		TimeBase: ctx.TimeBase(),
		Start:    ctx.Start(),
		End:      ctx.End(),
	})
}

func (ctx *AVChapter) String() string {
	return marshalToString(ctx)
}
