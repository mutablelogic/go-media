package schema

import (
	"encoding/json"
	"strings"

	// Packages
	uuid "github.com/google/uuid"
	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	pg "github.com/mutablelogic/go-pg"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type AudioProfileMeta struct {
	Name          string          `json:"codec"  arg:"" required:""` // "aac", "libmp3lame", "copy", ...
	Bitrate       *uint64         `json:"bitrate,omitempty"`         // bps
	SampleRate    *uint64         `json:"sample_rate,omitempty"`     // Hz
	SampleFormat  *string         `json:"sample_format,omitempty"`   // Audio sample format; "fltp", "s16"
	ChannelLayout *string         `json:"channel_layout,omitempty"`  // Audio Channel Layout; "mono", "stereo"
	Opts          json.RawMessage `json:"options,omitempty"`         // Additional codec options

	// Unexported fields
	codec    *ff.AVCodec          `json:"-"` // Internal codec
	par      ff.AVCodecParameters `json:"-"` // Internal parameters
	timebase ff.AVRational        `json:"-"` // Internal timebase
	opts     map[string]Option    `json:"-"` // Internal codec options
}

type AudioProfile struct {
	Id uuid.UUID `json:"id,omitempty"` // Unique identifier for the audio profile
	AudioProfileMeta
}

type AudioProfileUUID uuid.UUID

var _ Profile = (*AudioProfile)(nil)

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	OptionBitrate       = "bitrate"
	OptionSampleRate    = "sample_rate"
	OptionSampleFormat  = "sample_format"
	OptionChannelLayout = "channel_layout"
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewAudioProfile(codec string) (*AudioProfile, error) {
	// Create a new audio profile with default values
	encoder := ff.AVCodec_find_encoder_by_name(codec)
	if encoder == nil {
		return nil, gomedia.ErrBadParameter.Withf("codec %q is not found", codec)
	} else if encoder.Type() != ff.AVMEDIA_TYPE_AUDIO || encoder.IsEncoder() == false {
		return nil, gomedia.ErrBadParameter.Withf("codec %q is not an audio encoding codec", codec)
	}

	self := &AudioProfile{
		AudioProfileMeta: AudioProfileMeta{
			Name:  encoder.Name(),
			codec: encoder,
			opts:  optionsForCodec(encoder),
		},
	}

	// Update internal codec parameters and timebase
	if err := self.setPar(); err != nil {
		return nil, err
	}

	// Return success
	return self, nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r AudioProfile) String() string {
	return types.Stringify(r)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PROFILE INTERFACE

func (r AudioProfile) UUID() uuid.UUID {
	return r.Id
}

func (r AudioProfile) Type() CodecType {
	if r.codec == nil {
		return CodecType(ff.AVMEDIA_TYPE_UNKNOWN)
	}
	return CodecType(r.codec.Type())
}

func (r AudioProfile) Codec() *Codec {
	if r.codec == nil {
		return nil
	}
	return NewCodec(r.codec)
}

func (r AudioProfile) Par() *ff.AVCodecParameters {
	return types.Ptr(r.par)
}

func (r AudioProfile) TimeBase() *ff.AVRational {
	if r.timebase.Num() == 0 {
		return nil
	}
	return types.Ptr(r.timebase)
}

func (r AudioProfile) Options() json.RawMessage {
	return r.Opts
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - READER

// Expected column order: id, codec, bitrate, sample_rate, sample_format, channel_layout, opts.
func (r *AudioProfile) Scan(row pg.Row) error {
	if err := row.Scan(&r.Id, &r.Name, &r.Bitrate, &r.SampleRate, &r.SampleFormat, &r.ChannelLayout, &r.Opts); err != nil {
		return err
	}

	// Set context and options
	encoder := ff.AVCodec_find_encoder_by_name(r.Name)
	if encoder == nil {
		return gomedia.ErrBadParameter.Withf("codec %q is not found", r.Name)
	} else if encoder.Type() != ff.AVMEDIA_TYPE_AUDIO || encoder.IsEncoder() == false {
		return gomedia.ErrBadParameter.Withf("codec %q is not an audio encoding codec", r.Name)
	} else {
		r.codec = encoder
		r.opts = optionsForCodec(encoder)
	}

	// Set codec parameters and timebase
	if err := r.setPar(); err != nil {
		return err
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - SELECTOR

func (r AudioProfileUUID) Select(bind *pg.Bind, op pg.Op) (string, error) {
	bind.Set("id", uuid.UUID(r))

	switch op {
	case pg.Get:
		return bind.Query("profile.audio_get"), nil
	case pg.Delete:
		return bind.Query("profile.audio_delete"), nil
	case pg.Update:
		return bind.Query("profile.audio_update"), nil
	default:
		return "", gomedia.ErrInternalError.Withf("unsupported AudioProfileUUID operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - WRITER

// Insert binds values and returns the insert query for an audio profile row.
func (r AudioProfileMeta) Insert(bind *pg.Bind) (string, error) {
	bind.Set("codec", r.Name)
	bind.Set(OptionBitrate, r.Bitrate)
	bind.Set(OptionSampleRate, r.SampleRate)
	bind.Set(OptionSampleFormat, r.SampleFormat)
	bind.Set(OptionChannelLayout, r.ChannelLayout)
	if r.Opts == nil {
		bind.Set("opts", map[string]any{})
	} else {
		bind.Set("opts", r.Opts)
	}
	return bind.Query("profile.audio_insert"), nil
}

// Update binds patch values for an audio profile row update.
func (r AudioProfileMeta) Update(bind *pg.Bind) error {
	bind.Del("patch")

	if bitrate := types.Value(r.Bitrate); bitrate > 0 {
		bind.Append("patch", `"`+OptionBitrate+`" = `+bind.Set(OptionBitrate, bitrate))
	}
	if sampleRate := types.Value(r.SampleRate); sampleRate > 0 {
		bind.Append("patch", `"`+OptionSampleRate+`" = `+bind.Set(OptionSampleRate, sampleRate))
	}
	if value := strings.TrimSpace(types.Value(r.SampleFormat)); value != "" {
		bind.Append("patch", `"`+OptionSampleFormat+`" = `+bind.Set(OptionSampleFormat, value))
	}
	if value := strings.TrimSpace(types.Value(r.ChannelLayout)); value != "" {
		bind.Append("patch", `"`+OptionChannelLayout+`" = `+bind.Set(OptionChannelLayout, value))
	}
	if r.Opts != nil {
		bind.Append("patch", `"opts" = `+bind.Set("opts", r.Opts))
	}
	if patch := bind.Join("patch", ", "); patch == "" {
		return gomedia.ErrBadParameter.With("no fields to update")
	} else {
		bind.Set("patch", patch)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - GET/SET OPTIONS

// Set an audio profile option. If value is nil, the option is removed.
// TODO: If an error is returned, the option is not set and the profile is unchanged.
func (r *AudioProfileMeta) Set(name string, value any) error {
	// Check for existing option
	opt, exists := r.opts[name]
	if !exists {
		return gomedia.ErrBadParameter.Withf("option %q is not supported by codec %q", name, r.Name)
	}

	// Unmarshal the options JSON into a map
	var opts map[string]any
	if r.Opts == nil {
		opts = make(map[string]any)
	} else if err := json.Unmarshal(r.Opts, &opts); err != nil {
		return err
	}

	// Remove existing option
	if value == nil {
		switch name {
		case OptionBitrate:
			r.Bitrate = nil
		case OptionSampleRate:
			r.SampleRate = nil
		case OptionSampleFormat:
			r.SampleFormat = nil
		case OptionChannelLayout:
			r.ChannelLayout = nil
		default:
			delete(opts, name)
		}
	} else if value, err := opt.Validate(value); err != nil {
		return err
	} else {
		// Set the option value
		switch name {
		case OptionBitrate:
			r.Bitrate = types.Ptr(value.(uint64))
		case OptionSampleRate:
			r.SampleRate = types.Ptr(value.(uint64))
		case OptionSampleFormat:
			r.SampleFormat = types.Ptr(value.(string))
		case OptionChannelLayout:
			r.ChannelLayout = types.Ptr(value.(string))
		default:
			opts[name] = value
		}
	}

	// Set the option value in the options map
	if optsJSON, err := json.Marshal(opts); err != nil {
		return err
	} else {
		r.Opts = optsJSON
	}

	// Set the internal codec parameters and timebase
	if err := r.setPar(); err != nil {
		return err
	}

	// Return success
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - GET/SET OPTIONS

func (r *AudioProfileMeta) setPar() error {
	// Check for codec
	if r.codec == nil {
		return gomedia.ErrInternalError.With("codec is not set")
	} else {
		r.par.SetCodecType(r.codec.Type())
		r.par.SetCodecID(r.codec.ID())
		r.par.SetProfile(ff.AV_PROFILE_UNKNOWN)
	}

	// Sample Format
	if samplefmt := types.Value(r.SampleFormat); samplefmt != "" {
		if samplefmt_ := ff.AVUtil_get_sample_fmt(samplefmt); samplefmt_ == ff.AV_SAMPLE_FMT_NONE {
			return gomedia.ErrBadParameter.Withf("unknown sample format %q", samplefmt)
		} else {
			r.par.SetSampleFormat(samplefmt_)
		}
	}

	// Channel layout
	var ch ff.AVChannelLayout
	if channellayout := types.Value(r.ChannelLayout); channellayout != "" {
		if err := ff.AVUtil_channel_layout_from_string(&ch, channellayout); err != nil {
			return gomedia.ErrBadParameter.Withf("invalid channel layout %q: %w", channellayout, err)
		}
		if err := r.par.SetChannelLayout(ch); err != nil {
			return err
		}
	}

	// Sample rate
	if samplerate := types.Value(r.SampleRate); samplerate > 0 {
		r.par.SetSampleRate(int(samplerate))
		r.timebase = ff.AVUtil_rational(1, int(samplerate))
	}

	// Return success
	return nil
}
