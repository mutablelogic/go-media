package ffmpeg

import (
	"errors"
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

// Create a frame from an image. Creates a new Frame, which must
// be unreferenced after use and closed to release resources.
// TODO: Support SampleAspectRatio
func FrameFromImage(src image.Image) (*Frame, error) {
	switch src := src.(type) {
	case *image.Gray: // AV_PIX_FMT_GRAY8
		return newGray8(src)
	case *image.NRGBA: // AV_PIX_FMT_RGBA
		return newRGBA(src)
	case *imagex.RGB24: // AV_PIX_FMT_RGB24
		return newRGB24(src)
	case *image.YCbCr: // Planar YUV formats
		if pixfmt, exists := pixfmtYCbCr[src.SubsampleRatio]; exists {
			return newYUVP(src, pixfmt)
		}
	}
	return nil, ErrNotImplemented.Withf("unsupported image format: %T", src)
}

// Create an image from a frame. The frame should not be unreferenced
// until the image is no longer required, but the image can be discarded
// TODO: Add a copy flag which copies the memory?
func (frame *Frame) ImageFromFrame() (image.Image, error) {
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

func newRGBA(src *image.NRGBA) (*Frame, error) {
	par := new(Par)
	par.SetPixelFormat(ff.AV_PIX_FMT_RGBA)
	par.SetWidth(src.Rect.Dx())
	par.SetHeight(src.Rect.Dy())
	par.SetSampleAspectRatio(ff.AVUtil_rational(1, 1))

	// Make the frame and allocate buffers
	frame, err := NewFrame(par)
	if err != nil {
		return nil, err
	}
	if err := frame.AllocateBuffers(); err != nil {
		return nil, errors.Join(err, frame.Close())
	}
	if src.Stride == frame.Stride(0) {
		copy(frame.Bytes(0), src.Pix)
	} else {
		for y := 0; y < src.Rect.Dy(); y++ {
			copy(frame.Bytes(0)[y*frame.Stride(0):], src.Pix[y*src.Stride:])
		}
	}

	// Return success
	return frame, nil
}

func newGray8(src *image.Gray) (*Frame, error) {
	par := new(Par)
	par.SetPixelFormat(ff.AV_PIX_FMT_GRAY8)
	par.SetWidth(src.Rect.Dx())
	par.SetHeight(src.Rect.Dy())
	par.SetSampleAspectRatio(ff.AVUtil_rational(1, 1))

	// Make the frame and allocate buffers
	frame, err := NewFrame(par)
	if err != nil {
		return nil, err
	}
	if err := frame.AllocateBuffers(); err != nil {
		return nil, errors.Join(err, frame.Close())
	}
	if src.Stride == frame.Stride(0) {
		copy(frame.Bytes(0), src.Pix)
	} else {
		for y := 0; y < src.Rect.Dy(); y++ {
			copy(frame.Bytes(0)[y*frame.Stride(0):], src.Pix[y*src.Stride:])
		}
	}

	// Return success
	return frame, nil
}

func newRGB24(src *imagex.RGB24) (*Frame, error) {
	par := new(Par)
	par.SetPixelFormat(ff.AV_PIX_FMT_RGB24)
	par.SetWidth(src.Rect.Dx())
	par.SetHeight(src.Rect.Dy())
	par.SetSampleAspectRatio(ff.AVUtil_rational(1, 1))

	// Make the frame and allocate buffers
	frame, err := NewFrame(par)
	if err != nil {
		return nil, err
	}
	if err := frame.AllocateBuffers(); err != nil {
		return nil, errors.Join(err, frame.Close())
	}
	if src.Stride == frame.Stride(0) {
		copy(frame.Bytes(0), src.Pix)
	} else {
		for y := 0; y < src.Rect.Dy(); y++ {
			copy(frame.Bytes(0)[y*frame.Stride(0):], src.Pix[y*src.Stride:])
		}
	}

	// Return success
	return frame, nil
}

func newYUVP(src *image.YCbCr, pixfmt ff.AVPixelFormat) (*Frame, error) {
	par := new(Par)
	par.SetPixelFormat(pixfmt)
	par.SetWidth(src.Rect.Dx())
	par.SetHeight(src.Rect.Dy())
	par.SetSampleAspectRatio(ff.AVUtil_rational(1, 1))

	// Make the frame and allocate buffers
	frame, err := NewFrame(par)
	if err != nil {
		return nil, err
	}
	if err := frame.AllocateBuffers(); err != nil {
		return nil, errors.Join(err, frame.Close())
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
	return frame, nil
}
