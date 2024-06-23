package media

import (
	"encoding/json"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type packetmeta struct {
	StreamIndex int            `json:"stream_index" writer:",width:10,right"`
	MediaType   ff.AVMediaType `json:"media_type" writer:",width:20"`
	Size        int            `json:"size,omitempty"  writer:",width:7,right"`
	Pts         ff.AVTimestamp `json:"pts,omitempty"  writer:",width:9,right"`
	TimeBase    ff.AVRational  `json:"time_base,omitempty"  writer:",width:10,right"`
	Duration    ff.AVTimestamp `json:"duration,omitempty"  writer:",width:10,right"`
	Pos         *int64         `json:"pos,omitempty"  writer:",width:10,right"`
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
			StreamIndex: stream,
			MediaType:   t,
		},
	}
	if ctx != nil {
		pkt.packetmeta.Size = ctx.Size()
		pkt.packetmeta.Pts = ff.AVTimestamp(ctx.Pts())
		pkt.packetmeta.TimeBase = timeBase
		pkt.packetmeta.Duration = ff.AVTimestamp(ctx.Duration())
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
