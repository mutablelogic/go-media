package ffmpeg

import "unsafe"

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVCodecParser        C.struct_AVCodecParser
	AVCodecParserContext C.struct_AVCodecParserContext
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// Iterate over all registered codec parsers.
func AVCodec_parser_iterate(opaque *uintptr) *AVCodecParser {
	return (*AVCodecParser)(C.av_parser_iterate((*unsafe.Pointer)(unsafe.Pointer(opaque))))
}

// Initialize a parser context for the specified codec ID.
func AVCodec_parser_init(codec_id AVCodecID) *AVCodecParserContext {
	return (*AVCodecParserContext)(C.av_parser_init(C.int(codec_id)))
}

// Close and free a parser context.
func AVCodec_parser_close(parser *AVCodecParserContext) {
	C.av_parser_close((*C.struct_AVCodecParserContext)(parser))
}

// Parse a packet from input buffer.
func AVCodec_parser_parse(parser *AVCodecParserContext, ctx *AVCodecContext, packet *AVPacket, buf []byte, pts int64, dts int64, pos int64) int {
	return int(C.av_parser_parse2((*C.struct_AVCodecParserContext)(parser), (*C.struct_AVCodecContext)(ctx), &packet.data, &packet.size, (*C.uint8_t)(unsafe.Pointer(&buf[0])), C.int(len(buf)), C.int64_t(pts), C.int64_t(dts), C.int64_t(pos)))
}
