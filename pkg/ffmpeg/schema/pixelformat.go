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
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewPixelFormat(pixfmt ff.AVPixelFormat) *PixelFormat {
	if pixfmt == ff.AV_PIX_FMT_NONE {
		return nil
	}
	return &PixelFormat{
		AVPixelFormat: pixfmt,
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

func (r ListPixelFormatResponse) String() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
