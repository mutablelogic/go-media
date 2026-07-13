package schema

import (
	"strings"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type AudioProfile struct {
	Codec        string      `json:"codec"`                   // "aac", "libmp3lame", "copy", "none"
	Bitrate      int         `json:"bitrate,omitempty"`       // bps; 0 = use quality
	SampleRate   int         `json:"sample_rate,omitempty"`   // Hz; 0 = passthrough
	SampleFormat string      `json:"sample_format,omitempty"` // Audio sample format; "fltp", "s16", leave empty for passthrough
	Channels     string      `json:"channels,omitempty"`      // Audio Channel Layout; "mono", "stereo", leave empty for passthrough
	Opts         []string    `json:"options,omitempty"`       // Additional codec options
	ctx          *ff.AVCodec `json:"-"`                       // Internal codec context
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewAudioProfile(codec string) (*AudioProfile, error) {
	self := new(AudioProfile)

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
		self.SampleRate = samplerates[0]
	}
	if sampleformats := encoder.SampleFormats(); len(sampleformats) > 0 {
		if sampleformat := strings.TrimSpace(sampleformats[0].String()); sampleformat != "" {
			self.SampleFormat = sampleformat
		}
	}
	if layouts := encoder.ChannelLayouts(); len(layouts) > 0 {
		if layout := layouts[0]; layout.NumChannels() > 0 {
			if desc, err := ff.AVUtil_channel_layout_describe(&layout); err == nil {
				self.Channels = strings.TrimSpace(desc)
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
					self.Bitrate = int(def)
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
// PUBLIC METHODS

// Set an audio profile option. If value is nil, the option is removed.
func (r *AudioProfile) SetOption(name string, value any) error {
	// Implementation goes here
	return nil
}

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
