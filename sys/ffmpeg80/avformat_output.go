package ffmpeg

import (
	"encoding/json"
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
	AVOutputFormat C.struct_AVOutputFormat
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

// Standards-compliance levels, for use with AVFormat_query_codec.
const (
	FF_COMPLIANCE_VERY_STRICT  = C.FF_COMPLIANCE_VERY_STRICT  // Strictly conform to an older, more strict version of the spec or reference software.
	FF_COMPLIANCE_STRICT       = C.FF_COMPLIANCE_STRICT       // Strictly conform to all the things in the spec no matter what consequences.
	FF_COMPLIANCE_NORMAL       = C.FF_COMPLIANCE_NORMAL       // Default
	FF_COMPLIANCE_UNOFFICIAL   = C.FF_COMPLIANCE_UNOFFICIAL   // Allow unofficial extensions
	FF_COMPLIANCE_EXPERIMENTAL = C.FF_COMPLIANCE_EXPERIMENTAL // Allow nonstandardized experimental things.
)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx *AVOutputFormat) MarshalJSON() ([]byte, error) {
	type jsonAVOutputFormat struct {
		Name          string    `json:"name,omitempty"`
		LongName      string    `json:"long_name,omitempty"`
		MimeTypes     string    `json:"mime_types,omitempty"`
		Flags         AVFormat  `json:"flags,omitempty"`
		Extensions    string    `json:"extensions,omitempty"`
		VideoCodec    AVCodecID `json:"video_codec,omitempty"`
		AudioCodec    AVCodecID `json:"audio_codec,omitempty"`
		SubtitleCodec AVCodecID `json:"subtitle_codec,omitempty"`
	}
	return json.Marshal(jsonAVOutputFormat{
		Name:          C.GoString(ctx.name),
		LongName:      C.GoString(ctx.long_name),
		MimeTypes:     C.GoString(ctx.mime_type),
		Flags:         AVFormat(ctx.flags),
		Extensions:    C.GoString(ctx.extensions),
		VideoCodec:    AVCodecID(ctx.video_codec),
		AudioCodec:    AVCodecID(ctx.audio_codec),
		SubtitleCodec: AVCodecID(ctx.subtitle_codec),
	})
}

func (ctx *AVOutputFormat) String() string {
	return marshalToString(ctx)
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (ctx *AVOutputFormat) Name() string {
	return C.GoString(ctx.name)
}

func (ctx *AVOutputFormat) LongName() string {
	return C.GoString(ctx.long_name)
}

func (ctx *AVOutputFormat) Flags() AVFormat {
	return AVFormat(ctx.flags)
}

func (ctx *AVOutputFormat) MimeTypes() string {
	return C.GoString(ctx.mime_type)
}

func (ctx *AVOutputFormat) Extensions() string {
	return C.GoString(ctx.extensions)
}

func (ctx *AVOutputFormat) VideoCodec() AVCodecID {
	return AVCodecID(ctx.video_codec)
}

func (ctx *AVOutputFormat) AudioCodec() AVCodecID {
	return AVCodecID(ctx.audio_codec)
}

func (ctx *AVOutputFormat) SubtitleCodec() AVCodecID {
	return AVCodecID(ctx.subtitle_codec)
}

func (ctx *AVOutputFormat) PrivClass() *AVClass {
	return (*AVClass)(ctx.priv_class)
}

// CodecTags returns the codec tag tables associated with this output format.
// Each returned table is opaque; use its ID and Tag methods to query it.
func (ctx *AVOutputFormat) CodecTags() []*AVCodecTag {
	var tables []*AVCodecTag
	ptr := uintptr(unsafe.Pointer(ctx.codec_tag))
	if ptr == 0 {
		return nil
	}
	for {
		table := *(**C.struct_AVCodecTag)(unsafe.Pointer(ptr))
		if table == nil {
			break
		}
		tables = append(tables, (*AVCodecTag)(table))
		ptr += unsafe.Sizeof(uintptr(0))
	}
	return tables
}

////////////////////////////////////////////////////////////////////////////////
// BINDINGS

// AVFormat_query_codec reports whether a codec is supported by an output format.
// It returns 1 if the codec can be stored in this format, 0 if it cannot, and a
// negative number if this cannot be determined.
func AVFormat_query_codec(ofmt *AVOutputFormat, id AVCodecID, stdCompliance int) int {
	return int(C.avformat_query_codec((*C.AVOutputFormat)(ofmt), C.enum_AVCodecID(id), C.int(stdCompliance)))
}

// Iterate over all AVOutputFormats
func AVFormat_muxer_iterate(opaque *uintptr) *AVOutputFormat {
	return (*AVOutputFormat)(C.av_muxer_iterate((*unsafe.Pointer)(unsafe.Pointer(opaque))))
}

// Return the output format in the list of registered output formats which best matches the provided parameters, or return NULL if there is no match.
func AVFormat_guess_format(format, filename, mimetype string) *AVOutputFormat {
	var cFilename, cFormat, cMimeType *C.char
	if format != "" {
		cFormat = C.CString(format)
		defer C.free(unsafe.Pointer(cFormat))
	}
	if filename != "" {
		cFilename = C.CString(filename)
		defer C.free(unsafe.Pointer(cFilename))
	}
	if mimetype != "" {
		cMimeType = C.CString(mimetype)
		defer C.free(unsafe.Pointer(cMimeType))
	}
	return (*AVOutputFormat)(C.av_guess_format(cFormat, cFilename, cMimeType))
}

// Write a packet to an output media file ensuring correct interleaving.
// This function will buffer the packets internally as needed to make sure the
// packets in the output file are properly interleaved, usually ordered by
// increasing dts. Callers doing their own interleaving should call
// AVFormat_write_frame() instead of this function.
func AVFormat_interleaved_write_frame(ctx *AVFormatContext, pkt *AVPacket) error {
	if err := AVError(C.av_interleaved_write_frame((*C.AVFormatContext)(ctx), (*C.AVPacket)(pkt))); err != 0 {
		return err
	}
	return nil
}

// Write a packet to an output media file without interleaving.
// The caller is responsible for correctly interleaving the packets if the
// codec requires it. Most callers should use AVFormat_interleaved_write_frame()
// instead for proper interleaving.
func AVFormat_write_frame(ctx *AVFormatContext, pkt *AVPacket) error {
	if err := AVError(C.av_write_frame((*C.AVFormatContext)(ctx), (*C.AVPacket)(pkt))); err != 0 {
		return err
	}
	return nil
}
