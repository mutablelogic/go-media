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
// an opened reader.Reader) as a Profile - audio, video, subtitle, data and
// attachment streams alike, so a caller remuxing a whole container doesn't
// silently drop any of them. Returns an error for attached-pic streams
// (cover art, already surfaced separately via Reader.Metadata's "artwork"
// key), since representing it as a stream too would duplicate that.
func NewStreamProfile(stream *ff.AVStream) (*StreamProfile, error) {
	if stream == nil {
		return nil, gomedia.ErrBadParameter.With("stream is nil")
	}

	par := stream.CodecPar()
	if par == nil {
		return nil, gomedia.ErrBadParameter.With("stream has no codec parameters")
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

// MarshalJSON is required because every field of StreamProfile is
// unexported (see the type's doc comment) - without it, encoding/json has
// nothing to marshal and every stream serializes as "{}". ProfileMeta is
// built from the Profile interface, which StreamProfile fully implements.
func (r StreamProfile) MarshalJSON() ([]byte, error) {
	return json.Marshal(NewProfileMeta(r))
}

// UnmarshalJSON is the counterpart to MarshalJSON - without it, a client
// decoding a StreamProfile from JSON would silently get a zero-value struct
// (every field is unexported, so encoding/json has nothing to write into)
// that then re-marshals as bogus data (e.g. a zero AVCodecParameters, whose
// zero-valued codec_type of 0 happens to collide with AVMEDIA_TYPE_VIDEO).
//
// codec is resolved from CodecMeta's own name, which round-trips exactly
// (unlike a codec ID's generic name, which can differ from its default
// decoder's own registered name - e.g. ID "mp3"'s default decoder is
// "mp3float"; see AVCodecID.UnmarshalJSON). par is then rebuilt field by
// field from whichever of ProfileMetaAudio/ProfileMetaVideo is present,
// since ProfileMeta no longer carries a raw AVCodecParameters.
func (r *StreamProfile) UnmarshalJSON(data []byte) error {
	meta := new(ProfileMeta)
	if err := json.Unmarshal(data, meta); err != nil {
		return err
	}

	if meta.ProfileMetaStream != nil {
		r.index = meta.Index
		r.disposition = meta.Disposition
		if meta.TimeBase != nil {
			r.timebase = *meta.TimeBase
		}
		r.metadata = metadataFromMetaList(meta.Metadata)
	}

	var par ff.AVCodecParameters
	par.SetCodecType(ff.AVMediaType(meta.Type))

	if meta.Codec != nil {
		if codec := ff.AVCodec_find_decoder_by_name(meta.Codec.Name); codec != nil {
			r.codec = codec
		} else if codec := ff.AVCodec_find_encoder_by_name(meta.Codec.Name); codec != nil {
			r.codec = codec
		}
		// r.codec stays nil above whenever this build has no decoder or
		// encoder registered for the name (e.g. DVB teletext/EPG streams) -
		// but the ID itself is still resolvable from the name via ffmpeg's
		// static codec descriptor table, so par.CodecID() doesn't need a
		// registered codec to round-trip correctly either.
		if r.codec != nil {
			par.SetCodecID(r.codec.ID())
		} else {
			par.SetCodecID(ff.AVCodecID_from_name(meta.Codec.Name))
		}
	}

	switch {
	case meta.ProfileMetaAudio != nil:
		a := meta.ProfileMetaAudio
		if a.Bitrate != nil {
			par.SetBitRate(int64(*a.Bitrate))
		}
		if a.SampleRate != nil {
			par.SetSampleRate(int(*a.SampleRate))
		}
		if a.SampleFormat != nil {
			par.SetSampleFormat(ff.AVUtil_get_sample_fmt(*a.SampleFormat))
		}
		if a.ChannelLayout != nil {
			var ch ff.AVChannelLayout
			if err := ff.AVUtil_channel_layout_from_string(&ch, *a.ChannelLayout); err == nil {
				_ = par.SetChannelLayout(ch)
			}
		}
		if a.Profile != nil && r.codec != nil {
			if id, err := resolveProfileID(r.codec, *a.Profile); err == nil {
				par.SetProfile(id)
			}
		}
	case meta.ProfileMetaVideo != nil:
		v := meta.ProfileMetaVideo
		if v.Bitrate != nil {
			par.SetBitRate(int64(*v.Bitrate))
		}
		if v.Width != nil {
			par.SetWidth(int(*v.Width))
		}
		if v.Height != nil {
			par.SetHeight(int(*v.Height))
		}
		if v.PixelFormat != nil {
			par.SetPixelFormat(ff.AVUtil_get_pix_fmt(*v.PixelFormat))
		}
		if v.Profile != nil && r.codec != nil {
			if id, err := resolveProfileID(r.codec, *v.Profile); err == nil {
				par.SetProfile(id)
			}
		}
	}

	r.par = par
	return nil
}

// metadataFromMetaList is the reverse of newMetadataMetaList.
func metadataFromMetaList(entries []MetadataMeta) []gomedia.Metadata {
	if len(entries) == 0 {
		return nil
	}
	result := make([]gomedia.Metadata, 0, len(entries))
	for _, e := range entries {
		result = append(result, streamMeta{key: e.Key, value: e.Value})
	}
	return result
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
