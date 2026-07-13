package schema

import (
	"encoding/json"
	"strconv"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ListSampleFormatRequest struct {
	Name     string `json:"name"`
	IsPlanar *bool  `json:"is_planar,omitempty"`
}

type ListSampleFormatResponse []SampleFormat

type SampleFormat struct {
	ff.AVSampleFormat
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewSampleFormat(samplefmt ff.AVSampleFormat) *SampleFormat {
	if samplefmt == ff.AV_SAMPLE_FMT_NONE {
		return nil
	}
	return &SampleFormat{AVSampleFormat: samplefmt}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r SampleFormat) MarshalJSON() ([]byte, error) {
	return r.AVSampleFormat.MarshalJSON()
}

func (r SampleFormat) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (r ListSampleFormatResponse) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// TABLE WRITER

func (SampleFormat) Header() []string {
	return []string{"Name", "Planar", "Bytes", "Packed"}
}

func (r SampleFormat) Cell(col int) string {
	switch col {
	case 0:
		return r.Name()
	case 1:
		return strconv.FormatBool(r.IsPlanar())
	case 2:
		return strconv.Itoa(r.BytesPerSample())
	case 3:
		return r.PackedName()
	default:
		return ""
	}
}

func (SampleFormat) Width(col int) int {
	switch col {
	case 0:
		return 20
	case 1:
		return 8
	case 2:
		return 8
	default:
		return 0
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r SampleFormat) Name() string {
	return ff.AVUtil_get_sample_fmt_name(r.AVSampleFormat)
}

func (r SampleFormat) IsPlanar() bool {
	return ff.AVUtil_sample_fmt_is_planar(r.AVSampleFormat)
}

func (r SampleFormat) BytesPerSample() int {
	return ff.AVUtil_get_bytes_per_sample(r.AVSampleFormat)
}

func (r SampleFormat) PackedName() string {
	if !r.IsPlanar() {
		return r.Name()
	}
	if packed := ff.AVUtil_get_packed_sample_fmt(r.AVSampleFormat); packed != ff.AV_SAMPLE_FMT_NONE {
		return ff.AVUtil_get_sample_fmt_name(packed)
	}
	return ""
}
