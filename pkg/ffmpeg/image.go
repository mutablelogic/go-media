package ffmpeg

import (
	"image"

	// Packages
	media "github.com/mutablelogic/go-media"
	imagex "github.com/mutablelogic/go-media/pkg/image"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	yuvSubsampleRatio = map[ff.AVPixelFormat]image.YCbCrSubsampleRatio{
		ff.AV_PIX_FMT_YUV410P: image.YCbCrSubsampleRatio410,
		ff.AV_PIX_FMT_YUV411P: image.YCbCrSubsampleRatio411,
		ff.AV_PIX_FMT_YUV420P: image.YCbCrSubsampleRatio420,
		ff.AV_PIX_FMT_YUV422P: image.YCbCrSubsampleRatio422,
		ff.AV_PIX_FMT_YUV440P: image.YCbCrSubsampleRatio420,
		ff.AV_PIX_FMT_YUV444P: image.YCbCrSubsampleRatio444,
	}
	pixfmtYCbCr = map[image.YCbCrSubsampleRatio]ff.AVPixelFormat{
		image.YCbCrSubsampleRatio410: ff.AV_PIX_FMT_YUV410P,
		image.YCbCrSubsampleRatio411: ff.AV_PIX_FMT_YUV411P,
		image.YCbCrSubsampleRatio420: ff.AV_PIX_FMT_YUV420P,
		image.YCbCrSubsampleRatio422: ff.AV_PIX_FMT_YUV422P,
		image.YCbCrSubsampleRatio440: ff.AV_PIX_FMT_YUV440P,
		image.YCbCrSubsampleRatio444: ff.AV_PIX_FMT_YUV444P,
	}
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Repurpose a frame and copy image into it.
// TODO: Support SampleAspectRatio?
func (frame *Frame) FromImage(src image.Image) error {
	switch src := src.(type) {
	case *image.RGBA: // AV_PIX_FMT_RGBA
		return frame.fromRGBA(src, ff.AV_PIX_FMT_RGBA)
	case *image.NRGBA: // AV_PIX_FMT_RGBA
		return frame.fromNRGBA(src, ff.AV_PIX_FMT_RGBA)
	case *image.Gray: // AV_PIX_FMT_GRAY8
		return frame.fromGray8(src, ff.AV_PIX_FMT_GRAY8)
	case *imagex.RGB24: // AV_PIX_FMT_RGB24
		return frame.fromRGB24(src, ff.AV_PIX_FMT_RGB24)
	case *image.YCbCr: // Planar YUV formats
		if pixfmt, exists := pixfmtYCbCr[src.SubsampleRatio]; exists {
			return frame.fromYUVP(src, pixfmt)
		}
	}
	return ErrNotImplemented.Withf("unsupported image format: %T", src)
}

// Create an image from a frame. The frame should not be unreferenced
// until the image is no longer required, but the image can be discarded
// TODO: Add a copy flag which copies the memory?
func (frame *Frame) Image() (image.Image, error) {
	if frame.Type() != media.VIDEO {
		return nil, ErrBadParameter.With("unsupported frame type: ", frame.Type())
	}
	switch frame.PixelFormat() {
	case ff.AV_PIX_FMT_RGBA:
		return &image.RGBA{
			Pix:    frame.Bytes(0),
			Stride: frame.Stride(0),
			Rect:   image.Rect(0, 0, frame.Width(), frame.Height()),
		}, nil
	case ff.AV_PIX_FMT_GRAY8:
		return &image.Gray{
			Pix:    frame.Bytes(0),
			Stride: frame.Stride(0),
			Rect:   image.Rect(0, 0, frame.Width(), frame.Height()),
		}, nil
	case ff.AV_PIX_FMT_RGB24:
		return &imagex.RGB24{
			Pix:    frame.Bytes(0),
			Stride: frame.Stride(0),
			Rect:   image.Rect(0, 0, frame.Width(), frame.Height()),
		}, nil
	}

	// Check planar yuv formats
	if ratio, exists := yuvSubsampleRatio[frame.PixelFormat()]; exists {
		return &image.YCbCr{
			Y:              frame.Bytes(0),
			Cb:             frame.Bytes(1),
			Cr:             frame.Bytes(2),
			YStride:        frame.Stride(0),
			CStride:        frame.Stride(1),
			SubsampleRatio: ratio,
			Rect:           image.Rect(0, 0, frame.Width(), frame.Height()),
		}, nil
	}

	return nil, ErrNotImplemented.With("unsupported pixel format: ", frame.PixelFormat())
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (frame *Frame) fromImage(pixfmt ff.AVPixelFormat, w, h int, sar ff.AVRational) error {
	// Save timing parameters
	timeBase := frame.TimeBase()
	pts := frame.Pts()

	if frame.PixelFormat() != pixfmt || frame.Width() != w || frame.Height() != h {
		// Clear frame
		frame.Unref()

		// Allocate buffers for the frame
		((*ff.AVFrame)(frame)).SetPixFmt(pixfmt)
		((*ff.AVFrame)(frame)).SetWidth(w)
		((*ff.AVFrame)(frame)).SetHeight(h)
		if err := frame.AllocateBuffers(); err != nil {
			return err
		}
	}

	// Copy over timing parameters, etc.
	((*ff.AVFrame)(frame)).SetSampleAspectRatio(sar)
	((*ff.AVFrame)(frame)).SetTimeBase(timeBase)
	((*ff.AVFrame)(frame)).SetPts(pts)

	// Return success
	return nil
}

func (frame *Frame) fromNRGBA(src *image.NRGBA, pixfmt ff.AVPixelFormat) error {
	// Create a new frame if the pixel format or size is different
	if err := frame.fromImage(pixfmt, src.Bounds().Dx(), src.Bounds().Dy(), ff.AVUtil_rational(1, 1)); err != nil {
		return err
	}
	if src.Stride == frame.Stride(0) {
		copy(frame.Bytes(0), src.Pix)
	} else {
		for y := 0; y < src.Rect.Dy(); y++ {
			copy(frame.Bytes(0)[y*frame.Stride(0):], src.Pix[y*src.Stride:])
		}
	}

	// Return success
	return nil
}

func (frame *Frame) fromRGBA(src *image.RGBA, pixfmt ff.AVPixelFormat) error {
	if err := frame.fromImage(pixfmt, src.Bounds().Dx(), src.Bounds().Dy(), ff.AVUtil_rational(1, 1)); err != nil {
		return err
	}
	if src.Stride == frame.Stride(0) {
		copy(frame.Bytes(0), src.Pix)
	} else {
		for y := 0; y < src.Rect.Dy(); y++ {
			copy(frame.Bytes(0)[y*frame.Stride(0):], src.Pix[y*src.Stride:])
		}
	}

	// Return success
	return nil
}

func (frame *Frame) fromGray8(src *image.Gray, pixfmt ff.AVPixelFormat) error {
	if err := frame.fromImage(pixfmt, src.Bounds().Dx(), src.Bounds().Dy(), ff.AVUtil_rational(1, 1)); err != nil {
		return err
	}
	if src.Stride == frame.Stride(0) {
		copy(frame.Bytes(0), src.Pix)
	} else {
		for y := 0; y < src.Rect.Dy(); y++ {
			copy(frame.Bytes(0)[y*frame.Stride(0):], src.Pix[y*src.Stride:])
		}
	}

	// Return success
	return nil
}

func (frame *Frame) fromRGB24(src *imagex.RGB24, pixfmt ff.AVPixelFormat) error {
	if err := frame.fromImage(pixfmt, src.Bounds().Dx(), src.Bounds().Dy(), ff.AVUtil_rational(1, 1)); err != nil {
		return err
	}
	if src.Stride == frame.Stride(0) {
		copy(frame.Bytes(0), src.Pix)
	} else {
		for y := 0; y < src.Rect.Dy(); y++ {
			copy(frame.Bytes(0)[y*frame.Stride(0):], src.Pix[y*src.Stride:])
		}
	}

	// Return success
	return nil
}

func (frame *Frame) fromYUVP(src *image.YCbCr, pixfmt ff.AVPixelFormat) error {
	if err := frame.fromImage(pixfmt, src.Bounds().Dx(), src.Bounds().Dy(), ff.AVUtil_rational(1, 1)); err != nil {
		return err
	}
	if src.YStride == frame.Stride(0) && src.CStride == frame.Stride(1) && src.CStride == frame.Stride(2) {
		copy(frame.Bytes(0), src.Y)
		copy(frame.Bytes(1), src.Cb)
		copy(frame.Bytes(2), src.Cr)
	} else {
		for y := 0; y < src.Rect.Dy(); y++ {
			copy(frame.Bytes(0)[y*frame.Stride(0):], src.Y[y*src.YStride:])
			if y < src.Rect.Dy()/2 {
				copy(frame.Bytes(1)[y*frame.Stride(1):], src.Cb[y*src.CStride:])
				copy(frame.Bytes(2)[y*frame.Stride(2):], src.Cr[y*src.CStride:])
			}
		}
	}

	// Return success
	return nil
}
