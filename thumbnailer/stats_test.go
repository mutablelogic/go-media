package thumbnailer

import (
	"image"
	"testing"

	// Packages
	assert "github.com/stretchr/testify/assert"
)

func TestExposure(t *testing.T) {
	// Frames which are well exposed score fully, and dark and bright ones
	// less, in video range, where black is 16 and white is 235
	level := func(v float64) float64 { return (16 + v*219) / 255 }
	assert.Equal(t, 0.0, exposure(level(0)))
	assert.Equal(t, 0.0, exposure(level(0.02)))
	assert.InDelta(t, 0.5, exposure(level(0.11)), 0.01)
	assert.Equal(t, 1.0, exposure(level(0.2)))
	assert.Equal(t, 1.0, exposure(level(0.5)))
	assert.Equal(t, 1.0, exposure(level(0.8)))
	assert.InDelta(t, 0.5, exposure(level(0.9)), 0.01)
	assert.Equal(t, 0.0, exposure(level(1)))
}

// boxedFrame returns a frame of a mid-grey picture with noise, between bars of
// black at the top and bottom, rows high
func boxedFrame(w, h, rows int) *image.YCbCr {
	img := image.NewYCbCr(image.Rect(0, 0, w, h), image.YCbCrSubsampleRatio420)
	for y := range h {
		for x := range w {
			v := uint8(16)
			if y >= rows && y < h-rows {
				v = uint8(100 + (x*7+y*13)%40)
			}
			img.Y[img.YOffset(x, y)] = v
		}
	}
	for i := range img.Cb {
		img.Cb[i], img.Cr[i] = 128, 128
	}
	return img
}

func TestPictureOf(t *testing.T) {
	// Letterbox bars are left out
	assert.Equal(t, image.Rect(0, 40, 320, 200), pictureOf(boxedFrame(320, 240, 40)))

	// A frame without bars is all picture
	assert.Equal(t, image.Rect(0, 0, 320, 240), pictureOf(boxedFrame(320, 240, 0)))

	// Bars are at most a quarter of the frame each, so a frame which is all
	// black isn't all bars, though its sides are pillarbox bars too
	assert.Equal(t, image.Rect(80, 60, 240, 180), pictureOf(boxedFrame(320, 240, 120)))

	// A dark picture with grain isn't a bar
	img := boxedFrame(320, 240, 0)
	for y := range 240 {
		for x := range 320 {
			img.Y[img.YOffset(x, y)] = uint8(16 + (x*7+y*13)%20)
		}
	}
	assert.Equal(t, image.Rect(0, 0, 320, 240), pictureOf(img))
}

func TestMeasure_Letterbox(t *testing.T) {
	// The bars don't make a frame darker
	boxed, _ := measure(boxedFrame(320, 240, 40))
	full, _ := measure(boxedFrame(320, 240, 0))
	assert.InDelta(t, full.Brightness, boxed.Brightness, 0.01)
}
