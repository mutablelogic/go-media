package ffmpeg

import (
	"encoding/json"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type SampleFormat ff.AVSampleFormat

type jsonSampleFormat struct {
	Name           string `json:"name"`
	IsPlanar       bool   `json:"is_planar"`
	BytesPerSample int    `json:"bytes_per_sample"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newSampleFormat(samplefmt ff.AVSampleFormat) *SampleFormat {
	if samplefmt == ff.AV_SAMPLE_FMT_NONE {
		return nil
	} else if name := ff.AVUtil_get_sample_fmt_name(samplefmt); name == "" {
		return nil
	} else {
		return (*SampleFormat)(&samplefmt)
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (samplefmt *SampleFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonSampleFormat{
		Name:           samplefmt.Name(),
		IsPlanar:       samplefmt.IsPlanar(),
		BytesPerSample: samplefmt.BytesPerSample(),
	})
}

func (samplefmt *SampleFormat) String() string {
	data, _ := json.MarshalIndent(samplefmt, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (samplefmt *SampleFormat) Name() string {
	return ff.AVUtil_get_sample_fmt_name(ff.AVSampleFormat(*samplefmt))
}

func (samplefmt *SampleFormat) IsPlanar() bool {
	return ff.AVUtil_sample_fmt_is_planar(ff.AVSampleFormat(*samplefmt))
}

func (samplefmt *SampleFormat) BytesPerSample() int {
	return ff.AVUtil_get_bytes_per_sample(ff.AVSampleFormat(*samplefmt))
}
