package media

import (
	"context"
	"fmt"
	"io"

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
	context map[int]*mapentry
}

type mapentry struct {
	Decoder   *decoder   // Decoder context for the stream
	Resampler *resampler // Resampler context for the audio frames
	Scaler    *scaler    // Scaler context for the video frames
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
	if flags == MEDIA_FLAG_NONE {
		flags = MEDIA_FLAG_AUDIO | MEDIA_FLAG_VIDEO | MEDIA_FLAG_SUBTITLE
	}
	m.context = streamsByType(m.input, flags)
	if len(m.context) == 0 {
		return nil, ErrNotFound.With("No streams of type: ", flags)
	}

	// Create packet
	if packet := NewPacket(func(i int) Stream {
		return m.context[i].Decoder.stream
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

	// Close packet
	if m.packet != nil {
		if err := m.packet.Close(); err != nil {
			result = multierror.Append(result, err)
		}
		m.packet = nil
	}

	// Close all decoders
	for _, mapentry := range m.context {
		if err := mapentry.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	m.context = nil

	// Return any errors
	return result
}

func (mapentry *mapentry) Close() error {
	var result error

	// Call close on all objects
	if mapentry.Decoder != nil {
		if err := mapentry.Decoder.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if mapentry.Resampler != nil {
		if err := mapentry.Resampler.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if mapentry.Scaler != nil {
		if err := mapentry.Scaler.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Release resources
	mapentry.Decoder = nil
	mapentry.Resampler = nil
	mapentry.Scaler = nil

	// Return any errors
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
	for _, mapentry := range m.context {
		if mapentry.Decoder != nil {
			result = append(result, mapentry.Decoder.stream)
		}
	}
	return result
}

// Decode a packet, by calling a decoding function with a packet.
// If the stream associated with the packet is not in the map, then
// ignore it
func (m *decodemap) Demux(ctx context.Context, p Packet, fn DemuxFn) error {
	index := p.(*packet).StreamIndex()
	mapentry, exists := m.context[index]
	if !exists {
		return nil
	}
	// Send a packet into the decoder
	if err := ffmpeg.AVCodec_send_packet(mapentry.Decoder.ctx, p.(*packet).ctx); err != nil {
		return err
	} else {
		return fn(ctx, p)
	}
}

// PrintMap will print out a summary of the mapping
func (m *decodemap) PrintMap(w io.Writer) {
	for id := range m.context {
		stream := m.input.streams[id]
		fmt.Fprintf(w, "Stream %2d (%s): %s\n", id, toMediaType(stream.Flags()), stream)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Return media type (audio, video, subtitle, etc)
func toMediaType(flag MediaFlag) string {
	switch {
	case flag.Is(MEDIA_FLAG_AUDIO):
		return "audio"
	case flag.Is(MEDIA_FLAG_VIDEO):
		return "video"
	case flag.Is(MEDIA_FLAG_SUBTITLE):
		return "subtitle"
	case flag.Is(MEDIA_FLAG_DATA):
		return "data"
	case flag.Is(MEDIA_FLAG_ATTACHMENT):
		return "attachment"
	default:
		return "other"
	}
}

// Return streams of a given type for input media
func streamsByType(input *input, media_type MediaFlag) map[int]*mapentry {
	if input == nil || input.ctx == nil {
		return nil
	}
	result := make(map[int]*mapentry, input.ctx.NumStreams())
	for _, t := range mediaTypes {
		if !media_type.Is(t) {
			continue
		} else if f := toAVMediaType(t); f == ffmpeg.AVMEDIA_TYPE_UNKNOWN {
			continue
		} else if n, err := ffmpeg.AVFormat_av_find_best_stream(input.ctx, f, -1, -1, nil, 0); err != nil {
			continue
		} else if str, exists := input.streams[n]; !exists {
			continue
		} else if decoder := NewDecoder(str.(*stream)); decoder == nil {
			continue
		} else {
			result[n] = &mapentry{
				Decoder: decoder,
			}
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
