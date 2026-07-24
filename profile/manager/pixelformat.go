package manager

import (
	"context"

	// Packages
	schema "github.com/mutablelogic/go-media/profile/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ListPixelFormats returns all supported pixel formats.
func (profile *Profile) ListPixelFormats(_ context.Context, req schema.PixelFormatListRequest) (schema.PixelFormatList, error) {
	var opaque uintptr
	result := make(schema.PixelFormatList, 0, 256)

	matches := func(pf *schema.PixelFormat) bool {
		if req.Name != nil && pf.Name != *req.Name {
			return false
		}
		if req.NumPlanes != nil && uint64(pf.NumPlanes) != *req.NumPlanes {
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
