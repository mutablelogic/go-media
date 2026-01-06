package task

import (
	"context"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return all supported sample formats
func (manager *Manager) ListSampleFormats(_ context.Context, req *schema.ListSampleFormatRequest) (schema.ListSampleFormatResponse, error) {
	var opaque uintptr
	result := make(schema.ListSampleFormatResponse, 0, 16)

	// Filter function
	matches := func(req *schema.ListSampleFormatRequest, sf *schema.SampleFormat) bool {
		if req == nil {
			return true
		}
		if req.Name != "" && sf.Name != req.Name {
			return false
		}
		if req.IsPlanar != nil && sf.IsPlanar != *req.IsPlanar {
			return false
		}
		return true
	}

	for {
		samplefmt := ff.AVUtil_next_sample_fmt(&opaque)
		if samplefmt == ff.AV_SAMPLE_FMT_NONE {
			break
		}
		if sampleformat := schema.NewSampleFormat(samplefmt); sampleformat != nil {
			if matches(req, sampleformat) {
				result = append(result, *sampleformat)
			}
		}
	}
	return result, nil
}
