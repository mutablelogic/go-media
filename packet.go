package media

import (
	"encoding/json"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type packet struct {
	ctx *ff.AVPacket
}

var _ Packet = (*packet)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newPacket(ctx *ff.AVPacket) *packet {
	return &packet{ctx}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (packet *packet) MarshalJSON() ([]byte, error) {
	return json.Marshal(packet.ctx)
}

func (packet *packet) String() string {
	data, _ := json.MarshalIndent(packet, "", "  ")
	return string(data)
}
