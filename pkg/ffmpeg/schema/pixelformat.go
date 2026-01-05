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
	Name          string `json:"name"`
	NumComponents int    `json:"num_components"`       // Number of components (1-4)
	NumPlanes     int    `json:"num_planes,omitempty"` // Number of planes
	BitsPerPixel  int    `json:"bits_per_pixel"`       // Bits per pixel
	IsPlanar      bool   `json:"is_planar"`            // Is planar format
	IsRGB         bool   `json:"is_rgb"`               // Is RGB-like format
	HasAlpha      bool   `json:"has_alpha"`            // Has alpha channel
	IsFloat       bool   `json:"is_float"`             // Is floating point format
	IsBigEndian   bool   `json:"is_big_endian"`        // Is big-endian format
	IsHWAccel     bool   `json:"is_hwaccel"`           // Is hardware accelerated
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewPixelFormat(pixfmt ff.AVPixelFormat) *PixelFormat {
	if pixfmt == ff.AV_PIX_FMT_NONE {
		return nil
	}
	name := ff.AVUtil_get_pix_fmt_name(pixfmt)
	if name == "" {
		return nil
	}
	return &PixelFormat{
		Name:          name,
		NumComponents: ff.AVUtil_pix_fmt_num_components(pixfmt),
		NumPlanes:     ff.AVUtil_pix_fmt_count_planes(pixfmt),
		BitsPerPixel:  ff.AVUtil_get_bits_per_pixel(pixfmt),
		IsPlanar:      ff.AVUtil_pix_fmt_is_planar(pixfmt),
		IsRGB:         ff.AVUtil_pix_fmt_is_rgb(pixfmt),
		HasAlpha:      ff.AVUtil_pix_fmt_has_alpha(pixfmt),
		IsFloat:       ff.AVUtil_pix_fmt_is_float(pixfmt),
		IsBigEndian:   ff.AVUtil_pix_fmt_is_be(pixfmt),
		IsHWAccel:     ff.AVUtil_pix_fmt_is_hwaccel(pixfmt),
	}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r PixelFormat) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
