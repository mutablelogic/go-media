package schema

import (
	"encoding/json"

	// Packages
	uuid "github.com/google/uuid"
	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Profile interface {
	UUID() uuid.UUID            // Unique identifier for this profile
	Type() CodecType            // Type for this profile
	Codec() *Codec              // Codec for this profile
	Par() *ff.AVCodecParameters // Codec parameters for this profile
	TimeBase() *ff.AVRational   // Time base for this profile
	Options() json.RawMessage   // Additional codec-specific options
}

// ProfileMetaAudio holds the audio-specific fields of a ProfileMeta.
type ProfileMetaAudio struct {
	Bitrate       *uint64 `json:"bitrate,omitempty"`        // bps
	Profile       *string `json:"profile,omitempty"`        // Codec profile; "LC", "HE-AAC", ...
	SampleRate    *uint64 `json:"sample_rate,omitempty"`    // Hz
	SampleFormat  *string `json:"sample_format,omitempty"`  // Audio sample format; "fltp", "s16"
	ChannelLayout *string `json:"channel_layout,omitempty"` // Audio channel layout; "mono", "stereo"
}

// ProfileMetaVideo holds the video-specific fields of a ProfileMeta.
type ProfileMetaVideo struct {
	Bitrate     *uint64  `json:"bitrate,omitempty"`      // bps
	Profile     *string  `json:"profile,omitempty"`      // Codec profile; "high", "main", "baseline"
	Width       *uint64  `json:"width,omitempty"`        // Frame width in pixels
	Height      *uint64  `json:"height,omitempty"`       // Frame height in pixels
	PixelFormat *string  `json:"pixel_format,omitempty"` // Video pixel format; "yuv420p", "nv12"
	FrameRate   *float64 `json:"frame_rate,omitempty"`   // Frames per second
}

// ProfileMetaStream holds the fields specific to a stream already discovered
// in a media file - as opposed to a codec configuration to encode with -
// so it's only populated for a StreamProfile.
type ProfileMetaStream struct {
	Index       int              `json:"index"`
	Disposition ff.AVDisposition `json:"disposition,omitempty"`
	TimeBase    *ff.AVRational   `json:"timebase,omitempty"`
	Metadata    []MetadataMeta   `json:"metadata,omitempty"`
}

// MetadataMeta is a JSON-friendly key/value view of a gomedia.Metadata
// entry - binary payloads (e.g. artwork) aren't represented here.
type MetadataMeta struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

// ProfileMeta is a JSON representation of a Profile, uniform across
// AudioProfile, VideoProfile, SubtitleProfile, and StreamProfile.
// Audio/Video carry their type-specific fields; Subtitle has none beyond
// Options, which is already common to every profile type. A StreamProfile
// additionally carries Stream, describing the demuxed stream itself rather
// than its codec configuration.
type ProfileMeta struct {
	UUID    uuid.UUID       `json:"id,omitzero"`
	Type    CodecType       `json:"type"`
	Codec   *CodecMeta      `json:"codec,omitempty"`
	Options json.RawMessage `json:"options,omitempty"`

	*ProfileMetaAudio
	*ProfileMetaVideo
	*ProfileMetaStream
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewProfileMeta builds a JSON representation of any Profile implementation.
func NewProfileMeta(p Profile) *ProfileMeta {
	if p == nil {
		return nil
	}

	meta := &ProfileMeta{
		UUID:    p.UUID(),
		Type:    p.Type(),
		Options: p.Options(),
	}
	if codec := p.Codec(); codec != nil {
		meta.Codec = &codec.CodecMeta
	}

	switch v := p.(type) {
	case *AudioProfile:
		meta.ProfileMetaAudio = &ProfileMetaAudio{
			Bitrate:       v.Bitrate,
			Profile:       v.Profile,
			SampleRate:    v.SampleRate,
			SampleFormat:  v.SampleFormat,
			ChannelLayout: v.ChannelLayout,
		}
	case *VideoProfile:
		meta.ProfileMetaVideo = &ProfileMetaVideo{
			Bitrate:     v.Bitrate,
			Profile:     v.Profile,
			Width:       v.Width,
			Height:      v.Height,
			PixelFormat: v.PixelFormat,
			FrameRate:   v.FrameRate,
		}
	case StreamProfile:
		populateProfileMetaStream(meta, &v)
	case *StreamProfile:
		populateProfileMetaStream(meta, v)
	}

	return meta
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// populateProfileMetaStream fills in meta's Stream block, plus its
// Audio/Video block derived from the stream's demuxed codec parameters (a
// StreamProfile has no native Bitrate/SampleRate/Width/... fields of its
// own, unlike Audio/VideoProfile).
func populateProfileMetaStream(meta *ProfileMeta, v *StreamProfile) {
	meta.ProfileMetaStream = &ProfileMetaStream{
		Index:       v.index,
		Disposition: v.disposition,
		TimeBase:    v.TimeBase(),
		Metadata:    newMetadataMetaList(v.metadata),
	}
	// v.codec (and so meta.Codec, set above from Codec()) is nil whenever
	// this build has no decoder registered for the stream's codec - common
	// for e.g. DVB teletext/EPG streams. The ID itself, and its name and
	// description, are still known from ffmpeg's static codec descriptor
	// table regardless, so fall back to that rather than dropping codec
	// identification from the output entirely.
	if meta.Codec == nil {
		if id := v.par.CodecID(); id != ff.AV_CODEC_ID_NONE {
			meta.Codec = &CodecMeta{
				Name:        id.String(),
				Description: id.LongName(),
				Type:        v.Type(),
			}
		}
	}
	switch v.Type() {
	case CodecType(ff.AVMEDIA_TYPE_AUDIO):
		meta.ProfileMetaAudio = profileMetaAudioFromPar(v.Par(), v.codec)
	case CodecType(ff.AVMEDIA_TYPE_VIDEO):
		meta.ProfileMetaVideo = profileMetaVideoFromPar(v.Par(), v.codec)
	}
}

func newMetadataMetaList(entries []gomedia.Metadata) []MetadataMeta {
	if len(entries) == 0 {
		return nil
	}
	result := make([]MetadataMeta, 0, len(entries))
	for _, e := range entries {
		if e == nil {
			continue
		}
		result = append(result, MetadataMeta{Key: e.Key(), Value: e.Value()})
	}
	return result
}

// profileNameForID looks up a codec's declared profile name by numeric ID -
// the reverse of resolveProfileID (audio.go/video.go) - or nil if the codec
// declares no such profile.
func profileNameForID(codec *ff.AVCodec, id int) *string {
	if codec == nil || id == ff.AV_PROFILE_UNKNOWN {
		return nil
	}
	for _, p := range codec.Profiles() {
		if p.ID() == id {
			name := p.Name()
			return &name
		}
	}
	return nil
}

func profileMetaAudioFromPar(par *ff.AVCodecParameters, codec *ff.AVCodec) *ProfileMetaAudio {
	if par == nil {
		return nil
	}
	meta := &ProfileMetaAudio{
		Profile: profileNameForID(codec, par.Profile()),
	}
	if v := par.BitRate(); v > 0 {
		meta.Bitrate = types.Ptr(uint64(v))
	}
	if v := par.SampleRate(); v > 0 {
		meta.SampleRate = types.Ptr(uint64(v))
	}
	if v := par.SampleFormat(); v != ff.AV_SAMPLE_FMT_NONE {
		s := v.String()
		meta.SampleFormat = &s
	}
	if ch := par.ChannelLayout(); ch.NumChannels() > 0 {
		if desc, err := ff.AVUtil_channel_layout_describe(&ch); err == nil {
			meta.ChannelLayout = &desc
		}
	}
	return meta
}

func profileMetaVideoFromPar(par *ff.AVCodecParameters, codec *ff.AVCodec) *ProfileMetaVideo {
	if par == nil {
		return nil
	}
	meta := &ProfileMetaVideo{
		Profile: profileNameForID(codec, par.Profile()),
	}
	if v := par.BitRate(); v > 0 {
		meta.Bitrate = types.Ptr(uint64(v))
	}
	if v := par.Width(); v > 0 {
		meta.Width = types.Ptr(uint64(v))
	}
	if v := par.Height(); v > 0 {
		meta.Height = types.Ptr(uint64(v))
	}
	if v := par.PixelFormat(); v != ff.AV_PIX_FMT_NONE {
		s := v.String()
		meta.PixelFormat = &s
	}
	// FrameRate is deliberately left nil: AVCodecParameters has no frame
	// rate field of its own (it's a stream-level property), and a demuxed
	// stream's timebase is just the container's native clock granularity,
	// not reliably 1/frame-rate the way a VideoProfile being configured
	// for encoding deliberately sets it (see VideoProfileMeta.setPar).
	return meta
}

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	OptionBitrate = "bitrate"

	// Audio options
	OptionProfile       = "profile"
	OptionSampleRate    = "sample_rate"
	OptionSampleFormat  = "sample_format"
	OptionChannelLayout = "channel_layout"

	// Video options
	OptionWidth       = "width"
	OptionHeight      = "height"
	OptionPixelFormat = "pixel_format"
	OptionFrameRate   = "frame_rate"
)
