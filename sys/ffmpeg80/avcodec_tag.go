package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavformat
#include <libavformat/avformat.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

// AVCodecTag is a single NULL-terminated codec id/tag lookup table, as returned
// by AVOutputFormat.CodecTags. The struct is only forward-declared in the public
// ffmpeg headers, so its fields cannot be read from Go directly; use ID and Tag
// instead, which delegate to ffmpeg's own lookup functions.
type AVCodecTag C.struct_AVCodecTag

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ID returns the codec ID mapped to by tag in this table, or AV_CODEC_ID_NONE
// if the tag is not present.
func (t *AVCodecTag) ID(tag uint32) AVCodecID {
	table := [2]*C.struct_AVCodecTag{(*C.struct_AVCodecTag)(t), nil}
	return AVCodecID(C.av_codec_get_id(&table[0], C.uint(tag)))
}

// Tag returns the container-specific tag mapped to by id in this table, or
// zero if the codec ID is not present.
func (t *AVCodecTag) Tag(id AVCodecID) uint32 {
	table := [2]*C.struct_AVCodecTag{(*C.struct_AVCodecTag)(t), nil}
	return uint32(C.av_codec_get_tag(&table[0], C.enum_AVCodecID(id)))
}
