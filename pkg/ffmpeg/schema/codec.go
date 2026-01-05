package schema

import (
	"encoding/json"
	"strings"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ListCodecRequest struct {
	Name      string `json:"name,omitempty"`       // Filter by codec name
	Type      string `json:"type,omitempty"`       // Filter by media type: "video", "audio", "subtitle", "data"
	IsEncoder *bool  `json:"is_encoder,omitempty"` // Filter by encoder (true) or decoder (false), nil = no filter
}

type ListCodecResponse []Codec

type Codec struct {
	Name         string   `json:"name"`
	LongName     string   `json:"long_name,omitempty"`
	Type         string   `json:"type"`                      // "video", "audio", "subtitle", "data", "attachment", "unknown"
	ID           string   `json:"id"`                        // Codec ID name
	IsEncoder    bool     `json:"is_encoder"`                // Is this an encoder
	IsDecoder    bool     `json:"is_decoder"`                // Is this a decoder
	IsHardware   bool     `json:"is_hardware,omitempty"`     // Is hardware accelerated
	IsExperiment bool     `json:"is_experimental,omitempty"` // Is experimental
	Capabilities []string `json:"capabilities,omitempty"`    // Capability flags

	// Supported formats (only populated for relevant codec types)
	PixelFormats   []string `json:"pixel_formats,omitempty"`   // Supported pixel formats (video)
	SampleFormats  []string `json:"sample_formats,omitempty"`  // Supported sample formats (audio)
	SampleRates    []int    `json:"sample_rates,omitempty"`    // Supported sample rates (audio)
	ChannelLayouts []string `json:"channel_layouts,omitempty"` // Supported channel layouts (audio)
	Profiles       []string `json:"profiles,omitempty"`        // Supported profiles
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewCodec(codec *ff.AVCodec) *Codec {
	if codec == nil {
		return nil
	}
	name := codec.Name()
	if name == "" {
		return nil
	}

	// Determine media type string
	typeStr := mediaTypeString(codec.Type())

	// Get capabilities as slice
	caps := codec.Capabilities()
	var capabilities []string
	if capsStr := caps.String(); capsStr != "" && capsStr != "AV_CODEC_CAP_NONE" {
		capabilities = strings.Split(capsStr, "|")
	}

	// Get pixel formats
	var pixelFormats []string
	for _, pf := range codec.PixelFormats() {
		if name := ff.AVUtil_get_pix_fmt_name(pf); name != "" {
			pixelFormats = append(pixelFormats, name)
		}
	}

	// Get sample formats
	var sampleFormats []string
	for _, sf := range codec.SampleFormats() {
		if name := ff.AVUtil_get_sample_fmt_name(sf); name != "" {
			sampleFormats = append(sampleFormats, name)
		}
	}

	// Get channel layouts
	var channelLayouts []string
	for _, cl := range codec.ChannelLayouts() {
		if name, _ := ff.AVUtil_channel_layout_describe(&cl); name != "" {
			channelLayouts = append(channelLayouts, name)
		}
	}

	// Get profiles
	var profiles []string
	for _, p := range codec.Profiles() {
		if name := p.Name(); name != "" {
			profiles = append(profiles, name)
		}
	}

	return &Codec{
		Name:           name,
		LongName:       codec.LongName(),
		Type:           typeStr,
		ID:             codec.ID().Name(),
		IsEncoder:      ff.AVCodec_is_encoder(codec),
		IsDecoder:      ff.AVCodec_is_decoder(codec),
		IsHardware:     caps&ff.AV_CODEC_CAP_HARDWARE != 0,
		IsExperiment:   caps&ff.AV_CODEC_CAP_EXPERIMENTAL != 0,
		Capabilities:   capabilities,
		PixelFormats:   pixelFormats,
		SampleFormats:  sampleFormats,
		SampleRates:    codec.SupportedSamplerates(),
		ChannelLayouts: channelLayouts,
		Profiles:       profiles,
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func mediaTypeString(t ff.AVMediaType) string {
	switch t {
	case ff.AVMEDIA_TYPE_VIDEO:
		return "video"
	case ff.AVMEDIA_TYPE_AUDIO:
		return "audio"
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
// STRINGIFY

func (r Codec) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
