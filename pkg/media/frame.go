package media

import (

	// Packages
	"fmt"
	"time"

	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	//. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type frame struct {
	ctx *ffmpeg.AVFrame
}

// Ensure *input complies with Media interface
var _ Frame = (*frame)(nil)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFrame() *frame {
	frame := new(frame)

	if ctx := ffmpeg.AVUtil_frame_alloc(); ctx == nil {
		return nil
	} else {
		frame.ctx = ctx
	}

	// Return success
	return frame
}

func (frame *frame) Close() error {
	var result error

	// Callback
	if frame.ctx != nil {
		ffmpeg.AVUtil_frame_free_ptr(frame.ctx)
		frame.ctx = nil
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (frame *frame) String() string {
	str := "<media.frame"
	flags := frame.Flags()
	if flags != MEDIA_FLAG_NONE {
		str += fmt.Sprint(" flags=", flags)
	}
	if flags.Is(MEDIA_FLAG_AUDIO) {
		if audio_format := frame.AudioFormat(); audio_format.Rate > 0 {
			str += fmt.Sprint(" format=", frame.AudioFormat())
		}
		if samples := frame.NumSamples(); samples > 0 {
			str += fmt.Sprint(" nb_samples=", samples)
		}
		if channels := frame.Channels(); len(channels) > 0 {
			str += fmt.Sprint(" channels=", channels)
		}
		if duration := frame.Duration(); duration > 0 {
			str += fmt.Sprint(" duration=", duration)
		}
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Unref releases the packet
func (frame *frame) Release() {
	if frame.ctx != nil {
		ffmpeg.AVUtil_frame_unref(frame.ctx)
	}
}

// Flags
func (frame *frame) Flags() MediaFlag {
	flags := MEDIA_FLAG_NONE
	if frame.ctx == nil {
		return flags
	}
	if frame.ctx.PixelFormat() != ffmpeg.AV_PIX_FMT_NONE {
		flags |= MEDIA_FLAG_VIDEO
	}
	if frame.ctx.SampleFormat() != ffmpeg.AV_SAMPLE_FMT_NONE {
		flags |= MEDIA_FLAG_AUDIO
	}
	return flags
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS: AudioFrame

// Returns the audio format, if MEDIA_FLAG_AUDIO is set
func (frame *frame) AudioFormat() AudioFormat {
	if frame.ctx == nil || frame.ctx.SampleFormat() == ffmpeg.AV_SAMPLE_FMT_NONE {
		return AudioFormat{}
	}
	return AudioFormat{
		Rate:   uint(frame.ctx.SampleRate()),
		Format: fromSampleFormat(frame.ctx.SampleFormat()),
	}
}

// Number of samples, if MEDIA_FLAG_AUDIO is set
func (frame *frame) NumSamples() int {
	if frame.ctx == nil || frame.ctx.SampleFormat() == ffmpeg.AV_SAMPLE_FMT_NONE {
		return 0
	}
	return frame.ctx.NumSamples()
}

// Audio channels, if MEDIA_FLAG_AUDIO is set
func (frame *frame) Channels() []AudioChannel {
	if frame.ctx == nil || frame.ctx.SampleFormat() == ffmpeg.AV_SAMPLE_FMT_NONE {
		return nil
	}
	// TODO
	return []AudioChannel{}
}

// Duration of the frame, if MEDIA_FLAG_AUDIO is set
func (frame *frame) Duration() time.Duration {
	if frame.ctx == nil || frame.ctx.SampleFormat() == ffmpeg.AV_SAMPLE_FMT_NONE {
		return 0
	}
	return time.Second * time.Duration(frame.ctx.NumSamples()) / time.Duration(frame.ctx.SampleRate())
}
