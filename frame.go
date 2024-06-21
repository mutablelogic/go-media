package media

import (
	"encoding/json"
	"image"

	// Packages
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

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newFrame(ctx *ff.AVFrame) *frame {
	return &frame{ctx}
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (frame *frame) MarshalJSON() ([]byte, error) {
	type jsonFrame struct {
		Type MediaType `json:"type"`
		*audiopar
		*videopar
		*planepar
		*timingpar
	}
	if frame.Type() == AUDIO {
		return json.Marshal(jsonFrame{
			Type: frame.Type(),
			audiopar: &audiopar{
				Ch:           frame.ctx.ChannelLayout(),
				SampleFormat: frame.ctx.SampleFormat(),
				Samplerate:   frame.ctx.SampleRate(),
			},
			planepar: &planepar{
				NumPlanes: ff.AVUtil_frame_get_num_planes(frame.ctx),
			},
			timingpar: &timingpar{
				Pts:      frame.ctx.Pts(),
				TimeBase: frame.ctx.TimeBase(),
			},
		})
	} else if frame.Type() == VIDEO {
		return json.Marshal(jsonFrame{
			Type: frame.Type(),
			videopar: &videopar{
				PixelFormat: frame.ctx.PixFmt(),
				Width:       frame.ctx.Width(),
				Height:      frame.ctx.Height(),
			},
			planepar: &planepar{
				NumPlanes: ff.AVUtil_frame_get_num_planes(frame.ctx),
			},
		})
	} else {
		return json.Marshal(jsonFrame{
			Type: frame.Type(),
		})
	}
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

// Return the number of planes for a specific PixelFormat
// or SampleFormat and ChannelLayout combination
func (frame *frame) NumPlanes() int {
	return ff.AVUtil_frame_get_num_planes(frame.ctx)
}

// Return the byte data for a plane
func (frame *frame) Bytes(plane int) []byte {
	return frame.ctx.Bytes(plane)
}

// Return the number of bytes in a single row of the video frame
func (frame *frame) Stride(plane int) int {
	if frame.Type() == VIDEO {
		return frame.ctx.Linesize(plane)
	} else {
		return 0
	}
}

// Convert a frame into an image
func (frame *frame) Image() (image.Image, error) {
	if t := frame.Type(); t != VIDEO {
		return nil, ErrBadParameter.With("unsupported frame type", t)
	}
	switch frame.ctx.PixFmt() {
	case ff.AV_PIX_FMT_YUV420P:
		return &image.YCbCr{
			Y:              frame.Bytes(0),
			Cb:             frame.Bytes(1),
			Cr:             frame.Bytes(2),
			YStride:        frame.Stride(0),
			CStride:        frame.Stride(1),
			SubsampleRatio: image.YCbCrSubsampleRatio420,
			Rect:           image.Rect(0, 0, frame.Width(), frame.Height()),
		}, nil
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
		return &image.NRGBA{
			Pix:    frame.Bytes(0),
			Stride: frame.Stride(0),
			Rect:   image.Rect(0, 0, frame.Width(), frame.Height()),
		}, nil
	}
	return nil, ErrNotImplemented.With("unsupported pixel format", frame.ctx.PixFmt())
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
