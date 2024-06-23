package media

import (
	"encoding/json"
	"image"
	"time"

	// Packages
	imagex "github.com/mutablelogic/go-media/pkg/image"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type frame struct {
	ctx *ff.AVFrame
}

var _ Frame = (*frame)(nil)

var (
	yuvSubsampleRatio = map[ff.AVPixelFormat]image.YCbCrSubsampleRatio{
		ff.AV_PIX_FMT_YUV410P: image.YCbCrSubsampleRatio410,
		ff.AV_PIX_FMT_YUV411P: image.YCbCrSubsampleRatio410,
		ff.AV_PIX_FMT_YUV420P: image.YCbCrSubsampleRatio420,
		ff.AV_PIX_FMT_YUV422P: image.YCbCrSubsampleRatio422,
		ff.AV_PIX_FMT_YUV440P: image.YCbCrSubsampleRatio420,
		ff.AV_PIX_FMT_YUV444P: image.YCbCrSubsampleRatio444,
	}
	yuvaSubsampleRatio = map[ff.AVPixelFormat]image.YCbCrSubsampleRatio{
		ff.AV_PIX_FMT_YUVA420P: image.YCbCrSubsampleRatio420,
		ff.AV_PIX_FMT_YUVA422P: image.YCbCrSubsampleRatio422,
		ff.AV_PIX_FMT_YUVA444P: image.YCbCrSubsampleRatio444,
	}
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newFrame(ctx *ff.AVFrame) *frame {
	return &frame{ctx}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (frame *frame) MarshalJSON() ([]byte, error) {
	return json.Marshal(frame.ctx)
}

func (frame *frame) String() string {
	data, _ := json.MarshalIndent(frame, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PARAMETERS

// Return the media type (AUDIO, VIDEO)
func (frame *frame) Type() MediaType {
	if frame.ctx.NumSamples() > 0 {
		return AUDIO
	}
	if frame.ctx.Width() != 0 && frame.ctx.Height() != 0 {
		return VIDEO
	}
	return NONE
}

// Return the timestamp as a duration, or minus one if not set
func (frame *frame) Time() time.Duration {
	pts := frame.ctx.Pts()
	if pts == ff.AV_NOPTS_VALUE {
		return -1
	}
	return secondsToDuration(float64(pts) * ff.AVUtil_q2d(frame.ctx.TimeBase()))
}

// Return the number of planes for a specific PixelFormat
// or SampleFormat and ChannelLayout combination
func (frame *frame) NumPlanes() int {
	return ff.AVUtil_frame_get_num_planes(frame.ctx)
}

// Return the byte data for a plane
func (frame *frame) Bytes(plane int) []byte {
	return frame.ctx.Bytes(plane)[:frame.ctx.Planesize(plane)]
}

////////////////////////////////////////////////////////////////////////////////
// AUDIO PARAMETERS

// Return number of samples
func (frame *frame) NumSamples() int {
	if frame.Type() != AUDIO {
		return 0
	}
	return frame.ctx.NumSamples()
}

// Return channel layout
func (frame *frame) ChannelLayout() string {
	if frame.Type() != AUDIO {
		return ""
	}
	ch := frame.ctx.ChannelLayout()
	if name, err := ff.AVUtil_channel_layout_describe(&ch); err != nil {
		return ""
	} else {
		return name
	}
}

// Return the sample format
func (frame *frame) SampleFormat() string {
	if frame.Type() != AUDIO {
		return ""
	}
	return ff.AVUtil_get_sample_fmt_name(frame.ctx.SampleFormat())
}

// Return the sample rate (Hz)
func (frame *frame) Samplerate() int {
	if frame.Type() != AUDIO {
		return 0
	}
	return frame.ctx.SampleRate()

}

////////////////////////////////////////////////////////////////////////////////
// VIDEO PARAMETERS

// Convert a frame into an image
func (frame *frame) Image() (image.Image, error) {
	if t := frame.Type(); t != VIDEO {
		return nil, ErrBadParameter.With("unsupported frame type", t)
	}
	pixel_format := frame.ctx.PixFmt()
	switch pixel_format {
	case ff.AV_PIX_FMT_GRAY8:
		return &image.Gray{
			Pix:    frame.Bytes(0),
			Stride: frame.Stride(0),
			Rect:   image.Rect(0, 0, frame.Width(), frame.Height()),
		}, nil
	case ff.AV_PIX_FMT_RGBA:
		return &image.RGBA{
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
	default:
		if ratio, exists := yuvSubsampleRatio[pixel_format]; exists {
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
		if ratio, exists := yuvaSubsampleRatio[pixel_format]; exists {
			return &image.NYCbCrA{
				YCbCr: image.YCbCr{
					Y:              frame.Bytes(0),
					Cb:             frame.Bytes(1),
					Cr:             frame.Bytes(2),
					YStride:        frame.Stride(0),
					CStride:        frame.Stride(1),
					SubsampleRatio: ratio,
					Rect:           image.Rect(0, 0, frame.Width(), frame.Height()),
				},
				A:       frame.Bytes(3),
				AStride: frame.Stride(3),
			}, nil
		}
	}
	return nil, ErrNotImplemented.With("unsupported pixel format", frame.ctx.PixFmt())
}

// Return the number of bytes in a single row of the video frame
func (frame *frame) Stride(plane int) int {
	if frame.Type() == VIDEO {
		return frame.ctx.Linesize(plane)
	} else {
		return 0
	}
}

// Return the width of the video frame
func (frame *frame) Width() int {
	if frame.Type() != VIDEO {
		return 0
	}
	return frame.ctx.Width()
}

// Return the height of the video frame
func (frame *frame) Height() int {
	if frame.Type() != VIDEO {
		return 0
	}
	return frame.ctx.Height()
}

// Return the pixel format
func (frame *frame) PixelFormat() string {
	if frame.Type() != VIDEO {
		return ""
	}
	return ff.AVUtil_get_pix_fmt_name(frame.ctx.PixFmt())
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func secondsToDuration(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}
