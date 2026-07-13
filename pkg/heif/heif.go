package heif

import (
	"encoding/binary"
	"image"
	"image/color"
	"io"
	"os"

	// Packages
	media "github.com/mutablelogic/go-media"
	"github.com/mutablelogic/go-media/pkg/xmp"
	libheif "github.com/mutablelogic/go-media/sys/libheif"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// HEIF wraps a libheif context and exposes the primary image as an image.Image.
type HEIF struct {
	ctx  *libheif.Context
	data []byte
	img  image.Image
}

var _ io.Closer = (*HEIF)(nil)
var _ image.Image = (*HEIF)(nil)

func init() {
	image.RegisterFormat("heif", "????ftypheic", decodeImage, decodeConfig)
	image.RegisterFormat("heif", "????ftypheim", decodeImage, decodeConfig)
	image.RegisterFormat("heif", "????ftypheis", decodeImage, decodeConfig)
	image.RegisterFormat("heif", "????ftypheix", decodeImage, decodeConfig)
	image.RegisterFormat("heif", "????ftyphevc", decodeImage, decodeConfig)
	image.RegisterFormat("heif", "????ftyphevm", decodeImage, decodeConfig)
	image.RegisterFormat("heif", "????ftyphevs", decodeImage, decodeConfig)
	image.RegisterFormat("heif", "????ftypmif1", decodeImage, decodeConfig)
	image.RegisterFormat("avif", "????ftypavif", decodeImage, decodeConfig)
	image.RegisterFormat("avif", "????ftypavis", decodeImage, decodeConfig)
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Open opens a HEIF image by path and decodes its primary image.
func Open(path string) (*HEIF, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, media.ErrNotFound.Withf("%q", path)
	}
	ctx := libheif.Libheif_context_alloc()
	if ctx == nil {
		return nil, media.ErrInternalError.With("libheif context alloc failed")
	}
	if err := libheif.Libheif_context_read_from_file(ctx, path); err != nil {
		libheif.Libheif_context_free(ctx)
		return nil, err
	}
	return newHEIF(ctx, nil)
}

// Read opens a HEIF image from a reader and decodes its primary image.
func Read(r io.Reader) (*HEIF, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Parse(buf)
}

// Parse opens a HEIF image from a byte slice and decodes its primary image.
func Parse(buf []byte) (*HEIF, error) {
	if len(buf) == 0 {
		return nil, media.ErrBadParameter.With("empty data")
	}
	ctx := libheif.Libheif_context_alloc()
	if ctx == nil {
		return nil, media.ErrInternalError.With("libheif context alloc failed")
	}
	if err := libheif.Libheif_context_read_from_memory_without_copy(ctx, buf); err != nil {
		libheif.Libheif_context_free(ctx)
		return nil, err
	}
	return newHEIF(ctx, buf)
}

// Close releases the underlying libheif resources.
func (h *HEIF) Close() error {
	if h == nil {
		return nil
	}
	if h.ctx != nil {
		libheif.Libheif_context_free(h.ctx)
		h.ctx = nil
	}
	h.data = nil
	h.img = nil
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// IMAGE IMAGE

// Bounds returns the bounds of the primary image.
func (h *HEIF) Bounds() image.Rectangle {
	if h == nil || h.img == nil {
		return image.Rectangle{}
	}
	return h.img.Bounds()
}

// ColorModel returns the color model of the primary image.
func (h *HEIF) ColorModel() color.Model {
	if h == nil || h.img == nil {
		return color.NRGBAModel
	}
	return h.img.ColorModel()
}

// At returns the color of the pixel at x, y in the primary image.
func (h *HEIF) At(x, y int) color.Color {
	if h == nil || h.img == nil {
		return color.NRGBA{}
	}
	return h.img.At(x, y)
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Primary returns the decoded primary image.
func (h *HEIF) Primary() image.Image {
	if h == nil {
		return nil
	}
	return h.img
}

// Thumbnails returns any attached thumbnail images for the primary image.
func (h *HEIF) Thumbnails() []image.Image {
	if h == nil || h.ctx == nil {
		return nil
	}

	handle, err := libheif.Libheif_context_get_primary_image_handle(h.ctx)
	if err != nil || handle == nil {
		if h.img == nil {
			return nil
		}
		return []image.Image{h.img}
	}
	defer libheif.Libheif_image_handle_release(handle)

	count := libheif.Libheif_image_handle_get_number_of_thumbnails(handle)
	if count <= 0 {
		return nil
	}

	images := make([]image.Image, 0, count)

	for _, id := range libheif.Libheif_image_handle_get_list_of_thumbnail_IDs(handle, count) {
		thumbHandle, err := libheif.Libheif_image_handle_get_thumbnail(handle, id)
		if err != nil || thumbHandle == nil {
			continue
		}
		thumbImage, err := libheif.Libheif_decode_image(thumbHandle, libheif.HEIF_COLORSPACE_RGB, libheif.HEIF_CHROMA_INTERLEAVED_RGB)
		libheif.Libheif_image_handle_release(thumbHandle)
		if err != nil || thumbImage == nil {
			continue
		}
		images = append(images, decodeToImage(thumbImage))
		libheif.Libheif_image_release(thumbImage)
	}

	return images
}

// XMP returns the parsed XMP document from the primary image, if present.
func (h *HEIF) XMP() *xmp.XMP {
	if h == nil || h.ctx == nil {
		return nil
	}

	handle, err := libheif.Libheif_context_get_primary_image_handle(h.ctx)
	if err != nil || handle == nil {
		return nil
	}
	defer libheif.Libheif_image_handle_release(handle)

	doc := xmp.New()
	for _, block := range h.xmpMetadataBlocks(handle) {
		doc.Add(block...)
	}
	if len(doc.Items()) == 0 {
		return nil
	}
	return doc
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE

func newHEIF(ctx *libheif.Context, data []byte) (*HEIF, error) {
	h := &HEIF{ctx: ctx, data: data}
	img, err := decodePrimaryImage(ctx)
	if err != nil {
		libheif.Libheif_context_free(ctx)
		return nil, err
	}
	h.img = img
	return h, nil
}

func decodePrimaryImage(ctx *libheif.Context) (image.Image, error) {
	handle, err := libheif.Libheif_context_get_primary_image_handle(ctx)
	if err != nil {
		return nil, err
	}
	if handle == nil {
		return nil, media.ErrNotFound.With("primary image")
	}
	defer libheif.Libheif_image_handle_release(handle)

	alpha := libheif.Libheif_image_handle_has_alpha_channel(handle)
	lumaBits := libheif.Libheif_image_handle_get_luma_bits_per_pixel(handle)
	chromaBits := libheif.Libheif_image_handle_get_chroma_bits_per_pixel(handle)
	bits := lumaBits
	if chromaBits > bits {
		bits = chromaBits
	}

	chroma := libheif.HEIF_CHROMA_INTERLEAVED_RGB
	if alpha {
		chroma = libheif.HEIF_CHROMA_INTERLEAVED_RGBA
	}
	if bits > 8 {
		if alpha {
			chroma = libheif.HEIF_CHROMA_INTERLEAVED_RRGGBBAA_LE
		} else {
			chroma = libheif.HEIF_CHROMA_INTERLEAVED_RRGGBB_LE
		}
	}

	decoded, err := libheif.Libheif_decode_image(handle, libheif.HEIF_COLORSPACE_RGB, chroma)
	if err != nil {
		return nil, err
	}
	if decoded == nil {
		return nil, media.ErrInternalError.With("decode returned nil image")
	}
	defer libheif.Libheif_image_release(decoded)

	return decodeToImage(decoded), nil
}

func decodeToImage(img *libheif.Image) image.Image {
	width := libheif.Libheif_image_get_primary_width(img)
	height := libheif.Libheif_image_get_primary_height(img)
	if width <= 0 || height <= 0 {
		return image.NewNRGBA(image.Rect(0, 0, 0, 0))
	}

	plane, stride := libheif.Libheif_image_get_plane_readonly(img, libheif.HEIF_CHANNEL_INTERLEAVED)
	if len(plane) == 0 || stride <= 0 {
		return image.NewNRGBA(image.Rect(0, 0, 0, 0))
	}

	bits := libheif.Libheif_image_get_bits_per_pixel_range(img, libheif.HEIF_CHANNEL_INTERLEAVED)
	alpha := libheif.Libheif_image_has_channel(img, libheif.HEIF_CHANNEL_ALPHA)
	channels := 3
	if alpha {
		channels = 4
	}

	if bits <= 8 {
		out := image.NewNRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			srcRow := plane[y*stride : y*stride+width*channels]
			dstRow := out.Pix[y*out.Stride : y*out.Stride+width*4]
			for x := 0; x < width; x++ {
				si := x * channels
				di := x * 4
				dstRow[di+0] = srcRow[si+0]
				dstRow[di+1] = srcRow[si+1]
				dstRow[di+2] = srcRow[si+2]
				if alpha {
					dstRow[di+3] = srcRow[si+3]
				} else {
					dstRow[di+3] = 0xff
				}
			}
		}
		return out
	}

	out := image.NewNRGBA64(image.Rect(0, 0, width, height))
	pixelSize := channels * 2
	for y := 0; y < height; y++ {
		srcRow := plane[y*stride : y*stride+width*pixelSize]
		dstRow := out.Pix[y*out.Stride : y*out.Stride+width*8]
		for x := 0; x < width; x++ {
			si := x * pixelSize
			di := x * 8
			copy(dstRow[di:di+2], srcRow[si:si+2])
			copy(dstRow[di+2:di+4], srcRow[si+2:si+4])
			copy(dstRow[di+4:di+6], srcRow[si+4:si+6])
			if alpha {
				copy(dstRow[di+6:di+8], srcRow[si+6:si+8])
			} else {
				binary.LittleEndian.PutUint16(dstRow[di+6:di+8], 0xffff)
			}
		}
	}
	return out
}

func decodeImage(r io.Reader) (image.Image, error) {
	h, err := Read(r)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func decodeConfig(r io.Reader) (image.Config, error) {
	h, err := Read(r)
	if err != nil {
		return image.Config{}, err
	}
	defer h.Close()

	bounds := h.Bounds()
	return image.Config{
		ColorModel: h.ColorModel(),
		Width:      bounds.Dx(),
		Height:     bounds.Dy(),
	}, nil
}
