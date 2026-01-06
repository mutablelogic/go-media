package schema

import (
	"encoding/json"

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
	*ff.AVCodec
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewCodec(codec *ff.AVCodec) *Codec {
	if codec == nil {
		return nil
	}
	return &Codec{AVCodec: codec}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r Codec) MarshalJSON() ([]byte, error) {
	if r.AVCodec == nil {
		return json.Marshal(nil)
	}
	return r.AVCodec.MarshalJSON()
}

func (r Codec) String() string {
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
