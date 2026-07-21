package schema

import (
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
	Codec        string      `json:"codec"`                    // "aac", "libmp3lame", "copy", ...
	Bitrate      *uint64     `json:"bitrate,omitempty"`        // bps; 0 = use quality
	SampleRate   *uint64     `json:"sample_rate,omitempty"`    // Hz; 0 = passthrough
	SampleFormat *string     `json:"sample_format,omitempty"`  // Audio sample format; "fltp", "s16", leave empty for passthrough
	Channels     *string     `json:"channel_layout,omitempty"` // Audio Channel Layout; "mono", "stereo", leave empty for passthrough
	Opts         []string    `json:"options,omitempty"`        // Additional codec options
	ctx          *ff.AVCodec `json:"-"`                        // Internal codec
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

func NewAudioProfile(codec string) (*AudioProfileMeta, error) {
	self := new(AudioProfileMeta)

	// Create a new audio profile with default values
	encoder := ff.AVCodec_find_encoder_by_name(codec)
	if encoder == nil {
		return nil, gomedia.ErrBadParameter.Withf("codec %q is not found", codec)
	} else if encoder.Type() != ff.AVMEDIA_TYPE_AUDIO || encoder.IsEncoder() == false {
		return nil, gomedia.ErrBadParameter.Withf("codec %q is not an audio encoding codec", codec)
	} else {
		self.Codec = encoder.Name()
		self.ctx = encoder
	}

	// Set default values for the audio profile (if default is nil, option is removed/ignored)
	for _, opt := range AudioOptionsForCodec(encoder) {
		if err := self.Set(opt.Name, opt.Default); err != nil {
			return nil, err
		}
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

// Expected column order: id, bitrate, sample_rate, sample_format, channels, opts.
func (r *AudioProfile) Scan(row pg.Row) error {
	return row.Scan(&r.Id, &r.Bitrate, &r.SampleRate, &r.SampleFormat, &r.Channels, &r.Opts)
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
	bind.Set(OptionBitrate, r.Bitrate)
	bind.Set(OptionSampleRate, r.SampleRate)
	bind.Set(OptionSampleFormat, r.SampleFormat)
	bind.Set(OptionChannelLayout, r.Channels)
	if r.Opts == nil {
		bind.Set("opts", []string{})
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
	if value := strings.TrimSpace(types.Value(r.Channels)); value != "" {
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
	// TODO
	return gomedia.ErrNotImplemented
}
