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
	Name           string `json:"name"`
	BytesPerSample int    `json:"bytes_per_sample"` // Bytes per sample
	BitsPerSample  int    `json:"bits_per_sample"`  // Bits per sample
	IsPlanar       bool   `json:"is_planar"`        // Is planar format
	PackedName     string `json:"packed_name,omitempty"`
	PlanarName     string `json:"planar_name,omitempty"`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewSampleFormat(samplefmt ff.AVSampleFormat) *SampleFormat {
	if samplefmt == ff.AV_SAMPLE_FMT_NONE {
		return nil
	}
	name := ff.AVUtil_get_sample_fmt_name(samplefmt)
	if name == "" {
		return nil
	}
	bytesPerSample := ff.AVUtil_get_bytes_per_sample(samplefmt)
	packedFmt := ff.AVUtil_get_packed_sample_fmt(samplefmt)
	planarFmt := ff.AVUtil_get_planar_sample_fmt(samplefmt)

	return &SampleFormat{
		Name:           name,
		BytesPerSample: bytesPerSample,
		BitsPerSample:  bytesPerSample * 8,
		IsPlanar:       ff.AVUtil_sample_fmt_is_planar(samplefmt),
		PackedName:     ff.AVUtil_get_sample_fmt_name(packedFmt),
		PlanarName:     ff.AVUtil_get_sample_fmt_name(planarFmt),
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r SampleFormat) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
