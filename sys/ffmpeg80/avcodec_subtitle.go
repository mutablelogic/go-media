package ffmpeg

import (
	"encoding/json"
	"syscall"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec libavutil
#include <libavcodec/avcodec.h>
*/
import "C"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVSubtitle     C.struct_AVSubtitle
	AVSubtitleRect C.struct_AVSubtitleRect
	AVSubtitleType C.enum_AVSubtitleType
)

////////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	SUBTITLE_NONE   AVSubtitleType = C.SUBTITLE_NONE
	SUBTITLE_BITMAP AVSubtitleType = C.SUBTITLE_BITMAP
	SUBTITLE_TEXT   AVSubtitleType = C.SUBTITLE_TEXT
	SUBTITLE_ASS    AVSubtitleType = C.SUBTITLE_ASS
)

////////////////////////////////////////////////////////////////////////////////
// SUBTITLE MEMORY MANAGEMENT

// Free all allocated data in the given subtitle struct.
func AVSubtitle_free(sub *AVSubtitle) {
	C.avsubtitle_free((*C.struct_AVSubtitle)(sub))
}

// Initialize subtitle header for ASS-based subtitle formats (subrip, ass, ssa, etc.).
// This should be called before avcodec_open2() for subtitle decoders that need header initialization.
// Returns error if memory allocation fails.
func AVCodec_init_subtitle_header(ctx *AVCodecContext) error {
	// Default ASS subtitle header similar to ff_ass_subtitle_header_default
	header := "[Script Info]\n" +
		"ScriptType: v4.00+\n" +
		"PlayResX: 384\n" +
		"PlayResY: 288\n" +
		"ScaledBorderAndShadow: yes\n\n" +
		"[V4+ Styles]\n" +
		"Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n" +
		"Style: Default,Arial,16,&Hffffff,&Hffffff,&H0,&H0,0,0,0,0,100,100,0,0,1,1,0,2,10,10,10,1\n\n" +
		"[Events]\n" +
		"Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n"

	cHeader := C.CString(header)
	defer C.free(unsafe.Pointer(cHeader))

	// Allocate and copy the header
	headerSize := C.size_t(len(header))
	cCtx := (*C.struct_AVCodecContext)(ctx)
	cCtx.subtitle_header = (*C.uint8_t)(C.av_malloc(headerSize + 1))
	if cCtx.subtitle_header == nil {
		return AVError(syscall.ENOMEM)
	}
	C.memcpy(unsafe.Pointer(cCtx.subtitle_header), unsafe.Pointer(cHeader), headerSize)
	*(*C.uint8_t)(unsafe.Pointer(uintptr(unsafe.Pointer(cCtx.subtitle_header)) + uintptr(headerSize))) = 0
	cCtx.subtitle_header_size = C.int(headerSize)

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// AVSubtitle PROPERTIES

func (sub *AVSubtitle) Format() uint16 {
	return uint16(sub.format)
}

func (sub *AVSubtitle) SetFormat(format uint16) {
	sub.format = C.uint16_t(format)
}

func (sub *AVSubtitle) StartDisplayTime() uint32 {
	return uint32(sub.start_display_time)
}

func (sub *AVSubtitle) SetStartDisplayTime(ms uint32) {
	sub.start_display_time = C.uint32_t(ms)
}

func (sub *AVSubtitle) EndDisplayTime() uint32 {
	return uint32(sub.end_display_time)
}

func (sub *AVSubtitle) SetEndDisplayTime(ms uint32) {
	sub.end_display_time = C.uint32_t(ms)
}

func (sub *AVSubtitle) NumRects() uint {
	return uint(sub.num_rects)
}

func (sub *AVSubtitle) Rects() []*AVSubtitleRect {
	if sub.num_rects == 0 || sub.rects == nil {
		return nil
	}
	// Create a slice from the C array of pointers
	rects := make([]*AVSubtitleRect, sub.num_rects)
	ptr := uintptr(unsafe.Pointer(sub.rects))
	for i := range rects {
		rectPtr := *(**C.struct_AVSubtitleRect)(unsafe.Pointer(ptr))
		rects[i] = (*AVSubtitleRect)(rectPtr)
		ptr += unsafe.Sizeof(uintptr(0))
	}
	return rects
}

func (sub *AVSubtitle) PTS() int64 {
	return int64(sub.pts)
}

func (sub *AVSubtitle) SetPTS(pts int64) {
	sub.pts = C.int64_t(pts)
}

////////////////////////////////////////////////////////////////////////////////
// AVSubtitleRect PROPERTIES

func (rect *AVSubtitleRect) Type() AVSubtitleType {
	return AVSubtitleType(rect._type)
}

func (rect *AVSubtitleRect) SetType(t AVSubtitleType) {
	rect._type = C.enum_AVSubtitleType(t)
}

func (rect *AVSubtitleRect) X() int {
	return int(rect.x)
}

func (rect *AVSubtitleRect) SetX(x int) {
	rect.x = C.int(x)
}

func (rect *AVSubtitleRect) Y() int {
	return int(rect.y)
}

func (rect *AVSubtitleRect) SetY(y int) {
	rect.y = C.int(y)
}

func (rect *AVSubtitleRect) Width() int {
	return int(rect.w)
}

func (rect *AVSubtitleRect) SetWidth(w int) {
	rect.w = C.int(w)
}

func (rect *AVSubtitleRect) Height() int {
	return int(rect.h)
}

func (rect *AVSubtitleRect) SetHeight(h int) {
	rect.h = C.int(h)
}

func (rect *AVSubtitleRect) NumColors() int {
	return int(rect.nb_colors)
}

func (rect *AVSubtitleRect) SetNumColors(n int) {
	rect.nb_colors = C.int(n)
}

// Get bitmap data for SUBTITLE_BITMAP type
func (rect *AVSubtitleRect) Data(plane int) []byte {
	if plane < 0 || plane >= 4 || rect.data[plane] == nil {
		return nil
	}
	linesize := int(rect.linesize[plane])
	height := int(rect.h)
	if linesize <= 0 || height <= 0 {
		return nil
	}
	// Return a slice view of the C memory (be careful with lifetime)
	size := linesize * height
	// Sanity check: prevent unreasonably large allocations (> 100MB)
	if size < 0 || size > 100*1024*1024 {
		return nil
	}
	return unsafe.Slice((*byte)(rect.data[plane]), size)
}

func (rect *AVSubtitleRect) Linesize(plane int) int {
	if plane < 0 || plane >= 4 {
		return 0
	}
	return int(rect.linesize[plane])
}

// Get text content for SUBTITLE_TEXT or SUBTITLE_ASS types.
// Returns the text field for SUBTITLE_TEXT or the ass field for SUBTITLE_ASS.
// Returns empty string for other types (BITMAP, NONE).
func (rect *AVSubtitleRect) Text() string {
	switch rect.Type() {
	case SUBTITLE_TEXT:
		if rect.text == nil {
			return ""
		}
		return C.GoString(rect.text)
	case SUBTITLE_ASS:
		if rect.ass == nil {
			return ""
		}
		return C.GoString(rect.ass)
	default:
		return ""
	}
}

func (rect *AVSubtitleRect) Flags() int {
	return int(rect.flags)
}

func (rect *AVSubtitleRect) SetFlags(flags int) {
	rect.flags = C.int(flags)
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (t AVSubtitleType) String() string {
	switch t {
	case SUBTITLE_NONE:
		return "SUBTITLE_NONE"
	case SUBTITLE_BITMAP:
		return "SUBTITLE_BITMAP"
	case SUBTITLE_TEXT:
		return "SUBTITLE_TEXT"
	case SUBTITLE_ASS:
		return "SUBTITLE_ASS"
	default:
		return "[AVSubtitleType]"
	}
}

func (t AVSubtitleType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (sub *AVSubtitle) String() string {
	return marshalToString(sub)
}

func (sub *AVSubtitle) MarshalJSON() ([]byte, error) {
	type jsonSubtitleRect struct {
		Type      AVSubtitleType `json:"type"`
		X         int            `json:"x,omitempty"`
		Y         int            `json:"y,omitempty"`
		Width     int            `json:"width,omitempty"`
		Height    int            `json:"height,omitempty"`
		NumColors int            `json:"num_colors,omitempty"`
		Text      string         `json:"text,omitempty"`
	}

	type jsonAVSubtitle struct {
		Format           uint16             `json:"format"`
		StartDisplayTime uint32             `json:"start_display_time_ms"`
		EndDisplayTime   uint32             `json:"end_display_time_ms"`
		NumRects         uint               `json:"num_rects"`
		PTS              int64              `json:"pts"`
		Rects            []jsonSubtitleRect `json:"rects,omitempty"`
	}

	result := jsonAVSubtitle{
		Format:           sub.Format(),
		StartDisplayTime: sub.StartDisplayTime(),
		EndDisplayTime:   sub.EndDisplayTime(),
		NumRects:         sub.NumRects(),
		PTS:              sub.PTS(),
	}

	// Add rectangle information
	rects := sub.Rects()
	if len(rects) > 0 {
		result.Rects = make([]jsonSubtitleRect, len(rects))
		for i, rect := range rects {
			result.Rects[i] = jsonSubtitleRect{
				Type:      rect.Type(),
				X:         rect.X(),
				Y:         rect.Y(),
				Width:     rect.Width(),
				Height:    rect.Height(),
				NumColors: rect.NumColors(),
				Text:      rect.Text(),
			}
		}
	}

	return json.Marshal(result)
}

func (rect *AVSubtitleRect) String() string {
	return marshalToString(rect)
}

func (rect *AVSubtitleRect) MarshalJSON() ([]byte, error) {
	type jsonAVSubtitleRect struct {
		Type      AVSubtitleType `json:"type"`
		X         int            `json:"x,omitempty"`
		Y         int            `json:"y,omitempty"`
		Width     int            `json:"width,omitempty"`
		Height    int            `json:"height,omitempty"`
		NumColors int            `json:"num_colors,omitempty"`
		Text      string         `json:"text,omitempty"`
		Flags     int            `json:"flags,omitempty"`
	}

	return json.Marshal(jsonAVSubtitleRect{
		Type:      rect.Type(),
		X:         rect.X(),
		Y:         rect.Y(),
		Width:     rect.Width(),
		Height:    rect.Height(),
		NumColors: rect.NumColors(),
		Text:      rect.Text(),
		Flags:     rect.Flags(),
	})
}
