package generator

import (
	"encoding/json"
	"errors"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

type yuv420p struct {
	frame *ff.AVFrame
}

var _ Generator = (*yuv420p)(nil)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Video generator - YUV420P
func NewYUV420P(size string, framerate int) (*yuv420p, error) {
	yuv420p := new(yuv420p)

	// Check parameters
	if framerate <= 0 {
		return nil, errors.New("invalid framerate")
	}
	w, h, err := ff.AVUtil_parse_video_size(size)
	if err != nil {
		return nil, err
	}

	// Create a frame
	frame := ff.AVUtil_frame_alloc()
	if frame == nil {
		return nil, errors.New("failed to allocate frame")
	}

	frame.SetPixFmt(ff.AV_PIX_FMT_YUV420P)
	frame.SetWidth(w)
	frame.SetHeight(h)
	frame.SetSampleAspectRatio(ff.AVUtil_rational(1, 1))
	frame.SetTimeBase(ff.AVUtil_rational(1, framerate))
	frame.SetPts(ff.AV_NOPTS_VALUE)

	// Allocate buffer
	if err := ff.AVUtil_frame_get_buffer(frame, false); err != nil {
		return nil, err
	} else {
		yuv420p.frame = frame
	}

	// Return success
	return yuv420p, nil
}

// Free resources
func (yuv420p *yuv420p) Close() error {
	ff.AVUtil_frame_free(yuv420p.frame)
	yuv420p.frame = nil
	return nil
}

////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (yuv420p *yuv420p) String() string {
	data, _ := json.MarshalIndent(yuv420p.frame, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (yuv420p *yuv420p) Frame() media.Frame {
	// Set the Pts
	if yuv420p.frame.Pts() == ff.AV_NOPTS_VALUE {
		yuv420p.frame.SetPts(0)
	} else {
		yuv420p.frame.SetPts(yuv420p.frame.Pts() + 1)
	}

	/* Y */
	n := int(yuv420p.frame.Pts())
	yplane := yuv420p.frame.Bytes(0)
	ystride := yuv420p.frame.Linesize(0)
	for y := 0; y < yuv420p.frame.Height(); y++ {
		for x := 0; x < yuv420p.frame.Width(); x++ {
			yplane[y*ystride+x] = byte(x + y + n*3)
		}
	}

	/* Cb and Cr */
	cbplane := yuv420p.frame.Bytes(1)
	crplane := yuv420p.frame.Bytes(2)
	cstride := yuv420p.frame.Linesize(1)
	for y := 0; y < yuv420p.frame.Height()>>1; y++ {
		for x := 0; x < yuv420p.frame.Width()>>1; x++ {
			cbplane[y*cstride+x] = byte(128 + y + n*2)
			crplane[y*cstride+x] = byte(64 + x + n*5)
		}
	}

	// Return the frame
	return ffmpeg.NewFrame(yuv420p.frame)
}
