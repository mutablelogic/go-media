package schema

import (
	"encoding/json"
	"image"

	// Packages
	uuid "github.com/google/uuid"
	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// StreamProfile describes a stream already discovered in an opened media
// file (via a Reader), rather than a codec configuration to encode with.
// Unlike AudioProfile/VideoProfile/SubtitleProfile, there is nothing here to
// validate or assemble: the underlying AVCodecParameters were already fully
// populated by the demuxer, so StreamProfile just wraps them.
//
// A StreamProfile is valid only while the Reader it came from remains open:
// Par() aliases the AVCodecParameters extradata (SPS/PPS, OpusHead, ...)
// owned by the underlying AVFormatContext, and Codec() the AVCodec found for
// its codec ID. This is not a limitation in practice — the typical use (e.g.
// feeding writer.WithProfile for a remux) requires the source Reader to stay
// open anyway, since packets still need to be read from it.
type StreamProfile struct {
	index       int
	disposition ff.AVDisposition
	codec       *ff.AVCodec          // Decoder for this stream's codec, if any is registered
	par         ff.AVCodecParameters // Codec parameters, copied from the stream
	timebase    ff.AVRational        // Stream timebase, as demuxed
	metadata    []gomedia.Metadata   // Per-stream tags, e.g. "language", "title"
}

var _ Profile = (*StreamProfile)(nil)

// streamMeta is a minimal gomedia.Metadata implementation for a per-stream
// tag (language, title, handler_name, ...). Unlike container-level
// metadata, per-stream tags are always plain text, never artwork.
type streamMeta struct {
	key   string
	value string
}

var _ gomedia.Metadata = streamMeta{}

func (m streamMeta) Key() string        { return m.key }
func (m streamMeta) Value() string      { return m.value }
func (m streamMeta) Bytes() []byte      { return []byte(m.value) }
func (m streamMeta) Image() image.Image { return nil }
func (m streamMeta) Any() any           { return m.value }

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewStreamProfile describes an existing AVStream (typically obtained from
// an opened reader.Reader) as a Profile. Returns an error for stream types
// that can't be meaningfully described this way: data and attachment
// streams (fonts, generic metadata, ...), and attached-pic streams (cover
// art, already surfaced separately via Reader.Metadata's "artwork" key).
func NewStreamProfile(stream *ff.AVStream) (*StreamProfile, error) {
	if stream == nil {
		return nil, gomedia.ErrBadParameter.With("stream is nil")
	}

	par := stream.CodecPar()
	if par == nil {
		return nil, gomedia.ErrBadParameter.With("stream has no codec parameters")
	}

	switch par.CodecType() {
	case ff.AVMEDIA_TYPE_AUDIO, ff.AVMEDIA_TYPE_VIDEO, ff.AVMEDIA_TYPE_SUBTITLE:
		// Supported
	default:
		return nil, gomedia.ErrBadParameter.Withf("stream %d: unsupported stream type %q", stream.Index(), CodecType(par.CodecType()))
	}

	if stream.Disposition().Is(ff.AV_DISPOSITION_ATTACHED_PIC) {
		return nil, gomedia.ErrBadParameter.Withf("stream %d: attached-pic streams are not represented as a StreamProfile", stream.Index())
	}

	entries := ff.AVUtil_dict_entries(stream.Metadata())
	metadata := make([]gomedia.Metadata, 0, len(entries))
	for _, entry := range entries {
		metadata = append(metadata, streamMeta{key: entry.Key(), value: entry.Value()})
	}

	self := &StreamProfile{
		index:       stream.Index(),
		disposition: stream.Disposition(),
		codec:       ff.AVCodec_find_decoder(par.CodecID()),
		par:         *par,
		timebase:    stream.TimeBase(),
		metadata:    metadata,
	}

	return self, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r StreamProfile) String() string {
	return types.Stringify(r)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PROFILE INTERFACE

// UUID is always the zero value: a StreamProfile describes a stream
// discovered in a file, not a persisted, identifiable configuration.
func (r StreamProfile) UUID() uuid.UUID {
	return uuid.UUID{}
}

// Type is read directly from the stream's codec parameters, so it's always
// available even if no decoder is registered for the codec.
func (r StreamProfile) Type() CodecType {
	return CodecType(r.par.CodecType())
}

// Codec is nil if no decoder is registered for this codec in the current
// build. Par/TimeBase/Type remain valid either way.
func (r StreamProfile) Codec() *Codec {
	if r.codec == nil {
		return nil
	}
	return NewCodec(r.codec)
}

func (r StreamProfile) Par() *ff.AVCodecParameters {
	return types.Ptr(r.par)
}

func (r StreamProfile) TimeBase() *ff.AVRational {
	if r.timebase.Num() == 0 {
		return nil
	}
	return types.Ptr(r.timebase)
}

// Options is always nil: there is nothing to configure on a stream that's
// already been demuxed.
func (r StreamProfile) Options() json.RawMessage {
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Index returns the stream's index within its container, e.g. for Seek.
func (r StreamProfile) Index() int {
	return r.index
}

// Disposition returns the stream's disposition flags (default, forced,
// attached pic, ...).
func (r StreamProfile) Disposition() ff.AVDisposition {
	return r.disposition
}

// Metadata returns the stream's own tags — e.g. "language" (ISO 639-2,
// such as "eng"), "title", "handler_name" — as opposed to Reader.Metadata,
// which returns container-level tags. Unlike WithProfile's Options, this is
// read-only: there is no way to set per-stream metadata via a Profile.
func (r StreamProfile) Metadata() []gomedia.Metadata {
	return r.metadata
}
