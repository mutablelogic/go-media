package thumbnailer

import (
	"image"
	"testing"

	// Packages
	assert "github.com/stretchr/testify/assert"
)

func TestRotateYCbCr(t *testing.T) {
	// A 4x2 image, with a different luma at each corner, rotated clockwise
	src := image.NewYCbCr(image.Rect(0, 0, 4, 2), image.YCbCrSubsampleRatio420)
	set := func(img *image.YCbCr, x, y int, v uint8) { img.Y[img.YOffset(x, y)] = v }
	get := func(img *image.YCbCr, x, y int) uint8 { return img.Y[img.YOffset(x, y)] }
	set(src, 0, 0, 1) // top left
	set(src, 3, 0, 2) // top right
	set(src, 3, 1, 3) // bottom right
	set(src, 0, 1, 4) // bottom left
	src.Cb[src.COffset(0, 0)], src.Cb[src.COffset(2, 0)] = 10, 20

	// 90°: the top left goes to the top right
	dst := rotateYCbCr(src, 90)
	assert.Equal(t, image.Rect(0, 0, 2, 4), dst.Rect)
	assert.Equal(t, []uint8{1, 2, 3, 4}, []uint8{get(dst, 1, 0), get(dst, 1, 3), get(dst, 0, 3), get(dst, 0, 0)})
	assert.Equal(t, []uint8{10, 20}, []uint8{dst.Cb[dst.COffset(0, 0)], dst.Cb[dst.COffset(0, 2)]})

	// 180°: the top left goes to the bottom right
	dst = rotateYCbCr(src, 180)
	assert.Equal(t, image.Rect(0, 0, 4, 2), dst.Rect)
	assert.Equal(t, []uint8{1, 2, 3, 4}, []uint8{get(dst, 3, 1), get(dst, 0, 1), get(dst, 0, 0), get(dst, 3, 0)})

	// 270°: the top left goes to the bottom left
	dst = rotateYCbCr(src, 270)
	assert.Equal(t, image.Rect(0, 0, 2, 4), dst.Rect)
	assert.Equal(t, []uint8{1, 2, 3, 4}, []uint8{get(dst, 0, 3), get(dst, 0, 0), get(dst, 1, 0), get(dst, 1, 3)})

	// 0°: a copy
	dst = rotateYCbCr(src, 0)
	assert.Equal(t, src.Y, dst.Y)
}
