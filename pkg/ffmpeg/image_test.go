package ffmpeg_test

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"testing"

	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	imagex "github.com/mutablelogic/go-media/pkg/image"
	assert "github.com/stretchr/testify/assert"
)

func Test_image_001(t *testing.T) {
	assert := assert.New(t)

	r, err := os.Open("../../etc/test/sample.png")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()
	image, err := png.Decode(r)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create a frame from the image
	frame, err := ffmpeg.FrameFromImage(image)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	// Create a new image from the frame
	image2, err := frame.ImageFromFrame()
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Compare the two images
	for y := 0; y < frame.Height(); y++ {
		for x := 0; x < frame.Width(); x++ {
			r1, g1, b1, a1 := image.At(x, y).RGBA()
			r2, g2, b2, a2 := image2.At(x, y).RGBA()
			assert.Equal(r1, r2)
			assert.Equal(g1, g2)
			assert.Equal(b1, b2)
			assert.Equal(a1, a2)
		}
	}
}

func Test_image_002(t *testing.T) {
	assert := assert.New(t)

	r, err := os.Open("../../etc/test/sample.png")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()
	img, err := png.Decode(r)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Convert to Grey8
	grey8 := image.NewGray(img.Bounds())
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			grey8.Set(x, y, color.Gray{uint8((r + g + b) / 3 >> 8)})
		}
	}

	// Create a frame from the image
	frame, err := ffmpeg.FrameFromImage(grey8)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	// Create a new image from the frame
	grey8_dest, err := frame.ImageFromFrame()
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Compare the two images
	for y := 0; y < frame.Height(); y++ {
		for x := 0; x < frame.Width(); x++ {
			r1, g1, b1, a1 := grey8.At(x, y).RGBA()
			r2, g2, b2, a2 := grey8_dest.At(x, y).RGBA()
			assert.Equal(r1, r2)
			assert.Equal(g1, g2)
			assert.Equal(b1, b2)
			assert.Equal(a1, a2)
		}
	}
}

func Test_image_003(t *testing.T) {
	assert := assert.New(t)

	r, err := os.Open("../../etc/test/sample.png")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()
	img, err := png.Decode(r)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Convert to RGB24
	rgb24 := imagex.NewRGB24(img.Bounds())
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rgb24.Set(x, y, imagex.RGB{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8)})
		}
	}

	// Create a frame from the image
	frame, err := ffmpeg.FrameFromImage(rgb24)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	// Create a new image from the frame
	rgb24_dest, err := frame.ImageFromFrame()
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Compare the two images
	for y := 0; y < frame.Height(); y++ {
		for x := 0; x < frame.Width(); x++ {
			r1, g1, b1, a1 := rgb24.At(x, y).RGBA()
			r2, g2, b2, a2 := rgb24_dest.At(x, y).RGBA()
			assert.Equal(r1, r2)
			assert.Equal(g1, g2)
			assert.Equal(b1, b2)
			assert.Equal(a1, a2)
		}
	}
}

func Test_image_004(t *testing.T) {
	assert := assert.New(t)

	r, err := os.Open("../../etc/test/sample.jpg")
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()
	source, err := jpeg.Decode(r)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create a frame from the image
	frame, err := ffmpeg.FrameFromImage(source)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer frame.Close()

	// Create a new image from the frame
	dest, err := frame.ImageFromFrame()
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Compare the two images
	for y := 0; y < frame.Height(); y++ {
		for x := 0; x < frame.Width(); x++ {
			r1, g1, b1, a1 := source.At(x, y).RGBA()
			r2, g2, b2, a2 := dest.At(x, y).RGBA()
			assert.Equal(r1, r2)
			assert.Equal(g1, g2)
			assert.Equal(b1, b2)
			assert.Equal(a1, a2)
		}
	}
}
