package schema

import (
	"net/url"
	"strconv"
	"strings"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	types "github.com/mutablelogic/go-server/pkg/types"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type PixelFormatListRequest struct {
	Name      *string `json:"name,omitempty" help:"Filter by pixel format name." placeholder:"yuv420p" example:"yuv420p"`
	NumPlanes *uint64 `json:"num_planes,omitempty" help:"Filter by number of planes." placeholder:"3" example:"3"`
}

type PixelFormatList []PixelFormat

type PixelFormat struct {
	Name        string `json:"name" help:"Pixel format name." example:"yuv420p"`
	NumPlanes   int    `json:"num_planes" help:"Number of image planes." example:"3"`
	BitDepth    int    `json:"bit_depth" help:"Bits per pixel." example:"12"`
	IsPlanar    bool   `json:"is_planar" help:"Whether planes are stored separately rather than interleaved." example:"true"`
	IsRGB       bool   `json:"is_rgb" help:"Whether the format is RGB, rather than YUV." example:"false"`
	HasAlpha    bool   `json:"has_alpha" help:"Whether the format includes an alpha channel." example:"false"`
	IsHWAccel   bool   `json:"is_hwaccel" help:"Whether this is a hardware-acceleration surface with no accessible pixel data." example:"false"`
	IsFloat     bool   `json:"is_float" help:"Whether samples are floating point, rather than integer." example:"false"`
	IsBigEndian bool   `json:"is_bigendian" help:"Whether multi-byte samples are stored big-endian." example:"false"`

	// Private context for the pixel format
	ctx ff.AVPixelFormat
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewPixelFormat(pixfmt ff.AVPixelFormat) *PixelFormat {
	if pixfmt == ff.AV_PIX_FMT_NONE {
		return nil
	}
	return &PixelFormat{
		ctx:         pixfmt,
		Name:        ff.AVUtil_get_pix_fmt_name(pixfmt),
		NumPlanes:   ff.AVUtil_pix_fmt_count_planes(pixfmt),
		BitDepth:    ff.AVUtil_get_bits_per_pixel(pixfmt),
		IsPlanar:    ff.AVUtil_pix_fmt_is_planar(pixfmt),
		IsRGB:       ff.AVUtil_pix_fmt_is_rgb(pixfmt),
		HasAlpha:    ff.AVUtil_pix_fmt_has_alpha(pixfmt),
		IsHWAccel:   ff.AVUtil_pix_fmt_is_hwaccel(pixfmt),
		IsFloat:     ff.AVUtil_pix_fmt_is_float(pixfmt),
		IsBigEndian: ff.AVUtil_pix_fmt_is_be(pixfmt),
	}
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - QUERY

func (r PixelFormatListRequest) Query() url.Values {
	query := url.Values{}
	if r.Name != nil {
		query.Set("name", types.Value(r.Name))
	}
	if r.NumPlanes != nil {
		query.Set("num_planes", strconv.FormatUint(types.Value(r.NumPlanes), 10))
	}
	return query
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (r PixelFormat) String() string {
	return types.Stringify(r)
}

func (r PixelFormatList) String() string {
	return types.Stringify(r)
}

////////////////////////////////////////////////////////////////////////////////
// TABLE WRITER

func (PixelFormat) Header() []string {
	return []string{"Name", "Planes", "Bit Depth", "Flags"}
}

func (r PixelFormat) Cell(col int) string {
	switch col {
	case 0:
		return r.Name
	case 1:
		return strconv.Itoa(r.NumPlanes)
	case 2:
		if r.BitDepth > 0 {
			return strconv.Itoa(r.BitDepth)
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

// Context returns the underlying FFmpeg pixel format constant.
func (r PixelFormat) Context() ff.AVPixelFormat {
	return r.ctx
}

func (r PixelFormat) Flags() string {
	flags := make([]string, 0, 6)
	if r.IsPlanar {
		flags = append(flags, "planar")
	}
	if r.IsRGB {
		flags = append(flags, "rgb")
	}
	if r.HasAlpha {
		flags = append(flags, "alpha")
	}
	if r.IsHWAccel {
		flags = append(flags, "hw")
	}
	if r.IsFloat {
		flags = append(flags, "float")
	}
	if r.IsBigEndian {
		flags = append(flags, "bigendian")
	}
	return strings.Join(flags, ",")
}
