package schema

import (
	"encoding/json"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Packet struct {
	*ff.AVPacket
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewPacket(pkt *ff.AVPacket) *Packet {
	if pkt == nil {
		return nil
	}
	return &Packet{AVPacket: pkt}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (p *Packet) MarshalJSON() ([]byte, error) {
	if p.AVPacket == nil {
		return json.Marshal(nil)
	}
	return p.AVPacket.MarshalJSON()
}

func (p *Packet) String() string {
	if p == nil || p.AVPacket == nil {
		return "<nil>"
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Ts returns the timestamp in seconds, or -1.0 if the timestamp
// is undefined or timebase is not set
func (p *Packet) Ts() float64 {
	if p == nil || p.AVPacket == nil {
		return -1.0
	}
	pts := p.AVPacket.Pts()
	if pts == int64(ff.AV_NOPTS_VALUE) {
		return -1.0
	}
	tb := p.AVPacket.TimeBase()
	if tb.Num() == 0 || tb.Den() == 0 {
		return -1.0
	}
	return ff.AVUtil_rational_q2d(tb) * float64(pts)
}
