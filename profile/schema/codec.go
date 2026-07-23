package schema

import (
	"net/url"
	"strconv"
	"strings"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	pg "github.com/mutablelogic/go-pg"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type CodecType ff.AVMediaType

type Codec struct {
	Name        string      `json:"name"`                  // Codec name, e.g. "aac", "libmp3lame", "copy", ...
	Description string      `json:"description,omitempty"` // Codec description
	Type        CodecType   `json:"type"`                  // Codec type; "audio", "video", "subtitle"
	Opts        []Option    `json:"opts,omitempty"`        // Codec options
	ctx         *ff.AVCodec `json:"-"`                     // Internal codec
}

type CodecListRequest struct {
	Type *CodecType `json:"type,omitempty" enum:"audio,video,subtitle"` // Codec type to filter codecs by; "audio", "video", "subtitle"
	pg.OffsetLimit
}

type CodecList struct {
	CodecListRequest
	Count uint64   `json:"count"`          // Number of codecs
	Body  []*Codec `json:"body,omitempty"` // List of codecs
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewCodec(codec *ff.AVCodec) *Codec {
	if codec == nil {
		return nil
	}
	return &Codec{
		Name:        codec.Name(),
		Description: codec.LongName(),
		Type:        CodecType(codec.Type()),
		Opts:        OptionsForCodec(codec),
		ctx:         codec,
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r Codec) String() string {
	return types.Stringify(r)
}

func (r CodecList) String() string {
	return types.Stringify(r)
}

func (r CodecType) String() string {
	switch ff.AVMediaType(r) {
	case ff.AVMEDIA_TYPE_AUDIO:
		return "audio"
	case ff.AVMEDIA_TYPE_VIDEO:
		return "video"
	case ff.AVMEDIA_TYPE_SUBTITLE:
		return "subtitle"
	case ff.AVMEDIA_TYPE_DATA:
		return "data"
	case ff.AVMEDIA_TYPE_ATTACHMENT:
		return "attachment"
	default:
		return "unknown"
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r Codec) Context() *ff.AVCodec {
	return r.ctx
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

func (r CodecListRequest) Query() url.Values {
	query := url.Values{}
	if r.Type != nil {
		query.Set("type", strconv.FormatUint(uint64(types.Value(r.Type)), 10))
	}
	if r.Offset > 0 {
		query.Set("offset", strconv.FormatUint(r.Offset, 10))
	}
	if r.Limit != nil {
		query.Set("limit", strconv.FormatUint(types.Value(r.Limit), 10))
	}
	return query
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - MARSHALING

func (r *CodecType) UnmarshalText(data []byte) error {
	text := strings.ToLower(strings.TrimSpace(string(data)))
	if r == nil {
		return gomedia.ErrBadParameter.Withf("nil CodecType")
	}
	switch text {
	case "audio":
		*r = CodecType(ff.AVMEDIA_TYPE_AUDIO)
	case "video":
		*r = CodecType(ff.AVMEDIA_TYPE_VIDEO)
	case "subtitle":
		*r = CodecType(ff.AVMEDIA_TYPE_SUBTITLE)
	case "data":
		*r = CodecType(ff.AVMEDIA_TYPE_DATA)
	case "attachment":
		*r = CodecType(ff.AVMEDIA_TYPE_ATTACHMENT)
	default:
		*r = CodecType(ff.AVMEDIA_TYPE_UNKNOWN)
		return gomedia.ErrBadParameter.Withf("invalid codec type: %q", text)
	}
	return nil
}

func (r *CodecType) UnmarshalJSON(data []byte) error {
	text, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	return r.UnmarshalText([]byte(text))
}

func (r CodecType) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(r.String())), nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func AudioOptionsForCodec(codec *ff.AVCodec) []Option {
	if codec == nil {
		return nil
	}

	// Bitrate Option
	bitrate := Option{
		Name:        OptionBitrate,
		Description: "Audio bitrate in bits per second.",
		Type:        "int",
		Unit:        "bps",
	}

	// Sample Rate Option
	sample_rate := Option{
		Name:        OptionSampleRate,
		Description: "Sample rate in Hz.",
		Type:        "int",
		Unit:        "Hz",
		Min:         types.Ptr(0),
		Max:         types.Ptr(192000),
	}
	for i, rate := range codec.SupportedSamplerates() {
		if i == 0 {
			sample_rate.Default = uint64(rate)
		}
		sample_rate.Const = append(sample_rate.Const, Option{
			Default: uint64(rate),
			Type:    "int",
		})
	}

	// Sample Format Option
	sample_format := Option{
		Name:        OptionSampleFormat,
		Description: "Audio sample format.",
		Type:        "string",
	}
	for i, format := range codec.SampleFormats() {
		if i == 0 {
			sample_format.Default = strings.TrimSpace(format.String())
		}
		sample_format.Const = append(sample_format.Const, Option{
			Default: strings.TrimSpace(format.String()),
			Type:    "string",
		})
	}

	// Channel Layout Option
	channel_layout := Option{
		Name:        OptionChannelLayout,
		Description: "Audio channel layout.",
		Type:        "string",
	}
	for i, layout := range codec.ChannelLayouts() {
		if i == 0 {
			if desc, err := ff.AVUtil_channel_layout_describe(&layout); err == nil {
				channel_layout.Default = strings.TrimSpace(desc)
			}
		}
		if desc, err := ff.AVUtil_channel_layout_describe(&layout); err == nil {
			channel_layout.Const = append(channel_layout.Const, Option{
				Default: strings.TrimSpace(desc),
				Type:    "string",
			})
		}
	}

	result := []Option{bitrate}
	if len(codec.Profiles()) > 0 {
		// Only advertise "profile" for codecs that actually declare one —
		// many codecs (e.g. flac, pcm_*) have no concept of it at all.
		// Codecs that expose profile only as their own private option
		// (e.g. libx264/libx265) still get it via the private-class
		// extraction in OptionsForCodec, without needing it added here.
		result = append(result, profileOptionForCodec(codec))
	}
	return append(result, sample_rate, sample_format, channel_layout)
}

func VideoOptionsForCodec(codec *ff.AVCodec) []Option {
	if codec == nil {
		return nil
	}

	// Bitrate Option
	bitrate := Option{
		Name:        OptionBitrate,
		Description: "Video bitrate in bits per second.",
		Type:        "int",
		Unit:        "bps",
	}

	// Width Option
	width := Option{
		Name:        OptionWidth,
		Description: "Frame width in pixels.",
		Type:        "int",
		Unit:        "px",
	}

	// Height Option
	height := Option{
		Name:        OptionHeight,
		Description: "Frame height in pixels.",
		Type:        "int",
		Unit:        "px",
	}

	// Pixel Format Option
	pixel_format := Option{
		Name:        OptionPixelFormat,
		Description: "Video pixel format.",
		Type:        "string",
	}
	for i, format := range codec.PixelFormats() {
		name := strings.TrimSpace(ff.AVUtil_get_pix_fmt_name(format))
		if i == 0 {
			pixel_format.Default = name
		}
		pixel_format.Const = append(pixel_format.Const, Option{
			Default: name,
			Type:    "string",
		})
	}

	// Frame Rate Option
	frame_rate := Option{
		Name:        OptionFrameRate,
		Description: "Frame rate in frames per second.",
		Type:        "double",
		Unit:        "fps",
	}
	for i, rate := range codec.SupportedFramerates() {
		fps := ff.AVUtil_rational_q2d(rate)
		if i == 0 {
			frame_rate.Default = fps
		}
		frame_rate.Const = append(frame_rate.Const, Option{
			Default: fps,
			Type:    "double",
		})
	}

	result := []Option{bitrate}
	if len(codec.Profiles()) > 0 {
		// Only advertise "profile" for codecs that actually declare one —
		// many codecs (e.g. rawvideo, mpeg4) have no concept of it at all.
		// Codecs that expose profile only as their own private option
		// (e.g. libx264/libx265) still get it via the private-class
		// extraction in OptionsForCodec, without needing it added here.
		result = append(result, profileOptionForCodec(codec))
	}
	return append(result, width, height, pixel_format, frame_rate)
}

// profileOptionForCodec builds the shared "profile" Option (e.g. h264's
// baseline/main/high, or aac's LC/HE-AAC/HE-AACv2), listing every profile
// the codec advertises as a Const choice.
func profileOptionForCodec(codec *ff.AVCodec) Option {
	profile := Option{
		Name:        OptionProfile,
		Description: "Codec profile.",
		Type:        "string",
	}
	for i, p := range codec.Profiles() {
		name := strings.TrimSpace(p.Name())
		if name == "" {
			continue
		}
		if i == 0 {
			profile.Default = name
		}
		profile.Const = append(profile.Const, Option{
			Default: name,
			Type:    "string",
		})
	}
	return profile
}

// resolveProfileID looks up the AVProfile ID for name (case-insensitive). An
// empty name resolves to AV_PROFILE_UNKNOWN (the encoder picks its own
// default). Returns an error if name is set but not recognized by codec.
func resolveProfileID(codec *ff.AVCodec, name string) (int, error) {
	if name == "" {
		return ff.AV_PROFILE_UNKNOWN, nil
	}
	for _, p := range codec.Profiles() {
		if strings.EqualFold(strings.TrimSpace(p.Name()), name) {
			return p.ID(), nil
		}
	}
	return 0, gomedia.ErrBadParameter.Withf("unknown profile %q for codec %q", name, codec.Name())
}

func SubtitleOptionsForCodec(_ *ff.AVCodec) []Option {
	return nil
}

func OptionsForCodec(codec *ff.AVCodec) []Option {
	if codec == nil {
		return nil
	}

	// Codecs with no private class (e.g. raw pcm_* codecs, which have
	// nothing encoder-specific to configure) still get the universal
	// bitrate/sample_rate/sample_format/channel_layout options below —
	// only the codec-specific AVOption extraction is skipped for them.
	result := []Option{}
	if class := codec.PrivClass(); class != nil {
		ffopts := ff.AVUtil_opt_list_from_class(class)
		consts := make(map[string][]Option, len(ffopts))
		for _, opt := range ffopts {
			if opt == nil {
				continue
			}
			if opt.Type() == ff.AV_OPT_TYPE_CONST {
				key := opt.Unit()
				consts[key] = append(consts[key], NewOption(opt))
				continue
			}
			name := strings.TrimSpace(opt.Name())
			if name == "" {
				continue
			}
			result = append(result, NewOption(opt))
		}

		// Append the constants to the options
		for i, opt := range result {
			consts, exists := consts[opt.Unit]
			if exists && len(consts) > 0 {
				result[i].Const = consts
			}
		}
	}

	// Return the options - and prepend audio video, subtitle options if applicable
	switch codec.Type() {
	case ff.AVMEDIA_TYPE_AUDIO:
		return append(AudioOptionsForCodec(codec), result...)
	case ff.AVMEDIA_TYPE_VIDEO:
		return append(VideoOptionsForCodec(codec), result...)
	case ff.AVMEDIA_TYPE_SUBTITLE:
		return append(SubtitleOptionsForCodec(codec), result...)
	default:
		return result
	}
}

func optionsForCodec(codec *ff.AVCodec) map[string]Option {
	if codec == nil {
		return nil
	}

	// Extract options
	opts := OptionsForCodec(codec)
	result := make(map[string]Option, len(opts))
	for _, opt := range opts {
		result[opt.Name] = opt
	}

	// Return the options
	return result
}
