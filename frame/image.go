package frame

import (
	"image"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// ImageOpt is an option for VideoFrame.Image
type ImageOpt func(*imageOpts) error

type imageOpts struct {
	scale         bool // whether the image is scaled, with square pixels
	width, height int  // size, either of which can be zero to keep the aspect ratio
}

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

// Planar YUV pixel formats and the image subsample ratios which match them.
// YUV410P has a chroma sample for each 4x4 pixels, but YCbCrSubsampleRatio410
// for each 4x2, so there's no image ratio for it
var (
	yuvSubsampleRatio = map[ff.AVPixelFormat]image.YCbCrSubsampleRatio{
		ff.AV_PIX_FMT_YUV411P: image.YCbCrSubsampleRatio411,
		ff.AV_PIX_FMT_YUV420P: image.YCbCrSubsampleRatio420,
		ff.AV_PIX_FMT_YUV422P: image.YCbCrSubsampleRatio422,
		ff.AV_PIX_FMT_YUV440P: image.YCbCrSubsampleRatio440,
		ff.AV_PIX_FMT_YUV444P: image.YCbCrSubsampleRatio444,
	}
	pixfmtYCbCr = map[image.YCbCrSubsampleRatio]ff.AVPixelFormat{
		image.YCbCrSubsampleRatio411: ff.AV_PIX_FMT_YUV411P,
		image.YCbCrSubsampleRatio420: ff.AV_PIX_FMT_YUV420P,
		image.YCbCrSubsampleRatio422: ff.AV_PIX_FMT_YUV422P,
		image.YCbCrSubsampleRatio440: ff.AV_PIX_FMT_YUV440P,
		image.YCbCrSubsampleRatio444: ff.AV_PIX_FMT_YUV444P,
	}
)

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// WithSquarePixels returns an image of the frame at the size it's
// displayed, with its sample aspect ratio applied, so its pixels are square,
// as an anamorphic video's, such as a DVD's, aren't.
func WithSquarePixels() ImageOpt {
	return func(o *imageOpts) error {
		o.scale = true
		return nil
	}
}

// WithImageSize returns an image of the frame scaled to a size, with square
// pixels. Either can be zero, to keep the aspect ratio the frame is
// displayed with, and both zero is the size it's displayed at.
func WithImageSize(width, height int) ImageOpt {
	return func(o *imageOpts) error {
		if width < 0 || height < 0 {
			return gomedia.ErrBadParameter.Withf("invalid size %dx%d", width, height)
		}
		o.scale, o.width, o.height = true, width, height
		return nil
	}
}

// DisplaySize returns the size the frame is displayed at, with square
// pixels: its size with its sample aspect ratio (the width of a pixel
// divided by its height) applied to its width. An unknown or invalid ratio
// is square.
func (f *VideoFrame) DisplaySize() (int, int) {
	w, h := f.Width(), f.Height()
	sar := f.SampleAspectRatio()
	if sar.Num() <= 0 || sar.Den() <= 0 || sar.Num() == sar.Den() {
		return w, h
	}
	return int((int64(w)*int64(sar.Num()) + int64(sar.Den())/2) / int64(sar.Den())), h
}

// Image returns an image of the frame. Without options, it shares the
// frame's pixels, so it's only valid until the frame is closed or reused.
// With WithSquarePixels or WithImageSize, it's a copy, scaled so its pixels
// are square. The pixel format is one of RGBA, GRAY8, RGB24, or a planar YUV
// format (YUV411P, YUV420P, YUV422P, YUV440P or YUV444P); convert other
// formats first, such as with a writer.Resampler. An RGBA frame is an
// *image.NRGBA, as FFmpeg's alpha isn't premultiplied.
func (f *VideoFrame) Image(opt ...ImageOpt) (image.Image, error) {
	var o imageOpts
	for _, fn := range opt {
		if err := fn(&o); err != nil {
			return nil, err
		}
	}
	img, err := f.view()
	if err != nil || !o.scale {
		return img, err
	}

	// Scale to the size, keeping the displayed aspect ratio, rounded to
	// whole chroma samples
	w, h := f.scaledSize(o.width, o.height)
	if w == f.Width() && h == f.Height() {
		return cloneImage(img), nil
	}
	scaled, err := f.scale(w, h)
	if err != nil {
		return nil, err
	}
	defer scaled.Close()
	img, err = scaled.view()
	if err != nil {
		return nil, err
	}
	return cloneImage(img), nil
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - IMAGE

// view returns an image which shares the frame's pixels
func (f *VideoFrame) view() (image.Image, error) {
	rect := image.Rect(0, 0, f.Width(), f.Height())
	switch f.PixFmt() {
	case ff.AV_PIX_FMT_RGBA:
		return &image.NRGBA{Pix: f.Bytes(0), Stride: f.Linesize(0), Rect: rect}, nil
	case ff.AV_PIX_FMT_GRAY8:
		return &image.Gray{Pix: f.Bytes(0), Stride: f.Linesize(0), Rect: rect}, nil
	case ff.AV_PIX_FMT_RGB24:
		return &RGB24{Pix: f.Bytes(0), Stride: f.Linesize(0), Rect: rect}, nil
	}
	if ratio, exists := yuvSubsampleRatio[f.PixFmt()]; exists {
		return &image.YCbCr{
			Y:              f.Bytes(0),
			Cb:             f.Bytes(1),
			Cr:             f.Bytes(2),
			YStride:        f.Linesize(0),
			CStride:        f.Linesize(1),
			SubsampleRatio: ratio,
			Rect:           rect,
		}, nil
	}
	return nil, gomedia.ErrNotImplemented.Withf("unsupported pixel format: %v", f.PixFmt())
}

// FromImage copies an image into the frame, reallocating it if the pixel
// format or size is different, and keeping its timebase and pts. The image
// is one of *image.NRGBA or *image.RGBA (to RGBA, whose premultiplied alpha
// is converted), *image.Gray (to GRAY8), *RGB24 (to RGB24), or *image.YCbCr
// with any subsample ratio except 4:1:0 (to planar YUV).
func (f *VideoFrame) FromImage(src image.Image) error {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	min := src.Bounds().Min
	switch src := src.(type) {
	case *image.RGBA:
		return f.fromRGBA(src)
	case *image.NRGBA:
		return f.fromPacked(ff.AV_PIX_FMT_RGBA, w, h, src.Pix[src.PixOffset(min.X, min.Y):], src.Stride, 4*w)
	case *image.Gray:
		return f.fromPacked(ff.AV_PIX_FMT_GRAY8, w, h, src.Pix[src.PixOffset(min.X, min.Y):], src.Stride, w)
	case *RGB24:
		return f.fromPacked(ff.AV_PIX_FMT_RGB24, w, h, src.Pix[src.PixOffset(min.X, min.Y):], src.Stride, 3*w)
	case *image.YCbCr:
		if pixfmt, exists := pixfmtYCbCr[src.SubsampleRatio]; exists {
			return f.fromYCbCr(pixfmt, src)
		}
	}
	return gomedia.ErrNotImplemented.Withf("unsupported image: %T", src)
}

// scaledSize returns the size of an image of the frame, with square pixels,
// keeping the displayed aspect ratio when either is zero, rounded down to
// whole chroma samples, and at least one of them
func (f *VideoFrame) scaledSize(width, height int) (int, int) {
	dw, dh := f.DisplaySize()
	switch {
	case width == 0 && height == 0:
		width, height = dw, dh
	case height == 0:
		height = (width*dh + dw/2) / dw
	case width == 0:
		width = (height*dw + dh/2) / dh
	}
	sx, sy := 1, 1
	if ratio, exists := yuvSubsampleRatio[f.PixFmt()]; exists {
		sx, sy = chromaSubsample(ratio)
	}
	return max(sx, width/sx*sx), max(sy, height/sy*sy)
}

// scale returns a new frame of the frame's pixel format, scaled to a size
// with square pixels
func (f *VideoFrame) scale(width, height int) (*VideoFrame, error) {
	ctx := ff.SWScale_get_context(f.Width(), f.Height(), f.PixFmt(), width, height, f.PixFmt(), ff.SWS_BICUBIC, nil, nil, nil)
	if ctx == nil {
		return nil, gomedia.ErrInternalError.Withf("can't scale %v from %dx%d to %dx%d", f.PixFmt(), f.Width(), f.Height(), width, height)
	}
	defer ff.SWScale_free_context(ctx)

	dst, err := NewVideoFrame(f.Stream())
	if err != nil {
		return nil, err
	}
	dst.SetPixFmt(f.PixFmt())
	dst.SetWidth(width)
	dst.SetHeight(height)
	if err := dst.AllocateBuffers(); err != nil {
		dst.Close()
		return nil, err
	}
	if err := ff.SWScale_scale_frame(ctx, dst.AVFrame, f.AVFrame, false); err != nil {
		dst.Close()
		return nil, err
	}
	dst.SetSampleAspectRatio(ff.AVUtil_rational(1, 1))
	return dst, nil
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// alloc makes the frame a pixel format and size, with buffers which can be
// written to, keeping its timebase and pts
func (f *VideoFrame) alloc(pixfmt ff.AVPixelFormat, w, h int) error {
	if f.PixFmt() != pixfmt || f.Width() != w || f.Height() != h {
		timebase, pts := f.TimeBase(), f.Pts()
		ff.AVUtil_frame_unref(f.AVFrame)
		f.SetPixFmt(pixfmt)
		f.SetWidth(w)
		f.SetHeight(h)
		if err := f.AllocateBuffers(); err != nil {
			return err
		}
		f.SetTimeBase(timebase)
		f.SetPts(pts)
	} else if err := f.MakeWritable(); err != nil {
		return err
	}
	f.SetSampleAspectRatio(ff.AVUtil_rational(1, 1))
	return nil
}

// fromPacked copies a packed image, of rows of rowBytes bytes, into the frame
func (f *VideoFrame) fromPacked(pixfmt ff.AVPixelFormat, w, h int, pix []byte, stride, rowBytes int) error {
	if err := f.alloc(pixfmt, w, h); err != nil {
		return err
	}
	copyPlane(f.Bytes(0), f.Linesize(0), pix, stride, rowBytes, h)
	return nil
}

// fromRGBA copies an image with premultiplied alpha into the frame, whose
// alpha isn't premultiplied
func (f *VideoFrame) fromRGBA(src *image.RGBA) error {
	w, h := src.Rect.Dx(), src.Rect.Dy()
	if err := f.alloc(ff.AV_PIX_FMT_RGBA, w, h); err != nil {
		return err
	}
	dst, stride := f.Bytes(0), f.Linesize(0)
	for y := range h {
		s := src.Pix[src.PixOffset(src.Rect.Min.X, src.Rect.Min.Y+y):]
		d := dst[y*stride:]
		for x := range w {
			r, g, b, a := s[4*x], s[4*x+1], s[4*x+2], s[4*x+3]
			if a != 0 && a != 0xFF {
				r, g, b = unpremultiply(r, a), unpremultiply(g, a), unpremultiply(b, a)
			}
			d[4*x], d[4*x+1], d[4*x+2], d[4*x+3] = r, g, b, a
		}
	}
	return nil
}

// fromYCbCr copies a planar YUV image into the frame. When the image's origin
// isn't on a chroma sample's boundary, as for a sub-image, each chroma sample
// is taken from the one at the top left of the pixels it covers in the frame
func (f *VideoFrame) fromYCbCr(pixfmt ff.AVPixelFormat, src *image.YCbCr) error {
	w, h := src.Rect.Dx(), src.Rect.Dy()
	if err := f.alloc(pixfmt, w, h); err != nil {
		return err
	}
	sx, sy := chromaSubsample(src.SubsampleRatio)
	cw, ch := (w+sx-1)/sx, (h+sy-1)/sy
	min := src.Rect.Min
	copyPlane(f.Bytes(0), f.Linesize(0), src.Y[src.YOffset(min.X, min.Y):], src.YStride, w, h)
	if min.X%sx == 0 && min.Y%sy == 0 {
		copyPlane(f.Bytes(1), f.Linesize(1), src.Cb[src.COffset(min.X, min.Y):], src.CStride, cw, ch)
		copyPlane(f.Bytes(2), f.Linesize(2), src.Cr[src.COffset(min.X, min.Y):], src.CStride, cw, ch)
		return nil
	}
	cb, cr, stride := f.Bytes(1), f.Bytes(2), f.Linesize(1)
	for cy := range ch {
		for cx := range cw {
			i := src.COffset(min.X+cx*sx, min.Y+cy*sy)
			cb[cy*stride+cx], cr[cy*stride+cx] = src.Cb[i], src.Cr[i]
		}
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

// cloneImage returns a copy of an image which shares a frame's pixels, which
// is one of those view returns
func cloneImage(img image.Image) image.Image {
	switch img := img.(type) {
	case *image.NRGBA:
		dst := image.NewNRGBA(img.Rect)
		copyPlane(dst.Pix, dst.Stride, img.Pix, img.Stride, 4*img.Rect.Dx(), img.Rect.Dy())
		return dst
	case *image.Gray:
		dst := image.NewGray(img.Rect)
		copyPlane(dst.Pix, dst.Stride, img.Pix, img.Stride, img.Rect.Dx(), img.Rect.Dy())
		return dst
	case *RGB24:
		dst := NewRGB24(img.Rect)
		copyPlane(dst.Pix, dst.Stride, img.Pix, img.Stride, 3*img.Rect.Dx(), img.Rect.Dy())
		return dst
	case *image.YCbCr:
		dst := image.NewYCbCr(img.Rect, img.SubsampleRatio)
		sx, sy := chromaSubsample(img.SubsampleRatio)
		w, h := img.Rect.Dx(), img.Rect.Dy()
		cw, ch := (w+sx-1)/sx, (h+sy-1)/sy
		copyPlane(dst.Y, dst.YStride, img.Y, img.YStride, w, h)
		copyPlane(dst.Cb, dst.CStride, img.Cb, img.CStride, cw, ch)
		copyPlane(dst.Cr, dst.CStride, img.Cr, img.CStride, cw, ch)
		return dst
	default:
		return img
	}
}

// copyPlane copies rows of rowBytes bytes from src to dst, which have their
// own strides
func copyPlane(dst []byte, dstStride int, src []byte, srcStride, rowBytes, rows int) {
	for y := range rows {
		copy(dst[y*dstStride:y*dstStride+rowBytes], src[y*srcStride:y*srcStride+rowBytes])
	}
}

// chromaSubsample returns how many pixels across and down each chroma sample
// of a YUV image covers
func chromaSubsample(ratio image.YCbCrSubsampleRatio) (int, int) {
	switch ratio {
	case image.YCbCrSubsampleRatio422:
		return 2, 1
	case image.YCbCrSubsampleRatio420:
		return 2, 2
	case image.YCbCrSubsampleRatio440:
		return 1, 2
	case image.YCbCrSubsampleRatio411:
		return 4, 1
	default:
		return 1, 1
	}
}

// unpremultiply returns a colour component without its alpha premultiplied,
// rounded to the nearest value
func unpremultiply(c, a uint8) uint8 {
	return uint8(min(0xFF, (uint32(c)*0xFF+uint32(a)/2)/uint32(a)))
}
