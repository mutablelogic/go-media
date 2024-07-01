package ffmpeg

import (
	"encoding/json"
	"image"
	"time"

	// Packages
	media "github.com/mutablelogic/go-media"
	imagex "github.com/mutablelogic/go-media/pkg/image"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Frame struct {
	ctx    *ff.AVFrame
	stream int
}

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

func NewFrame(ctx *ff.AVFrame, stream int) *Frame {
	return &Frame{ctx, stream}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (frame *Frame) MarshalJSON() ([]byte, error) {
	return json.Marshal(frame.ctx)
}

func (frame *Frame) String() string {
	data, _ := json.MarshalIndent(frame, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////////
// PARAMETERS

// Return the context
func (frame *Frame) AVFrame() *ff.AVFrame {
	return frame.ctx
}

// Return the media type (AUDIO, VIDEO)
func (frame *Frame) Type() media.MediaType {
	if frame.ctx.NumSamples() > 0 {
		return media.AUDIO
	}
	if frame.ctx.Width() != 0 && frame.ctx.Height() != 0 {
		return media.VIDEO
	}
	return media.NONE
}

// Return the stream
func (frame *Frame) Id() int {
	return frame.stream
}

// Return the timestamp as a duration, or minus one if not set
func (frame *Frame) Time() time.Duration {
	pts := frame.ctx.Pts()
	if pts == ff.AV_NOPTS_VALUE {
		return -1
	}
	if frame.ctx.TimeBase().Den() == 0 {
		return -1
	}
	return secondsToDuration(float64(pts) * ff.AVUtil_rational_q2d(frame.ctx.TimeBase()))
}

// Return the number of planes for a specific PixelFormat
// or SampleFormat and ChannelLayout combination
func (frame *Frame) NumPlanes() int {
	return ff.AVUtil_frame_get_num_planes(frame.ctx)
}

// Return the byte data for a plane
func (frame *Frame) Bytes(plane int) []byte {
	return frame.ctx.Bytes(plane)[:frame.ctx.Planesize(plane)]
}

// Return the int16 data for a plane
func (frame *Frame) Int16(plane int) []int16 {
	sz := frame.ctx.Planesize(plane) >> 1
	return frame.ctx.Int16(plane)[:sz]
}

////////////////////////////////////////////////////////////////////////////////
// AUDIO PARAMETERS

// Return number of samples
func (frame *Frame) NumSamples() int {
	if frame.Type() != media.AUDIO {
		return 0
	}
	return frame.ctx.NumSamples()
}

// Return channel layout
func (frame *Frame) ChannelLayout() string {
	if frame.Type() != media.AUDIO {
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
func (frame *Frame) SampleFormat() string {
	if frame.Type() != media.AUDIO {
		return ""
	}
	return ff.AVUtil_get_sample_fmt_name(frame.ctx.SampleFormat())
}

// Return the sample rate (Hz)
func (frame *Frame) Samplerate() int {
	if frame.Type() != media.AUDIO {
		return 0
	}
	return frame.ctx.SampleRate()

}

////////////////////////////////////////////////////////////////////////////////
// VIDEO PARAMETERS

// Convert a frame into an image
func (frame *Frame) Image() (image.Image, error) {
	if t := frame.Type(); t != media.VIDEO {
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
func (frame *Frame) Stride(plane int) int {
	if frame.Type() == media.VIDEO {
		return frame.ctx.Linesize(plane)
	} else {
		return 0
	}
}

// Return the width of the video frame
func (frame *Frame) Width() int {
	if frame.Type() != media.VIDEO {
		return 0
	}
	return frame.ctx.Width()
}

// Return the height of the video frame
func (frame *Frame) Height() int {
	if frame.Type() != media.VIDEO {
		return 0
	}
	return frame.ctx.Height()
}

// Return the pixel format
func (frame *Frame) PixelFormat() string {
	if frame.Type() != media.VIDEO {
		return ""
	}
	return ff.AVUtil_get_pix_fmt_name(frame.ctx.PixFmt())
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func secondsToDuration(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}
