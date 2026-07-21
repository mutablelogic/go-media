package schema

import (
	// Packages
	"net/url"
	"strconv"
	"strings"

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
