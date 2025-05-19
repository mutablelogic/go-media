package generator

import (
	"encoding/json"
	"errors"

	// Packages
	media "github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

type yuv420p struct {
	frame *ffmpeg.Frame
}

var _ Generator = (*yuv420p)(nil)

////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new video generator which generates YUV420P frames
// of the specified size and framerate (in frames per second)
func NewYUV420P(par *ffmpeg.Par) (*yuv420p, error) {
	yuv420p := new(yuv420p)

	// Check parameters
	if par.Type() != media.VIDEO {
		return nil, errors.New("invalid codec type")
	} else if par.PixelFormat() != ff.AV_PIX_FMT_YUV420P {
		return nil, errors.New("invalid pixel format, only yuv420p is supported")
	}
	if framerate := par.FrameRate(); framerate <= 0 {
		return nil, errors.New("invalid framerate")
	}

	// Create a frame
	frame, err := ffmpeg.NewFrame(par)
	if err != nil {
		return nil, err
	}

	// Allocate buffer
	if err := frame.AllocateBuffers(); err != nil {
		return nil, errors.Join(err, frame.Close())
	}

	// Set parameters
	yuv420p.frame = frame

	// Return success
	return yuv420p, nil
}

// Free resources for the generator
func (yuv420p *yuv420p) Close() error {
	result := yuv420p.frame.Close()
	yuv420p.frame = nil
	return result
}

////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (yuv420p *yuv420p) String() string {
	data, _ := json.MarshalIndent(yuv420p.frame, "", "  ")
	return string(data)
}

////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return the first and subsequent frames of raw video data
func (yuv420p *yuv420p) Frame() *ffmpeg.Frame {
	// Set the Pts
	if yuv420p.frame.Pts() == ffmpeg.PTS_UNDEFINED {
		yuv420p.frame.SetPts(0)
	} else {
		// Increment by one frame
		yuv420p.frame.IncPts(1)
	}

	/* Y */
	n := int(yuv420p.frame.Pts())
	yplane := yuv420p.frame.Bytes(0)
	ystride := yuv420p.frame.Stride(0)
	for y := 0; y < yuv420p.frame.Height(); y++ {
		for x := 0; x < yuv420p.frame.Width(); x++ {
			yplane[y*ystride+x] = byte(x + y + n*3)
		}
	}

	/* Cb and Cr */
	cbplane := yuv420p.frame.Bytes(1)
	crplane := yuv420p.frame.Bytes(2)
	cstride := yuv420p.frame.Stride(1)
	for y := 0; y < yuv420p.frame.Height()>>1; y++ {
		for x := 0; x < yuv420p.frame.Width()>>1; x++ {
			cbplane[y*cstride+x] = byte(128 + y + n*2)
			crplane[y*cstride+x] = byte(64 + x + n*5)
		}
	}

	// Return the frame
	return yuv420p.frame
}
