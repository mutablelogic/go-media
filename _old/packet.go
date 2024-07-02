package media

import (
	"encoding/json"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type packetmeta struct {
	Stream    int            `json:"stream" writer:",width:10,right"`
	MediaType ff.AVMediaType `json:"media_type" writer:",width:20"`
	Size      int            `json:"size,omitempty"  writer:",width:7,right"`
	Pts       int64          `json:"pts,omitempty"  writer:",width:9,right"`
	TimeBase  ff.AVRational  `json:"time_base,omitempty"  writer:",width:10,right"`
	Duration  int64          `json:"duration,omitempty"  writer:",width:10,right"`
	Pos       *int64         `json:"pos,omitempty"  writer:",width:10,right"`
}

type packet struct {
	packetmeta
	ctx *ff.AVPacket
}

var _ Packet = (*packet)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newPacket(ctx *ff.AVPacket, stream int, t ff.AVMediaType, timeBase ff.AVRational) *packet {
	pkt := &packet{
		ctx: ctx,
		packetmeta: packetmeta{
			Stream:    stream,
			MediaType: t,
		},
	}
	if ctx != nil {
		pkt.packetmeta.Size = ctx.Size()
		pkt.packetmeta.Pts = ctx.Pts()
		pkt.packetmeta.TimeBase = timeBase
		pkt.packetmeta.Duration = ctx.Duration()
		if ctx.Pos() != -1 {
			pos := ctx.Pos()
			pkt.packetmeta.Pos = &pos
		}
	}
	return pkt
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (packet *packet) MarshalJSON() ([]byte, error) {
	return json.Marshal(packet.packetmeta)
}

func (packet *packet) String() string {
	data, _ := json.MarshalIndent(packet, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (packet *packet) Id() int {
	return packet.packetmeta.Stream
}

func (packet *packet) Type() MediaType {
	switch packet.packetmeta.MediaType {
	case ff.AVMEDIA_TYPE_AUDIO:
		return AUDIO
	case ff.AVMEDIA_TYPE_VIDEO:
		return VIDEO
	case ff.AVMEDIA_TYPE_SUBTITLE:
		return SUBTITLE
	default:
		return DATA
	}
}
