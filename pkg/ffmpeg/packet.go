package ffmpeg

import (
	"encoding/json"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Packet ff.AVPacket

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (packet *Packet) String() string {
	if packet == nil {
		return "<nil>"
	}
	data, _ := json.MarshalIndent((*ff.AVPacket)(packet), "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return stream index for this packet
func (packet *Packet) Stream() int {
	if packet == nil {
		return -1
	}
	return (*ff.AVPacket)(packet).StreamIndex()
}

// Return the timestamp in seconds, or TS_UNDEFINED if the timestamp
// is undefined or timebase is not set
func (packet *Packet) Ts() float64 {
	if packet == nil {
		return TS_UNDEFINED
	}
	ctx := (*ff.AVPacket)(packet)
	pts := ctx.Pts()
	if pts == int64(ff.AV_NOPTS_VALUE) {
		return TS_UNDEFINED
	}
	tb := ctx.TimeBase()
	if tb.Num() == 0 || tb.Den() == 0 {
		return TS_UNDEFINED
	}
	return ff.AVUtil_rational_q2d(tb) * float64(pts)
}

// Return presentation timestamp (PTS)
func (packet *Packet) Pts() int64 {
	if packet == nil {
		return int64(ff.AV_NOPTS_VALUE)
	}
	return (*ff.AVPacket)(packet).Pts()
}

// Return decode timestamp (DTS)
func (packet *Packet) Dts() int64 {
	if packet == nil {
		return int64(ff.AV_NOPTS_VALUE)
	}
	return (*ff.AVPacket)(packet).Dts()
}

// Return packet duration
func (packet *Packet) Duration() int64 {
	if packet == nil {
		return 0
	}
	return (*ff.AVPacket)(packet).Duration()
}

// Return packet size in bytes
func (packet *Packet) Size() int {
	if packet == nil {
		return 0
	}
	return (*ff.AVPacket)(packet).Size()
}

// Return packet data as bytes
func (packet *Packet) Bytes() []byte {
	if packet == nil {
		return nil
	}
	return (*ff.AVPacket)(packet).Bytes()
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Create a new packet wrapper from an AVPacket pointer
func newPacket(pkt *ff.AVPacket) *Packet {
	return (*Packet)(pkt)
}
