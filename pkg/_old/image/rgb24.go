package image

import (
	"image"
	"image/color"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type RGB24 struct {
	// Pix holds the image's stream, in RGB order.
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

// ColorModel returns RGB color model.
func (p *RGB24) ColorModel() color.Model {
	return ColorModel
}

// Bounds implements image.Image.At
func (p *RGB24) Bounds() image.Rectangle {
	return p.Rect
}

// At implements image.Image.At
func (p *RGB24) At(x, y int) color.Color {
	return p.RGBAAt(x, y)
}

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

// RGBAAt returns the color of the pixel at (x, y) as RGBA.
func (p *RGB24) RGBAAt(x, y int) color.RGBA {
	if !(image.Point{x, y}.In(p.Rect)) {
		return color.RGBA{}
	}
	i := (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
	return color.RGBA{p.Pix[i+0], p.Pix[i+1], p.Pix[i+2], 0xFF}
}

// ColorModel is RGB color model instance
var ColorModel = color.ModelFunc(rgbModel)

func rgbModel(c color.Color) color.Color {
	if _, ok := c.(RGB); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	return RGB{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)}
}

// RGB color
type RGB struct {
	R, G, B uint8
}

// RGBA implements Color.RGBA
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
