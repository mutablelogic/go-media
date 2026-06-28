package exif

import (
	"encoding/binary"
	"encoding/json"
	"image"
	"math"

	media "github.com/mutablelogic/go-media"
	libexif "github.com/mutablelogic/go-media/sys/libexif"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// TagType is the numeric EXIF tag identifier.
type TagType uint

// IFD identifies which Image File Directory an entry belongs to.
type IFD uint

type Tag struct {
	ifd        IFD
	tag        TagType
	format     libexif.Format
	components uint
	data       []byte
	order      libexif.ByteOrder
	str        string
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE CONSTRUCTOR

func newTag(entry *libexif.Entry, ifd libexif.IFD, order libexif.ByteOrder) *Tag {
	return &Tag{
		ifd:        IFD(ifd),
		tag:        TagType(libexif.Exif_entry_get_tag(entry)),
		format:     libexif.Exif_entry_get_format(entry),
		components: libexif.Exif_entry_get_components(entry),
		data:       libexif.Exif_entry_get_data(entry),
		order:      order,
		str:        libexif.Exif_entry_get_value(entry),
	}
}

var _ media.Metadata = (*Tag)(nil)

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

// String implements fmt.Stringer.
func (t *Tag) String() string {
	return t.str
}

// MarshalJSON implements json.Marshaler.
func (t *Tag) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name       string `json:"name"`
		Value      string `json:"value"`
		Tag        uint   `json:"tag"`
		IFD        string `json:"ifd"`
		Order      string `json:"order"`
		Components uint   `json:"components"`
		Format     string `json:"format"`
	}{
		Name:       t.Key(),
		Value:      t.str,
		Tag:        uint(t.tag),
		IFD:        libexif.Exif_ifd_get_name(libexif.IFD(t.ifd)),
		Order:      libexif.Exif_byte_order_get_name(t.order),
		Components: t.components,
		Format:     libexif.Exif_format_get_name(t.format),
	})
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Tag returns the numeric tag identifier.
func (t *Tag) Tag() TagType {
	return t.tag
}

// IFD returns which Image File Directory this tag belongs to.
func (t *Tag) IFD() IFD {
	return t.ifd
}

// Key returns the tag name (implements media.Metadata).
func (t *Tag) Key() string {
	return libexif.Exif_tag_get_name_in_ifd(libexif.Tag(t.tag), libexif.IFD(t.ifd))
}

// Name returns the tag name.
func (t *Tag) Name() string {
	return t.Key()
}

// Format returns the EXIF data format of the tag.
func (t *Tag) Format() libexif.Format {
	return t.format
}

// Components returns the number of values in the tag's data array.
func (t *Tag) Components() uint {
	return t.components
}

// Value returns the libexif-formatted string representation of the tag value
// (implements media.Metadata).
func (t *Tag) Value() string {
	return t.str
}

// Bytes returns nil for EXIF tags (implements media.Metadata).
func (t *Tag) Bytes() []byte {
	return nil
}

// Image returns nil for EXIF tags (implements media.Metadata).
func (t *Tag) Image() image.Image {
	return nil
}

// Any returns the tag's data decoded into an appropriate Go type
// (implements media.Metadata). For single-component tags: uint8, int8, uint16,
// int16, uint32, int32, libexif.Rational, libexif.SRational, float32, float64,
// or string (ASCII). For multi-component tags the corresponding slice type is
// returned. UNDEFINED format is returned as []byte.
func (t *Tag) Any() any {
	if len(t.data) == 0 {
		return nil
	}
	size := libexif.Exif_format_get_size(t.format)
	if size == 0 {
		return nil
	}

	n := t.components
	pick := func(i uint) []byte { return t.data[i*size : i*size+size] }

	switch t.format {
	case libexif.EXIF_FORMAT_ASCII:
		// Strip trailing null bytes
		s := string(t.data)
		if l := len(s); l > 0 && s[l-1] == 0 {
			s = s[:l-1]
		}
		return s
	case libexif.EXIF_FORMAT_UNDEFINED:
		return append([]byte(nil), t.data...)
	case libexif.EXIF_FORMAT_BYTE:
		if n == 1 {
			return t.data[0]
		}
		return append([]byte(nil), t.data...)
	case libexif.EXIF_FORMAT_SBYTE:
		if n == 1 {
			return int8(t.data[0])
		}
		out := make([]int8, n)
		for i := range out {
			out[i] = int8(t.data[i])
		}
		return out
	case libexif.EXIF_FORMAT_SHORT:
		if n == 1 {
			return libexif.Exif_get_short(pick(0), t.order)
		}
		out := make([]uint16, n)
		for i := range out {
			out[i] = libexif.Exif_get_short(pick(uint(i)), t.order)
		}
		return out
	case libexif.EXIF_FORMAT_SSHORT:
		if n == 1 {
			return libexif.Exif_get_sshort(pick(0), t.order)
		}
		out := make([]int16, n)
		for i := range out {
			out[i] = libexif.Exif_get_sshort(pick(uint(i)), t.order)
		}
		return out
	case libexif.EXIF_FORMAT_LONG:
		if n == 1 {
			return libexif.Exif_get_long(pick(0), t.order)
		}
		out := make([]uint32, n)
		for i := range out {
			out[i] = libexif.Exif_get_long(pick(uint(i)), t.order)
		}
		return out
	case libexif.EXIF_FORMAT_SLONG:
		if n == 1 {
			return libexif.Exif_get_slong(pick(0), t.order)
		}
		out := make([]int32, n)
		for i := range out {
			out[i] = libexif.Exif_get_slong(pick(uint(i)), t.order)
		}
		return out
	case libexif.EXIF_FORMAT_RATIONAL:
		if n == 1 {
			return libexif.Exif_get_rational(pick(0), t.order)
		}
		out := make([]libexif.Rational, n)
		for i := range out {
			out[i] = libexif.Exif_get_rational(pick(uint(i)), t.order)
		}
		return out
	case libexif.EXIF_FORMAT_SRATIONAL:
		if n == 1 {
			return libexif.Exif_get_srational(pick(0), t.order)
		}
		out := make([]libexif.SRational, n)
		for i := range out {
			out[i] = libexif.Exif_get_srational(pick(uint(i)), t.order)
		}
		return out
	case libexif.EXIF_FORMAT_FLOAT:
		byteOrder := t.byteOrder()
		if n == 1 {
			return math.Float32frombits(byteOrder.Uint32(pick(0)))
		}
		out := make([]float32, n)
		for i := range out {
			out[i] = math.Float32frombits(byteOrder.Uint32(pick(uint(i))))
		}
		return out
	case libexif.EXIF_FORMAT_DOUBLE:
		byteOrder := t.byteOrder()
		if n == 1 {
			return math.Float64frombits(byteOrder.Uint64(pick(0)))
		}
		out := make([]float64, n)
		for i := range out {
			out[i] = math.Float64frombits(byteOrder.Uint64(pick(uint(i))))
		}
		return out
	}
	return nil
}

func (t *Tag) byteOrder() binary.ByteOrder {
	if t.order == libexif.EXIF_BYTE_ORDER_MOTOROLA {
		return binary.BigEndian
	}
	return binary.LittleEndian
}
