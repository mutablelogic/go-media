package media

import (
	"fmt"
	"time"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	//. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type packet struct {
	ctx *ffmpeg.AVPacket
	fn  func(int) Stream
}

// Ensure *input complies with Media interface
var _ Packet = (*packet)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewPacket(fn func(int) Stream) *packet {
	packet := new(packet)

	if ctx := ffmpeg.AVCodec_av_packet_alloc(); ctx == nil {
		return nil
	} else {
		packet.ctx = ctx
		packet.fn = fn
	}

	// Return success
	return packet
}

func (packet *packet) Close() error {
	var result error

	// Callback
	if packet.ctx != nil {
		ffmpeg.AVCodec_av_packet_free_ptr(packet.ctx)
		packet.ctx = nil
		packet.fn = nil
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (packet *packet) String() string {
	str := "<media.packet"
	if flags := packet.Flags(); flags != MEDIA_FLAG_NONE {
		str += fmt.Sprint(" flags=", flags)
	}
	if stream := packet.Stream(); stream != nil {
		str += fmt.Sprint(" stream_index=", stream.Index())
	}
	if duration := packet.Duration(); duration != 0 {
		str += fmt.Sprint(" duration=", duration)
	}
	if pos := packet.Pos(); pos != -1 {
		str += fmt.Sprint(" pos=", pos)
	}
	if size := packet.Size(); size > 0 {
		str += fmt.Sprint(" size=", size)
	}
	if key := packet.IsKeyFrame(); key {
		str += " contains_keyframe"
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Unref releases the packet
func (packet *packet) Release() {
	if packet.ctx != nil {
		ffmpeg.AVCodec_av_packet_unref(packet.ctx)
	}
}

// Flags returns the  flags from the stream
func (packet *packet) Flags() MediaFlag {
	if stream := packet.Stream(); stream != nil {
		return stream.Flags()
	} else {
		return MEDIA_FLAG_NONE
	}
}

// StreamIndex returns the stream which the packet belongs to
func (packet *packet) StreamIndex() int {
	if packet.ctx == nil {
		return -1
	}
	return packet.ctx.StreamIndex()
}

// Stream returns the stream which the packet belongs to
func (packet *packet) Stream() Stream {
	if packet.ctx == nil || packet.fn == nil {
		return nil
	}
	return packet.fn(packet.ctx.StreamIndex())
}

// IsKeyFrame returns true if the packet contains a key frame
func (packet *packet) IsKeyFrame() bool {
	if packet.ctx == nil {
		return false
	}
	return packet.ctx.Flags()&ffmpeg.AV_PKT_FLAG_KEY != 0
}

// Pos returns the byte position of the packet in the media
func (packet *packet) Pos() int64 {
	if packet.ctx == nil {
		return -1
	}
	return packet.ctx.Pos()
}

// Size returns the size of the packet
func (packet *packet) Size() int {
	if packet.ctx == nil {
		return 0
	}
	return packet.ctx.Size()
}

// Duration returns the duration of the packet
func (packet *packet) Duration() time.Duration {
	if packet.ctx == nil {
		return -1
	}
	duration := packet.ctx.Duration()
	if duration == 0 {
		return 0
	} else if stream, ok := packet.Stream().(*stream); !ok {
		return 0
	} else {
		tb := stream.ctx.TimeBase()
		return time.Second * time.Duration(duration) * time.Duration(tb.Num()) / time.Duration(tb.Den())
	}
}

// Bytes returns the raw bytes of the packet
func (packet *packet) Bytes() []byte {
	return packet.ctx.Bytes()
}
