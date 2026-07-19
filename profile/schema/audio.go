package schema

import (
	"net/url"
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

type AudioCodec struct {
	Name        string `json:"name"`                  // Codec name, e.g. "aac", "libmp3lame", "copy", ...
	Description string `json:"description,omitempty"` // Codec description
}

type AudioCodecList struct {
	Count uint64       `json:"count"` // Number of audio codecs
	Body  []AudioCodec `json:"body"`  // List of audio codecs
}

type AudioProfileMeta struct {
	Codec        string      `json:"codec"`                   // "aac", "libmp3lame", "copy", ...
	Bitrate      *uint64     `json:"bitrate,omitempty"`       // bps; 0 = use quality
	SampleRate   *uint64     `json:"sample_rate,omitempty"`   // Hz; 0 = passthrough
	SampleFormat *string     `json:"sample_format,omitempty"` // Audio sample format; "fltp", "s16", leave empty for passthrough
	Channels     *string     `json:"channels,omitempty"`      // Audio Channel Layout; "mono", "stereo", leave empty for passthrough
	Opts         []string    `json:"options,omitempty"`       // Additional codec options
	ctx          *ff.AVCodec `json:"-"`                       // Internal codec
}

type AudioProfile struct {
	Id uuid.UUID `json:"id,omitempty"` // Unique identifier for the audio profile
	AudioProfileMeta
}

type AudioProfileUUID uuid.UUID

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewAudioProfile(codec string) (*AudioProfileMeta, error) {
	self := new(AudioProfileMeta)

	// Create a new audio profile with default values
	encoder := ff.AVCodec_find_encoder_by_name(codec)
	if encoder == nil {
		return nil, gomedia.ErrBadParameter.Withf("codec %q is not found", codec)
	} else if encoder.Type() != ff.AVMEDIA_TYPE_AUDIO {
		return nil, gomedia.ErrBadParameter.Withf("codec %q is not an audio encoding codec", codec)
	} else {
		self.Codec = encoder.Name()
		self.ctx = encoder
	}

	// Fill defaults from codec capabilities where available.
	if samplerates := encoder.SupportedSamplerates(); len(samplerates) > 0 {
		self.SampleRate = types.Ptr(uint64(samplerates[0]))
	}
	if sampleformats := encoder.SampleFormats(); len(sampleformats) > 0 {
		if sampleformat := strings.TrimSpace(sampleformats[0].String()); sampleformat != "" {
			self.SampleFormat = types.Ptr(sampleformat)
		}
	}
	if layouts := encoder.ChannelLayouts(); len(layouts) > 0 {
		if layout := layouts[0]; layout.NumChannels() > 0 {
			if desc, err := ff.AVUtil_channel_layout_describe(&layout); err == nil {
				self.Channels = types.Ptr(strings.TrimSpace(desc))
			}
		}
	}

	// Prefer default bitrate from the codec private class if available.
	if class := encoder.PrivClass(); class != nil {
		// Extract any bitrate value
		for _, opt := range ff.AVUtil_opt_list_from_class(class) {
			if opt == nil || opt.Type() == ff.AV_OPT_TYPE_CONST {
				continue
			}
			name := strings.TrimSpace(opt.Name())
			if name == "" {
				continue
			}
			if name == "b" || name == "bitrate" {
				if def, ok := opt.DefaultVal().(int64); ok && def > 0 {
					self.Bitrate = types.Ptr(uint64(def))
				}
			}
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
	bind.Set("bitrate", r.Bitrate)
	bind.Set("sample_rate", r.SampleRate)
	bind.Set("sample_format", r.SampleFormat)
	bind.Set("channels", r.Channels)
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
		bind.Append("patch", `"bitrate" = `+bind.Set("bitrate", bitrate))
	}
	if sampleRate := types.Value(r.SampleRate); sampleRate > 0 {
		bind.Append("patch", `"sample_rate" = `+bind.Set("sample_rate", sampleRate))
	}
	if value := strings.TrimSpace(types.Value(r.SampleFormat)); value != "" {
		bind.Append("patch", `"sample_format" = `+bind.Set("sample_format", value))
	}
	if value := strings.TrimSpace(types.Value(r.Channels)); value != "" {
		bind.Append("patch", `"channels" = `+bind.Set("channels", value))
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
func (r *AudioProfileMeta) Set(opts url.Values) error {
	// Implementation goes here
	return nil
}

// Return all options for the audio profile, including codec private options.
func (r AudioProfile) Options() []Option {
	opts := make([]Option, 0, 10)
	byUnit := make(map[string]int)
	constsByUnit := make(map[string][]*ff.AVOption)

	if r.ctx == nil {
		return opts
	}

	// Enumerate options from the codec private class if available.
	if class := r.ctx.PrivClass(); class != nil {
		for _, opt := range ff.AVUtil_opt_list_from_class(class) {
			if opt == nil {
				continue
			}
			if opt.Type() == ff.AV_OPT_TYPE_CONST {
				if unit := strings.TrimSpace(opt.Unit()); unit != "" {
					constsByUnit[unit] = append(constsByUnit[unit], opt)
				}
				continue
			}

			option := NewOption(opt)
			if option.Name == "" {
				continue
			}
			if option.Unit != "" {
				byUnit[option.Unit] = len(opts)
			}
			opts = append(opts, option)
		}
	}

	for unit, consts := range constsByUnit {
		index, exists := byUnit[unit]
		if !exists {
			continue
		}

		for _, opt := range consts {
			if name := strings.TrimSpace(opt.Name()); name != "" {
				opts[index].Enum = append(opts[index].Enum, name)
			}
		}

		if current, ok := opts[index].Value.(int64); ok {
			for _, opt := range consts {
				if value, ok := opt.DefaultVal().(int64); ok && value == current {
					if name := strings.TrimSpace(opt.Name()); name != "" {
						opts[index].Value = name
						break
					}
				}
			}
		}
	}
	return opts
}
