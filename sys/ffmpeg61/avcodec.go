package ffmpeg

////////////////////////////////////////////////////////////////////////////////
// CGO

/*
#cgo pkg-config: libavcodec
#include <libavcodec/avcodec.h>
*/
import "C"
import "encoding/json"

////////////////////////////////////////////////////////////////////////////////
// TYPES

type (
	AVPacket C.struct_AVPacket
)

type jsonAVPacket struct {
	Pts           int64 `json:"pts,omitempty"`
	Dts           int64 `json:"dts,omitempty"`
	Size          int   `json:"size,omitempty"`
	StreamIndex   int   `json:"stream_index"` // Stream index starts at 0
	Flags         int   `json:"flags,omitempty"`
	SideDataElems int   `json:"side_data_elems,omitempty"`
	Duration      int64 `json:"duration,omitempty"`
	Pos           int64 `json:"pos,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// JSON OUTPUT

func (ctx AVPacket) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAVPacket{
		Pts:           int64(ctx.pts),
		Dts:           int64(ctx.dts),
		Size:          int(ctx.size),
		StreamIndex:   int(ctx.stream_index),
		Flags:         int(ctx.flags),
		SideDataElems: int(ctx.side_data_elems),
		Duration:      int64(ctx.duration),
		Pos:           int64(ctx.pos),
	})
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (ctx AVPacket) String() string {
	if str, err := json.MarshalIndent(ctx, "", "  "); err != nil {
		return err.Error()
	} else {
		return string(str)
	}
}
