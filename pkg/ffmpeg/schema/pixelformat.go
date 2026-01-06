package schema

import (
	"encoding/json"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type ListPixelFormatRequest struct {
	Name      string `json:"name"`
	NumPlanes int    `json:"num_planes"`
}

type ListPixelFormatResponse []PixelFormat

type PixelFormat struct {
	ff.AVPixelFormat
	IsRGB    bool `json:"is_rgb"`
	HasAlpha bool `json:"has_alpha"`
	IsPlanar bool `json:"is_planar"`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewPixelFormat(pixfmt ff.AVPixelFormat) *PixelFormat {
	if pixfmt == ff.AV_PIX_FMT_NONE {
		return nil
	}
	return &PixelFormat{
		AVPixelFormat: pixfmt,
		IsRGB:         ff.AVUtil_pix_fmt_is_rgb(pixfmt),
		HasAlpha:      ff.AVUtil_pix_fmt_has_alpha(pixfmt),
		IsPlanar:      ff.AVUtil_pix_fmt_is_planar(pixfmt),
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r PixelFormat) MarshalJSON() ([]byte, error) {
	return r.AVPixelFormat.MarshalJSON()
}

func (r PixelFormat) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
