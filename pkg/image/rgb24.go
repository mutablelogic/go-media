package image

import (
	"image"
	"image/color"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// RGB24 represents an in-memory image with 24-bit RGB color.
// Each pixel is stored as three consecutive bytes in RGB order.
// This format is commonly used for uncompressed image data from video frames.
type RGB24 struct {
	// Pix holds the image's pixels, in RGB order. The pixel at (x, y)
	// starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*3].
	Pix []uint8
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

// Make sure RGB24 implements image.Image.
var _ image.Image = (*RGB24)(nil)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewRGB24 returns a new RGB24 image with the given bounds.
// The image's pixels are initialized to zero (black).
func NewRGB24(r image.Rectangle) *RGB24 {
	w, h := r.Dx(), r.Dy()
	return &RGB24{
		Pix:    make([]uint8, 3*w*h),
		Stride: 3 * w,
		Rect:   r,
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// ColorModel returns the RGB color model.
func (p *RGB24) ColorModel() color.Model {
	return ColorModel
}

// Bounds returns the domain for which At can return non-zero color.
// The bounds do not necessarily contain the point (0, 0).
func (p *RGB24) Bounds() image.Rectangle {
	return p.Rect
}

// At returns the color of the pixel at (x, y).
// At(Bounds().Min.X, Bounds().Min.Y) returns the upper-left pixel of the grid.
// At(Bounds().Max.X-1, Bounds().Max.Y-1) returns the lower-right pixel.
func (p *RGB24) At(x, y int) color.Color {
	return p.RGBAt(x, y)
}

// Set sets the pixel at (x, y) to the given color.
// If the point is outside the image bounds, Set is a no-op.
func (p *RGB24) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
	r, g, b, _ := c.RGBA()
	p.Pix[i+0] = uint8(r >> 8)
	p.Pix[i+1] = uint8(g >> 8)
	p.Pix[i+2] = uint8(b >> 8)
}

// RGBAt returns the RGB color of the pixel at (x, y).
// If the point is outside the image bounds, RGBAt returns zero RGB.
func (p *RGB24) RGBAt(x, y int) RGB {
	if !(image.Point{x, y}.In(p.Rect)) {
		return RGB{}
	}
	i := (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
	return RGB{p.Pix[i+0], p.Pix[i+1], p.Pix[i+2]}
}

// ColorModel is the color model for RGB24 images, converting any color
// to 24-bit RGB format.
var ColorModel = color.ModelFunc(rgbModel)

func rgbModel(c color.Color) color.Color {
	if _, ok := c.(RGB); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	return RGB{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
}

// RGB represents a 24-bit color consisting of red, green, and blue components.
// Each component is an 8-bit value, giving a total of 16,777,216 possible colors.
type RGB struct {
	R, G, B uint8
}

// RGBA returns the alpha-premultiplied red, green, blue and alpha values
// for the RGB color. The alpha value is always fully opaque (0xFFFF).
// The red, green, and blue values are returned in the range [0, 0xFFFF].
func (c RGB) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	g = uint32(c.G)
	g |= g << 8
	b = uint32(c.B)
	b |= b << 8
	a = uint32(0xFFFF)
	return
}
