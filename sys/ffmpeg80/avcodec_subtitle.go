package ffmpeg

import (
	"encoding/json"
	"errors"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec libavutil
#include <libavcodec/avcodec.h>
#include <libavutil/mem.h>
#include <stdlib.h>
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

////////////////////////////////////////////////////////////////////////////////
// AVSubtitle CONSTRUCTION (for encoding)

// NewSubtitle creates an empty subtitle with the given PTS (in AV_TIME_BASE
// units, matching AVSubtitle's own documented convention). Use SetText to
// add content before passing it to AVCodec_encode_subtitle.
func NewSubtitle(pts int64) *AVSubtitle {
	sub := &AVSubtitle{}
	sub.pts = C.int64_t(pts)
	return sub
}

// SetText replaces any existing rects with a single SUBTITLE_TEXT rect
// containing text, displayed from startMs to endMs. Use this for plain-text
// subtitle codecs (srt, webvtt, mov_text, ...).
func (sub *AVSubtitle) SetText(text string, startMs, endMs uint32) error {
	return sub.setRect(SUBTITLE_TEXT, text, startMs, endMs)
}

// SetASS replaces any existing rects with a single SUBTITLE_ASS rect, for
// the "ass" subtitle codec. dialogue is the event's comma-separated field
// list as used internally by FFmpeg's ASS handling - everything a
// "Dialogue:" line carries after the timing fields, e.g.
// "0,Default,,0,0,0,,Hello world" (Layer,Style,Name,MarginL,MarginR,MarginV,Effect,Text).
func (sub *AVSubtitle) SetASS(dialogue string, startMs, endMs uint32) error {
	return sub.setRect(SUBTITLE_ASS, dialogue, startMs, endMs)
}

// setRect replaces any existing rects with a single rect of the given type
// and text/ass content (SUBTITLE_TEXT uses the rect's text field,
// SUBTITLE_ASS its ass field). The rect and its content are allocated with
// FFmpeg's own allocator, so AVSubtitle_free can release them the same way
// it releases a decoded subtitle's rects.
func (sub *AVSubtitle) setRect(t AVSubtitleType, content string, startMs, endMs uint32) error {
	// Discard any existing rects before replacing them
	if sub.num_rects > 0 {
		C.avsubtitle_free((*C.struct_AVSubtitle)(sub))
	}

	rect := (*C.struct_AVSubtitleRect)(C.av_mallocz(C.size_t(unsafe.Sizeof(C.struct_AVSubtitleRect{}))))
	if rect == nil {
		return errors.New("failed to allocate subtitle rect")
	}
	rect._type = C.enum_AVSubtitleType(t)

	cContent := C.CString(content)
	avContent := C.av_strdup(cContent)
	C.free(unsafe.Pointer(cContent))
	if avContent == nil {
		C.av_free(unsafe.Pointer(rect))
		return errors.New("failed to allocate subtitle content")
	}
	switch t {
	case SUBTITLE_ASS:
		rect.ass = avContent
	default:
		rect.text = avContent
	}

	rects := (**C.struct_AVSubtitleRect)(C.av_malloc(C.size_t(unsafe.Sizeof(rect))))
	if rects == nil {
		C.av_freep(unsafe.Pointer(&avContent))
		C.av_free(unsafe.Pointer(rect))
		return errors.New("failed to allocate subtitle rects array")
	}
	*rects = rect

	sub.rects = rects
	sub.num_rects = 1
	sub.start_display_time = C.uint32_t(startMs)
	sub.end_display_time = C.uint32_t(endMs)

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
