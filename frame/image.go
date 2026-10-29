package frame

import (
	"image"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

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

// Image returns an image of the frame, which shares the frame's pixels, so
// it's only valid until the frame is closed or reused. The pixel format is
// one of RGBA, GRAY8, RGB24, or a planar YUV format (YUV411P, YUV420P,
// YUV422P, YUV440P or YUV444P); convert other formats first, such as with a
// writer.Resampler.
func (f *VideoFrame) Image() (image.Image, error) {
	rect := image.Rect(0, 0, f.Width(), f.Height())
	switch f.PixFmt() {
	case ff.AV_PIX_FMT_RGBA:
		return &image.RGBA{Pix: f.Bytes(0), Stride: f.Linesize(0), Rect: rect}, nil
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
// is one of *image.RGBA or *image.NRGBA (to RGBA), *image.Gray (to GRAY8),
// *RGB24 (to RGB24), or *image.YCbCr with any subsample ratio except 4:1:0
// (to planar YUV).
func (f *VideoFrame) FromImage(src image.Image) error {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	min := src.Bounds().Min
	switch src := src.(type) {
	case *image.RGBA:
		return f.fromPacked(ff.AV_PIX_FMT_RGBA, w, h, src.Pix[src.PixOffset(min.X, min.Y):], src.Stride, 4*w)
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

// fromYCbCr copies a planar YUV image into the frame
func (f *VideoFrame) fromYCbCr(pixfmt ff.AVPixelFormat, src *image.YCbCr) error {
	w, h := src.Rect.Dx(), src.Rect.Dy()
	if err := f.alloc(pixfmt, w, h); err != nil {
		return err
	}
	cw, ch := chromaSize(src.SubsampleRatio, w, h)
	min := src.Rect.Min
	copyPlane(f.Bytes(0), f.Linesize(0), src.Y[src.YOffset(min.X, min.Y):], src.YStride, w, h)
	copyPlane(f.Bytes(1), f.Linesize(1), src.Cb[src.COffset(min.X, min.Y):], src.CStride, cw, ch)
	copyPlane(f.Bytes(2), f.Linesize(2), src.Cr[src.COffset(min.X, min.Y):], src.CStride, cw, ch)
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

// copyPlane copies rows of rowBytes bytes from src to dst, which have their
// own strides
func copyPlane(dst []byte, dstStride int, src []byte, srcStride, rowBytes, rows int) {
	for y := range rows {
		copy(dst[y*dstStride:y*dstStride+rowBytes], src[y*srcStride:y*srcStride+rowBytes])
	}
}

// chromaSize returns the width and height of the chroma planes of a YUV image
// of width w and height h
func chromaSize(ratio image.YCbCrSubsampleRatio, w, h int) (int, int) {
	switch ratio {
	case image.YCbCrSubsampleRatio422:
		return (w + 1) / 2, h
	case image.YCbCrSubsampleRatio420:
		return (w + 1) / 2, (h + 1) / 2
	case image.YCbCrSubsampleRatio440:
		return w, (h + 1) / 2
	case image.YCbCrSubsampleRatio411:
		return (w + 3) / 4, h
	default:
		return w, h
	}
}
