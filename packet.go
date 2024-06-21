package media

import (
	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type packet struct {
	X string `json:"type"`
	*ff.AVPacket
}

var _ Packet = (*packet)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newPacket(ctx *ff.AVPacket) *packet {
	return &packet{"X", ctx}
}
