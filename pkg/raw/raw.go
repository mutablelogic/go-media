package raw

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	_ "image/jpeg"
	"io"
	"os"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
	libraw "github.com/mutablelogic/go-media/sys/libraw"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// RAW wraps a libraw data handle for decoding RAW image files.
type RAW struct {
	data *libraw.Data
}

var _ io.Closer = (*RAW)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Open opens a RAW image file by path.
func Open(path string) (*RAW, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, media.ErrNotFound.Withf("%q", path)
	}
	data := libraw.Libraw_init(0)
	if data == nil {
		return nil, media.ErrInternalError.With("libraw init failed")
	}
	if rc := libraw.Libraw_open_file(data, path); rc != 0 {
		libraw.Libraw_close(data)
		return nil, media.ErrBadParameter.With(libraw.Libraw_strerror(rc))
	}
	return &RAW{data: data}, nil
}

// Read opens a RAW image from a reader.
func Read(r io.Reader) (*RAW, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Parse(buf)
}

// Parse opens a RAW image from a byte slice.
func Parse(buf []byte) (*RAW, error) {
	if len(buf) == 0 {
		return nil, media.ErrBadParameter.With("empty data")
	}
	data := libraw.Libraw_init(0)
	if data == nil {
		return nil, media.ErrInternalError.With("libraw init failed")
	}
	if rc := libraw.Libraw_open_buffer(data, buf); rc != 0 {
		libraw.Libraw_close(data)
		return nil, media.ErrBadParameter.With(libraw.Libraw_strerror(rc))
	}
	return &RAW{data: data}, nil
}

// Close releases the underlying libraw resources.
func (r *RAW) Close() error {
	if r.data != nil {
		libraw.Libraw_close(r.data)
		r.data = nil
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// ACCESSORS

// Make returns the camera manufacturer.
func (r *RAW) Make() string {
	return libraw.IParams_make(libraw.Libraw_get_iparams(r.data))
}

// Model returns the camera model.
func (r *RAW) Model() string {
	return libraw.IParams_model(libraw.Libraw_get_iparams(r.data))
}

// Software returns the firmware or software string embedded in the file.
func (r *RAW) Software() string {
	return libraw.IParams_software(libraw.Libraw_get_iparams(r.data))
}

// ISOSpeed returns the ISO speed rating.
func (r *RAW) ISOSpeed() float32 {
	return libraw.ImgOther_iso_speed(libraw.Libraw_get_imgother(r.data))
}

// Shutter returns the exposure time in seconds.
func (r *RAW) Shutter() float32 {
	return libraw.ImgOther_shutter(libraw.Libraw_get_imgother(r.data))
}

// Aperture returns the f-number.
func (r *RAW) Aperture() float32 {
	return libraw.ImgOther_aperture(libraw.Libraw_get_imgother(r.data))
}

// FocalLength returns the focal length in mm.
func (r *RAW) FocalLength() float32 {
	return libraw.ImgOther_focal_len(libraw.Libraw_get_imgother(r.data))
}

// Timestamp returns the capture time.
func (r *RAW) Timestamp() time.Time {
	return time.Unix(libraw.ImgOther_timestamp(libraw.Libraw_get_imgother(r.data)), 0)
}

// Width returns the processed image width in pixels (after crop/rotation).
func (r *RAW) Width() int {
	return libraw.Libraw_get_iwidth(r.data)
}

// Height returns the processed image height in pixels (after crop/rotation).
func (r *RAW) Height() int {
	return libraw.Libraw_get_iheight(r.data)
}

// RawWidth returns the full sensor width before any processing.
func (r *RAW) RawWidth() int {
	return libraw.Libraw_get_raw_width(r.data)
}

// RawHeight returns the full sensor height before any processing.
func (r *RAW) RawHeight() int {
	return libraw.Libraw_get_raw_height(r.data)
}

// XMP returns the raw XMP packet bytes embedded in the file, or nil if absent.
func (r *RAW) XMP() []byte {
	return libraw.IParams_xmpdata(libraw.Libraw_get_iparams(r.data))
}

////////////////////////////////////////////////////////////////////////////////
// METADATA

// Metadata returns key/value pairs for camera info and shooting parameters.
// Keys follow XMP namespace conventions (tiff:Make, exif:ISOSpeedRatings, etc.)
// and implement the media.Metadata interface.
func (r *RAW) Metadata() []media.Metadata {
	return newMetadata(r)
}

////////////////////////////////////////////////////////////////////////////////
// THUMBNAIL

// Thumbnail extracts and decodes the embedded preview image. This is fast
// since the thumbnail is pre-rendered inside the RAW file.
func (r *RAW) Thumbnail() (image.Image, error) {
	data, format, w, h, err := r.thumbnailRaw()
	if err != nil {
		return nil, err
	}
	switch format {
	case libraw.THUMBNAIL_JPEG:
		img, _, err := image.Decode(bytes.NewReader(data))
		return img, err
	case libraw.THUMBNAIL_BITMAP:
		return bitmapToImage(data, w, h, 3, 8), nil
	case libraw.THUMBNAIL_BITMAP16:
		return bitmapToImage(data, w, h, 3, 16), nil
	default:
		return nil, media.ErrNotImplemented.Withf("thumbnail format %v", format)
	}
}

// ThumbnailBytes returns the raw thumbnail bytes without decoding. For JPEG
// thumbnails (the common case) this can be passed directly to exif.Parse to
// extract embedded EXIF metadata.
func (r *RAW) ThumbnailBytes() ([]byte, error) {
	data, _, _, _, err := r.thumbnailRaw()
	return data, err
}

func (r *RAW) thumbnailRaw() ([]byte, libraw.ThumbnailFormat, int, int, error) {
	if rc := libraw.Libraw_unpack_thumb(r.data); rc != 0 {
		return nil, libraw.THUMBNAIL_UNKNOWN, 0, 0, media.ErrInternalError.With(libraw.Libraw_strerror(rc))
	}
	thumb := libraw.Libraw_get_thumbnail(r.data)
	data := libraw.Thumbnail_data(thumb)
	if len(data) == 0 {
		return nil, libraw.THUMBNAIL_UNKNOWN, 0, 0, media.ErrNotFound.With("no thumbnail data")
	}
	return data,
		libraw.Thumbnail_format(thumb),
		int(libraw.Thumbnail_width(thumb)),
		int(libraw.Thumbnail_height(thumb)),
		nil
}

////////////////////////////////////////////////////////////////////////////////
// IMAGE

// Image demosaics and returns the full-resolution image. This is a slow
// operation; use Thumbnail for quick previews.
func (r *RAW) Image() (image.Image, error) {
	if rc := libraw.Libraw_unpack(r.data); rc != 0 {
		return nil, media.ErrInternalError.With(libraw.Libraw_strerror(rc))
	}
	if rc := libraw.Libraw_dcraw_process(r.data); rc != 0 {
		return nil, media.ErrInternalError.With(libraw.Libraw_strerror(rc))
	}
	img, rc := libraw.Libraw_dcraw_make_mem_image(r.data)
	if img == nil || rc != 0 {
		return nil, media.ErrInternalError.With(libraw.Libraw_strerror(rc))
	}
	defer libraw.Libraw_dcraw_clear_mem(img)

	data := libraw.ProcessedImage_data(img)
	switch libraw.ProcessedImage_type(img) {
	case libraw.IMAGE_JPEG:
		goimg, _, err := image.Decode(bytes.NewReader(data))
		return goimg, err
	case libraw.IMAGE_BITMAP:
		return bitmapToImage(data,
			int(libraw.ProcessedImage_width(img)),
			int(libraw.ProcessedImage_height(img)),
			int(libraw.ProcessedImage_colors(img)),
			int(libraw.ProcessedImage_bits(img))), nil
	default:
		return nil, media.ErrNotImplemented.Withf("image format %v", libraw.ProcessedImage_type(img))
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE

// bitmapToImage converts libraw's interleaved RGB bitmap to image.Image.
// Libraw stores 16-bit values in little-endian host byte order.
func bitmapToImage(data []byte, width, height, colors, bits int) image.Image {
	if bits <= 8 {
		img := image.NewNRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				off := (y*width + x) * colors
				var rv, gv, bv uint8
				switch {
				case colors >= 3:
					rv, gv, bv = data[off], data[off+1], data[off+2]
				case colors == 1:
					rv, gv, bv = data[off], data[off], data[off]
				}
				img.SetNRGBA(x, y, color.NRGBA{R: rv, G: gv, B: bv, A: 255})
			}
		}
		return img
	}
	img := image.NewNRGBA64(image.Rect(0, 0, width, height))
	stride := colors * 2
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			off := (y*width + x) * stride
			var rv, gv, bv uint16
			switch {
			case colors >= 3:
				rv = binary.LittleEndian.Uint16(data[off:])
				gv = binary.LittleEndian.Uint16(data[off+2:])
				bv = binary.LittleEndian.Uint16(data[off+4:])
			case colors == 1:
				rv = binary.LittleEndian.Uint16(data[off:])
				gv, bv = rv, rv
			}
			img.SetNRGBA64(x, y, color.NRGBA64{R: rv, G: gv, B: bv, A: 0xffff})
		}
	}
	return img
}
