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
// TYPES

type (
	AVFormatContext C.struct_AVFormatContext
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Allocate an AVFormatContext.
func AVFormat_alloc_context() *AVFormatContext {
	return (*AVFormatContext)(C.avformat_alloc_context())
}

// Free an AVFormatContext and all its streams.
func AVFormat_free_context(ctx *AVFormatContext) {
	C.avformat_free_context((*C.struct_AVFormatContext)(ctx))
}

func (ctx *AVFormatContext) Input() *AVInputFormat {
	return (*AVInputFormat)(ctx.iformat)
}

func (ctx *AVFormatContext) Output() *AVOutputFormat {
	return (*AVOutputFormat)(ctx.oformat)
}

func (ctx *AVFormatContext) Metadata() *AVDictionary {
	return &AVDictionary{ctx.metadata}
}

func (ctx *AVFormatContext) SetMetadata(dict *AVDictionary) {
	if dict == nil {
		ctx.metadata = nil
	} else {
		ctx.metadata = dict.ctx
	}
}

func (ctx *AVFormatContext) SetPb(pb *AVIOContextEx) {
	if pb == nil {
		ctx.pb = nil
	} else {
		ctx.pb = (*C.struct_AVIOContext)(pb.AVIOContext)
	}
}

func (ctx *AVFormatContext) NumStreams() uint {
	return uint(ctx.nb_streams)
}

func (ctx *AVFormatContext) Streams() []*AVStream {
	return cAVStreamSlice(unsafe.Pointer(ctx.streams), C.int(ctx.nb_streams))
}

func (ctx *AVFormatContext) Stream(stream int) *AVStream {
	streams := ctx.Streams()
	if stream < 0 || stream >= len(streams) {
		return nil
	} else {
		return streams[stream]
	}
}

func (ctx *AVFormatContext) Flags() AVFormatFlag {
	return AVFormatFlag(ctx.flags)
}

func (ctx *AVFormatContext) SetFlags(flag AVFormatFlag) {
	ctx.flags = C.int(flag)
}

func (ctx *AVFormatContext) Duration() int64 {
	return int64(ctx.duration)
}

func (ctx *AVFormatContext) StartTime() int64 {
	return int64(ctx.start_time)
}

func (ctx *AVFormatContext) BitRate() int64 {
	return int64(ctx.bit_rate)
}

func (ctx *AVFormatContext) Filename() string {
	return C.GoString(ctx.url)
}

func (ctx *AVFormatContext) ProbeSize() int64 {
	return int64(ctx.probesize)
}

func (ctx *AVFormatContext) SetProbeSize(size int64) {
	ctx.probesize = C.int64_t(size)
}

func (ctx *AVFormatContext) MaxAnalyzeDuration() int64 {
	return int64(ctx.max_analyze_duration)
}

func (ctx *AVFormatContext) SetMaxAnalyzeDuration(duration int64) {
	ctx.max_analyze_duration = C.int64_t(duration)
}

func (ctx *AVFormatContext) NumChapters() uint {
	return uint(ctx.nb_chapters)
}

func (ctx *AVFormatContext) NumPrograms() uint {
	return uint(ctx.nb_programs)
}

func (ctx *AVFormatContext) ContextFlags() int {
	return int(ctx.ctx_flags)
}
