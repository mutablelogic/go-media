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
	Name        string    `json:"name"`                  // Codec name, e.g. "aac", "libmp3lame", "copy", ...
	Description string    `json:"description,omitempty"` // Codec description
	Type        CodecType `json:"type"`                  // Codec type; "audio", "video", "subtitle"
	Opts        []Option  `json:"opts,omitempty"`        // Codec options
}

type CodecListRequest struct {
	Type *CodecType `json:"type,omitempty" enum:"audio,video,subtitle"` // Codec type to filter codecs by; "audio", "video", "subtitle"
	pg.OffsetLimit
}

type CodecList struct {
	CodecListRequest
	Count uint64  `json:"count"` // Number of audio codecs
	Body  []Codec `json:"body"`  // List of audio codecs
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

	return []Option{bitrate, sample_rate, sample_format, channel_layout}
}

func VideoOptionsForCodec(_ *ff.AVCodec) []Option {
	// TODO
	return nil
}

func SubtitleOptionsForCodec(_ *ff.AVCodec) []Option {
	return nil
}

func OptionsForCodec(codec *ff.AVCodec) []Option {
	if codec == nil {
		return nil
	}

	// Prefer default bitrate from the codec private class if available.
	class := codec.PrivClass()
	if class == nil {
		return nil
	}

	// Extract options
	ffopts := ff.AVUtil_opt_list_from_class(class)
	consts := make(map[string][]Option, len(ffopts))
	result := make([]Option, 0, len(ffopts))
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
		consts, exists := consts[opt.Name]
		if exists && len(consts) > 0 {
			result[i].Const = consts
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
