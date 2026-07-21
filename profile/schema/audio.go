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
	Codec         string            `json:"codec"  arg:"" required:""` // "aac", "libmp3lame", "copy", ...
	Bitrate       *uint64           `json:"bitrate,omitempty"`         // bps
	SampleRate    *uint64           `json:"sample_rate,omitempty"`     // Hz
	SampleFormat  *string           `json:"sample_format,omitempty"`   // Audio sample format; "fltp", "s16"
	ChannelLayout *string           `json:"channel_layout,omitempty"`  // Audio Channel Layout; "mono", "stereo"
	Opts          json.RawMessage   `json:"options,omitempty"`         // Additional codec options
	ctx           *ff.AVCodec       `json:"-"`                         // Internal codec
	opts          map[string]Option `json:"-"`                         // Internal codec options
}

type AudioProfile struct {
	Id uuid.UUID `json:"id,omitempty"` // Unique identifier for the audio profile
	AudioProfileMeta
}

type AudioProfileUUID uuid.UUID

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
			Codec: encoder.Name(),
			ctx:   encoder,
			opts:  optionsForCodec(encoder),
		},
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
// PUBLIC METHODS - READER

// Expected column order: id, codec, bitrate, sample_rate, sample_format, channel_layout, opts.
func (r *AudioProfile) Scan(row pg.Row) error {
	if err := row.Scan(&r.Id, &r.Codec, &r.Bitrate, &r.SampleRate, &r.SampleFormat, &r.ChannelLayout, &r.Opts); err != nil {
		return err
	}

	// Set context and options
	encoder := ff.AVCodec_find_encoder_by_name(r.Codec)
	if encoder == nil {
		return gomedia.ErrBadParameter.Withf("codec %q is not found", r.Codec)
	} else if encoder.Type() != ff.AVMEDIA_TYPE_AUDIO || encoder.IsEncoder() == false {
		return gomedia.ErrBadParameter.Withf("codec %q is not an audio encoding codec", r.Codec)
	} else {
		r.ctx = encoder
		r.opts = optionsForCodec(encoder)
	}
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
	default:
		return "", gomedia.ErrInternalError.Withf("unsupported AudioProfileUUID operation %q", op)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - WRITER

// Insert binds values and returns the insert query for an audio profile row.
func (r AudioProfileMeta) Insert(bind *pg.Bind) (string, error) {
	bind.Set("codec", r.Codec)
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
func (r *AudioProfileMeta) Set(name string, value any) error {
	// Check for existing option
	opt, exists := r.opts[name]
	if !exists {
		return gomedia.ErrBadParameter.Withf("option %q is not supported by codec %q", name, r.Codec)
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

	// Return success
	return nil
}
