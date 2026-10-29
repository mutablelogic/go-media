package frame_test

import (
	"image"
	"image/color"
	"testing"

	// Packages
	frame "github.com/mutablelogic/go-media/frame"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

// fill sets every pixel of an image to a colour which depends on its position
func fill(img interface {
	image.Image
	Set(x, y int, c color.Color)
}) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 17), uint8(y * 31), uint8(x*y + 7), 0xFF})
		}
	}
}

// assertSame checks that two images have the same size and colours, comparing
// pixels relative to the top left of each
func assertSame(t *testing.T, want, got image.Image) {
	t.Helper()
	wb, gb := want.Bounds(), got.Bounds()
	require.Equal(t, wb.Dx(), gb.Dx())
	require.Equal(t, wb.Dy(), gb.Dy())
	for y := range wb.Dy() {
		for x := range wb.Dx() {
			wr, wg, wbl, wa := want.At(wb.Min.X+x, wb.Min.Y+y).RGBA()
			gr, gg, gbl, ga := got.At(gb.Min.X+x, gb.Min.Y+y).RGBA()
			if wr != gr || wg != gg || wbl != gbl || wa != ga {
				t.Fatalf("pixel (%d,%d): want %v, got %v", x, y, want.At(wb.Min.X+x, wb.Min.Y+y), got.At(gb.Min.X+x, gb.Min.Y+y))
			}
		}
	}
}

func roundTrip(t *testing.T, src image.Image) image.Image {
	t.Helper()
	f, err := frame.NewVideoFrame(0)
	require.NoError(t, err)
	t.Cleanup(func() { f.Close() })
	require.NoError(t, f.FromImage(src))
	img, err := f.Image()
	require.NoError(t, err)
	return img
}

func TestImage_Packed(t *testing.T) {
	r := image.Rect(0, 0, 13, 7)
	rgba, nrgba, gray, rgb := image.NewRGBA(r), image.NewNRGBA(r), image.NewGray(r), frame.NewRGB24(r)
	for _, img := range []interface {
		image.Image
		Set(x, y int, c color.Color)
	}{rgba, nrgba, gray, rgb} {
		fill(img)
	}

	got := roundTrip(t, rgba)
	assert.IsType(t, &image.RGBA{}, got)
	assertSame(t, rgba, got)

	got = roundTrip(t, nrgba)
	assert.IsType(t, &image.RGBA{}, got)
	assertSame(t, nrgba, got)

	got = roundTrip(t, gray)
	assert.IsType(t, &image.Gray{}, got)
	assertSame(t, gray, got)

	got = roundTrip(t, rgb)
	assert.IsType(t, &frame.RGB24{}, got)
	assertSame(t, rgb, got)
}

func TestImage_YCbCr(t *testing.T) {
	for _, ratio := range []image.YCbCrSubsampleRatio{
		image.YCbCrSubsampleRatio411, image.YCbCrSubsampleRatio420,
		image.YCbCrSubsampleRatio422, image.YCbCrSubsampleRatio440, image.YCbCrSubsampleRatio444,
	} {
		t.Run(ratio.String(), func(t *testing.T) {
			// An odd size, so the chroma planes are rounded up
			src := image.NewYCbCr(image.Rect(0, 0, 13, 7), ratio)
			for i := range src.Y {
				src.Y[i] = uint8(i * 7)
			}
			for i := range src.Cb {
				src.Cb[i], src.Cr[i] = uint8(i*11), uint8(i*13)
			}
			got := roundTrip(t, src)
			require.IsType(t, &image.YCbCr{}, got)
			assert.Equal(t, ratio, got.(*image.YCbCr).SubsampleRatio)
			assertSame(t, src, got)
		})
	}
}

func TestImage_SubImage(t *testing.T) {
	// A sub-image is copied from its own top left, not the start of Pix
	rgba := image.NewRGBA(image.Rect(0, 0, 20, 10))
	fill(rgba)
	sub := rgba.SubImage(image.Rect(5, 3, 15, 8))
	assertSame(t, sub, roundTrip(t, sub))

	yuv := image.NewYCbCr(image.Rect(0, 0, 20, 10), image.YCbCrSubsampleRatio420)
	for i := range yuv.Y {
		yuv.Y[i] = uint8(i)
	}
	for i := range yuv.Cb {
		yuv.Cb[i], yuv.Cr[i] = uint8(i*3), uint8(i*5)
	}
	subYUV := yuv.SubImage(image.Rect(4, 2, 14, 8))
	assertSame(t, subYUV, roundTrip(t, subYUV))
}

func TestImage_Reuse(t *testing.T) {
	// A frame is reallocated for a different format or size, and keeps its pts
	f, err := frame.NewVideoFrame(0)
	require.NoError(t, err)
	defer f.Close()
	f.SetPts(42)

	small := image.NewGray(image.Rect(0, 0, 4, 4))
	fill(small)
	require.NoError(t, f.FromImage(small))
	large := image.NewRGBA(image.Rect(0, 0, 16, 9))
	fill(large)
	require.NoError(t, f.FromImage(large))

	assert.Equal(t, ff.AV_PIX_FMT_RGBA, f.PixFmt())
	assert.Equal(t, int64(42), f.Pts())
	img, err := f.Image()
	require.NoError(t, err)
	assertSame(t, large, img)
}

func TestImage_Unsupported(t *testing.T) {
	f, err := frame.NewVideoFrame(0)
	require.NoError(t, err)
	defer f.Close()

	// Image types which aren't supported. FFmpeg's YUV410P subsamples chroma
	// 4x4, but a 4:1:0 image subsamples it 4x2
	assert.Error(t, f.FromImage(image.NewPaletted(image.Rect(0, 0, 4, 4), nil)))
	assert.Error(t, f.FromImage(image.NewYCbCr(image.Rect(0, 0, 4, 4), image.YCbCrSubsampleRatio410)))

	// A pixel format which isn't supported
	f.SetPixFmt(ff.AV_PIX_FMT_NV12)
	f.SetWidth(4)
	f.SetHeight(4)
	require.NoError(t, f.AllocateBuffers())
	_, err = f.Image()
	assert.Error(t, err)
}

func TestRGB24(t *testing.T) {
	img := frame.NewRGB24(image.Rect(2, 3, 6, 5))
	assert.Equal(t, image.Rect(2, 3, 6, 5), img.Bounds())

	img.Set(3, 4, color.RGBA{0x10, 0x20, 0x30, 0xFF})
	assert.Equal(t, frame.RGB{R: 0x10, G: 0x20, B: 0x30}, img.RGBAt(3, 4))
	assert.Equal(t, frame.RGB{}, img.RGBAt(0, 0))

	// Outside the bounds is ignored
	img.Set(0, 0, color.White)
	assert.Equal(t, frame.RGB{}, img.RGBAt(0, 0))

	// Colours are converted to opaque 24-bit RGB
	assert.Equal(t, frame.RGB{R: 0xFF, G: 0xFF, B: 0xFF}, frame.RGB24Model.Convert(color.White))
	_, _, _, a := frame.RGB{}.RGBA()
	assert.Equal(t, uint32(0xFFFF), a)
}
