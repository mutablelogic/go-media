package task

import (
	"context"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg80/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return all supported pixel formats
func (manager *Manager) ListPixelFormat(_ context.Context, req *schema.ListPixelFormatRequest) (schema.ListPixelFormatResponse, error) {
	var opaque uintptr
	result := make(schema.ListPixelFormatResponse, 0, 256)

	// Filter function
	matches := func(req *schema.ListPixelFormatRequest, pf *schema.PixelFormat) bool {
		if req == nil {
			return true
		}
		if req.Name != "" && pf.Name != req.Name {
			return false
		}
		if req.NumPlanes != 0 && pf.NumPlanes != req.NumPlanes {
			return false
		}
		return true
	}

	for {
		pixfmt := ff.AVUtil_next_pixel_fmt(&opaque)
		if pixfmt == ff.AV_PIX_FMT_NONE {
			break
		}
		if pixelformat := schema.NewPixelFormat(pixfmt); pixelformat != nil {
			if matches(req, pixelformat) {
				result = append(result, *pixelformat)
			}
		}
	}
	return result, nil
}
