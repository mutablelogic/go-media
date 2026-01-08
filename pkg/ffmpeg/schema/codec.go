package schema

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ListCodecRequest struct {
	Name      string `json:"name,omitempty" help:"Filter by codec name (partial match)"`
	Type      string `json:"type,omitempty" help:"Filter by media type: video, audio, subtitle, data"`
	IsEncoder *bool  `json:"is_encoder,omitempty" help:"Filter by encoder (true) or decoder (false)"`
}

type ListCodecResponse []Codec

type Codec struct {
	*ff.AVCodec    `json:"-"`
	Type           string               `json:"type"`
	Name           string               `json:"name,omitempty"`
	LongName       string               `json:"long_name,omitempty"`
	ID             string               `json:"id,omitempty"`
	IsEncoder      bool                 `json:"is_encoder"`
	IsDecoder      bool                 `json:"is_decoder"`
	Capabilities   string               `json:"capabilities,omitempty"`
	Framerates     []ff.AVRational      `json:"supported_framerates,omitempty"`
	SampleFormats  []ff.AVSampleFormat  `json:"sample_formats,omitempty"`
	PixelFormats   []ff.AVPixelFormat   `json:"pixel_formats,omitempty"`
	Samplerates    []int                `json:"samplerates,omitempty"`
	Profiles       []string             `json:"profiles,omitempty"`
	ChannelLayouts []ff.AVChannelLayout `json:"channel_layouts,omitempty"`
	Opts           []*ff.AVOption       `json:"options,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewCodec(codec *ff.AVCodec) *Codec {
	if codec == nil {
		return nil
	}

	c := &Codec{
		AVCodec:       codec,
		Type:          mediaTypeString(codec.Type()),
		Name:          codec.Name(),
		LongName:      codec.LongName(),
		ID:            codec.ID().String(),
		IsEncoder:     codec.IsEncoder(),
		IsDecoder:     codec.IsDecoder(),
		Capabilities:  codec.Capabilities().String(),
		Framerates:    codec.SupportedFramerates(),
		SampleFormats: codec.SampleFormats(),
		PixelFormats:  codec.PixelFormats(),
		Samplerates:   codec.SupportedSamplerates(),
	}

	// Convert profiles to strings
	profiles := codec.Profiles()
	if len(profiles) > 0 {
		c.Profiles = make([]string, len(profiles))
		for i, p := range profiles {
			c.Profiles[i] = p.Name()
		}
	}

	// TODO: Convert channel layouts to structured format
	c.ChannelLayouts = codec.ChannelLayouts()

	// Get options directly from codec's priv_class using FAKE_OBJ trick
	if class := codec.PrivClass(); class != nil {
		c.Opts = ff.AVUtil_opt_list_from_class(class)
	}

	return c
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

func (r ListCodecResponse) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// QueryValues returns the URL query values for the request
func (r *ListCodecRequest) QueryValues() url.Values {
	values := url.Values{}
	if name := strings.TrimSpace(r.Name); name != "" {
		values.Set("name", name)
	}
	if typ := strings.TrimSpace(r.Type); typ != "" {
		values.Set("type", typ)
	}
	if r.IsEncoder != nil {
		values.Set("is_encoder", fmt.Sprint(*r.IsEncoder))
	}
	return values
}

// Type returns the media type as a string
func (c *Codec) MediaType() string {
	if c.AVCodec == nil {
		return "unknown"
	}
	return c.Type
}

////////////////////////////////////////////////////////////////////////////////
// HELPER FUNCTIONS

// mediaTypeString converts AVMediaType to string representation
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
