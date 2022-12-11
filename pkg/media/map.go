package media

import (
	"context"

	// Packages
	multierror "github.com/hashicorp/go-multierror"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type decodemap struct {
	packet  *packet
	input   *input
	context map[int]decodecontext
}

type decodecontext struct {
	Stream *stream
}

// Ensure decodemap complies with Map interface
var _ Map = (*decodemap)(nil)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	mediaTypes = []MediaFlag{
		MEDIA_FLAG_AUDIO, MEDIA_FLAG_VIDEO,
		MEDIA_FLAG_SUBTITLE,
		MEDIA_FLAG_DATA, MEDIA_FLAG_ATTACHMENT,
	}
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewMap(media Media, flags MediaFlag) (*decodemap, error) {
	m := new(decodemap)

	// Set input media
	if input, ok := media.(*input); !ok || input == nil {
		return nil, ErrBadParameter.With("media")
	} else {
		m.input = input
	}

	// Create map of streams
	m.context = streamsByType(m.input, flags)

	// Create packet
	if packet := NewPacket(func(i int) Stream {
		return m.context[i].Stream
	}); packet == nil {
		return nil, ErrInternalAppError.With("NewPacket")
	} else {
		m.packet = packet
	}

	// Return success
	return m, nil
}

func (m *decodemap) Close() error {
	var result error
	if m.packet != nil {
		if err := m.packet.Close(); err != nil {
			result = multierror.Append(result, err)
		}
		m.packet = nil
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return a reusable packet which is used for decoding
func (m *decodemap) Packet() Packet {
	return m.packet
}

// Return the input media
func (m *decodemap) Input() Media {
	if m.input == nil || m.input.ctx == nil {
		return nil
	}
	return m.input
}

// Return the input media streams which should be decoded
func (m *decodemap) Streams() []Stream {
	var result []Stream
	for _, stream := range m.context {
		result = append(result, stream.Stream)
	}
	return result
}

// Decode a packet, by calling a decoding function with a packet.
// If the stream associated with the packet is not in the map, then
// ignore it
func (m *decodemap) Decode(ctx context.Context, p Packet, fn DecodeFn) error {
	index := p.(*packet).StreamIndex()
	if _, exists := m.context[index]; exists {
		return fn(ctx, p)
	} else {
		return nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Return streams of a given type for input media
func streamsByType(input *input, media_type MediaFlag) map[int]decodecontext {
	if input == nil || input.ctx == nil {
		return nil
	}
	result := make(map[int]decodecontext, input.ctx.NumStreams())
	for _, t := range mediaTypes {
		if !media_type.Is(t) {
			continue
		} else if f := toAVMediaType(t); f == ffmpeg.AVMEDIA_TYPE_UNKNOWN {
			continue
		} else if n, err := ffmpeg.AVFormat_av_find_best_stream(input.ctx, f, -1, -1, nil, 0); err != nil {
			continue
		} else if str, exists := input.streams[n]; exists {
			result[n] = decodecontext{Stream: str.(*stream)}
		}
	}

	// Return streams
	return result
}

func toAVMediaType(media_type MediaFlag) ffmpeg.AVMediaType {
	switch media_type {
	case MEDIA_FLAG_AUDIO:
		return ffmpeg.AVMEDIA_TYPE_AUDIO
	case MEDIA_FLAG_VIDEO:
		return ffmpeg.AVMEDIA_TYPE_VIDEO
	case MEDIA_FLAG_SUBTITLE:
		return ffmpeg.AVMEDIA_TYPE_SUBTITLE
	case MEDIA_FLAG_DATA:
		return ffmpeg.AVMEDIA_TYPE_DATA
	case MEDIA_FLAG_ATTACHMENT:
		return ffmpeg.AVMEDIA_TYPE_ATTACHMENT
	default:
		return ffmpeg.AVMEDIA_TYPE_UNKNOWN
	}
}
