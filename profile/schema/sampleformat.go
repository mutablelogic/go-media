package schema

import (
	"net/url"
	"strconv"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type SampleFormatListRequest struct {
	Name     *string `json:"name"`
	IsPlanar *bool   `json:"is_planar,omitempty"`
}

type SampleFormatList []SampleFormat

type SampleFormat struct {
	Name      string `json:"name" help:"Sample format name." example:"fltp"`
	ByteDepth int    `json:"byte_depth" help:"Bytes per sample." example:"4"`
	IsPlanar  bool   `json:"is_planar" help:"Whether planes are stored separately rather than interleaved." example:"true"`

	// Private context for the sample format
	ctx ff.AVSampleFormat
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewSampleFormat(samplefmt ff.AVSampleFormat) *SampleFormat {
	if samplefmt == ff.AV_SAMPLE_FMT_NONE {
		return nil
	}
	return &SampleFormat{
		Name:      ff.AVUtil_get_sample_fmt_name(samplefmt),
		ByteDepth: ff.AVUtil_get_bytes_per_sample(samplefmt),
		IsPlanar:  ff.AVUtil_sample_fmt_is_planar(samplefmt),
		ctx:       samplefmt,
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

func (r SampleFormatListRequest) Query() url.Values {
	query := url.Values{}
	if r.Name != nil {
		query.Set("name", types.Value(r.Name))
	}
	if r.IsPlanar != nil {
		query.Set("is_planar", strconv.FormatBool(types.Value(r.IsPlanar)))
	}
	return query
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r SampleFormat) String() string {
	return types.Stringify(r)
}

func (r SampleFormatList) String() string {
	return types.Stringify(r)
}

////////////////////////////////////////////////////////////////////////////////
// TABLE WRITER

func (SampleFormat) Header() []string {
	return []string{"Name", "Planar", "Bytes"}
}

func (r SampleFormat) Cell(col int) string {
	switch col {
	case 0:
		return r.Name
	case 1:
		return strconv.FormatBool(r.IsPlanar)
	case 2:
		return strconv.Itoa(r.ByteDepth)
	default:
		return ""
	}
}

func (SampleFormat) Width(col int) int {
	switch col {
	case 0:
		return 20
	case 1:
		return 5
	case 2:
		return 5
	default:
		return 0
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Context returns the underlying FFmpeg sample format constant.
func (r SampleFormat) Context() ff.AVSampleFormat {
	return r.ctx
}
