package schema

import (
	"encoding/json"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ListSampleFormatRequest struct {
	Name     string `json:"name"`
	IsPlanar *bool  `json:"is_planar,omitempty"` // Filter by planar/packed (nil = no filter)
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
