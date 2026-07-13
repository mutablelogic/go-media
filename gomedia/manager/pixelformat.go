package manager

import (
	"context"

	// Packages
	schema "github.com/mutablelogic/go-media/gomedia/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListPixelFormats returns all supported pixel formats.
func (m *Media) ListPixelFormats(_ context.Context, req schema.ListPixelFormatRequest) (schema.ListPixelFormatResponse, error) {
	var opaque uintptr
	result := make(schema.ListPixelFormatResponse, 0, 256)

	matches := func(pf *schema.PixelFormat) bool {
		if req.Name != "" && ff.AVUtil_get_pix_fmt_name(pf.AVPixelFormat) != req.Name {
			return false
		}
		if req.NumPlanes != 0 && ff.AVUtil_pix_fmt_count_planes(pf.AVPixelFormat) != req.NumPlanes {
			return false
		}
		return true
	}

	for {
		pixfmt := ff.AVUtil_next_pixel_fmt(&opaque)
		if pixfmt == ff.AV_PIX_FMT_NONE {
			break
		}
		if pixelformat := schema.NewPixelFormat(pixfmt); pixelformat != nil && matches(pixelformat) {
			result = append(result, *pixelformat)
		}
	}

	return result, nil
}
