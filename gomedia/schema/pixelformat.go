package schema

import (
	"encoding/json"
	"strconv"
	"strings"

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
	return &PixelFormat{AVPixelFormat: pixfmt}
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

////////////////////////////////////////////////////////////////////////////////
// TABLE WRITER

func (PixelFormat) Header() []string {
	return []string{"Name", "Planes", "Bit Depth", "Flags"}
}

func (r PixelFormat) Cell(col int) string {
	switch col {
	case 0:
		return r.Name()
	case 1:
		return strconv.Itoa(r.NumPlanes())
	case 2:
		if d := r.BitDepth(); d > 0 {
			return strconv.Itoa(d)
		}
		return ""
	case 3:
		return r.Flags()
	default:
		return ""
	}
}

func (PixelFormat) Width(col int) int {
	switch col {
	case 0:
		return 20
	case 1:
		return 8
	case 2:
		return 10
	default:
		return 0
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (r PixelFormat) Name() string {
	return ff.AVUtil_get_pix_fmt_name(r.AVPixelFormat)
}

func (r PixelFormat) NumPlanes() int {
	return ff.AVUtil_pix_fmt_count_planes(r.AVPixelFormat)
}

func (r PixelFormat) BitDepth() int {
	return ff.AVUtil_get_bits_per_pixel(r.AVPixelFormat)
}

func (r PixelFormat) Flags() string {
	flags := make([]string, 0, 6)
	if ff.AVUtil_pix_fmt_is_planar(r.AVPixelFormat) {
		flags = append(flags, "planar")
	}
	if ff.AVUtil_pix_fmt_is_rgb(r.AVPixelFormat) {
		flags = append(flags, "rgb")
	}
	if ff.AVUtil_pix_fmt_has_alpha(r.AVPixelFormat) {
		flags = append(flags, "alpha")
	}
	if ff.AVUtil_pix_fmt_is_hwaccel(r.AVPixelFormat) {
		flags = append(flags, "hw")
	}
	if ff.AVUtil_pix_fmt_is_float(r.AVPixelFormat) {
		flags = append(flags, "float")
	}
	if ff.AVUtil_pix_fmt_is_be(r.AVPixelFormat) {
		flags = append(flags, "be")
	}
	return strings.Join(flags, ",")
}
