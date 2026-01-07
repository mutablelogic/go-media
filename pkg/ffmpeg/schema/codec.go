package schema

import (
	"encoding/json"

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
	*ff.AVCodec
	Opts []*ff.AVOption `json:"options,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewCodec(codec *ff.AVCodec) *Codec {
	if codec == nil {
		return nil
	}
	c := &Codec{AVCodec: codec}

	// Get options directly from codec's priv_class using FAKE_OBJ trick
	if class := codec.PrivClass(); class != nil {
		c.Opts = ff.AVUtil_opt_list_from_class(class)
	}

	return c
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r Codec) MarshalJSON() ([]byte, error) {
	if r.AVCodec == nil {
		return json.Marshal(nil)
	}

	// Get base codec JSON
	codecJSON, err := r.AVCodec.MarshalJSON()
	if err != nil {
		return nil, err
	}

	// If no options, return base JSON
	if len(r.Opts) == 0 {
		return codecJSON, nil
	}

	// Unmarshal to map and add options
	var result map[string]interface{}
	if err := json.Unmarshal(codecJSON, &result); err != nil {
		return nil, err
	}
	result["options"] = r.Opts

	return json.Marshal(result)
}

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

// Type returns a MediaType wrapper that provides proper string formatting
func (c *Codec) Type() *MediaType {
	if c.AVCodec == nil {
		return nil
	}
	mt := MediaType(c.AVCodec.Type())
	return &mt
}

////////////////////////////////////////////////////////////////////////////////
// HELPER TYPES

type MediaType ff.AVMediaType

func (mt MediaType) String() string {
	return mediaTypeString(ff.AVMediaType(mt))
}

func (mt MediaType) MarshalJSON() ([]byte, error) {
	return json.Marshal(mt.String())
}

// Equals checks if the MediaType equals the given AVMediaType
func (mt *MediaType) Equals(t ff.AVMediaType) bool {
	if mt == nil {
		return false
	}
	return ff.AVMediaType(*mt) == t
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
