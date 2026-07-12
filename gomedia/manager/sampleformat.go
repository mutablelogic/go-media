package manager

import (
	"context"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListSampleFormats returns all supported sample formats.
func (m *Media) ListSampleFormats(_ context.Context, req schema.ListSampleFormatRequest) (schema.ListSampleFormatResponse, error) {
	var opaque uintptr
	result := make(schema.ListSampleFormatResponse, 0, 16)

	matches := func(sf *schema.SampleFormat) bool {
		if req.Name != "" && ff.AVUtil_get_sample_fmt_name(sf.AVSampleFormat) != req.Name {
			return false
		}
		if req.IsPlanar != nil && ff.AVUtil_sample_fmt_is_planar(sf.AVSampleFormat) != *req.IsPlanar {
			return false
		}
		return true
	}

	for {
		samplefmt := ff.AVUtil_next_sample_fmt(&opaque)
		if samplefmt == ff.AV_SAMPLE_FMT_NONE {
			break
		}
		if sampleformat := schema.NewSampleFormat(samplefmt); sampleformat != nil && matches(sampleformat) {
			result = append(result, *sampleformat)
		}
	}

	return result, nil
}
