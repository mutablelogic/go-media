package ffmpeg

import (
	"encoding/json"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type PixelFormat ff.AVPixelFormat

type jsonPixelFormat struct {
	Name      string `json:"name"`
	IsPlanar  bool   `json:"is_planar"`
	NumPlanes int    `json:"num_planes,omitempty"`
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newPixelFormat(pixfmt ff.AVPixelFormat) *PixelFormat {
	if pixfmt == ff.AV_PIX_FMT_NONE {
		return nil
	} else if name := ff.AVUtil_get_pix_fmt_name(pixfmt); name == "" {
		return nil
	} else {
		return (*PixelFormat)(&pixfmt)
	}
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (pixfmt *PixelFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonPixelFormat{
		Name:      pixfmt.Name(),
		IsPlanar:  pixfmt.IsPlanar(),
		NumPlanes: pixfmt.NumPlanes(),
	})
}

func (pixfmt *PixelFormat) String() string {
	data, _ := json.MarshalIndent(pixfmt, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PROPERTIES

func (pixfmt *PixelFormat) Name() string {
	return ff.AVUtil_get_pix_fmt_name(ff.AVPixelFormat(*pixfmt))
}

func (pixfmt *PixelFormat) IsPlanar() bool {
	return ff.AVUtil_pix_fmt_count_planes(ff.AVPixelFormat(*pixfmt)) > 1
}

func (pixfmt *PixelFormat) NumPlanes() int {
	return ff.AVUtil_pix_fmt_count_planes(ff.AVPixelFormat(*pixfmt))
}
