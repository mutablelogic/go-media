package sdl

import (
	"errors"
	"fmt"
	"unsafe"

	// Packages
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Player combines window and audio for easy playback of decoded frames.
type Player struct {
	window *Window
	audio  *Audio
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewPlayer creates a new player for displaying video and playing audio.
// It will auto-setup when the first frame is received.
func (c *Context) NewPlayer() *Player {
	return &Player{}
}

// Close closes the player and releases all resources.
func (p *Player) Close() error {
	var result error

	if p.window != nil {
		if err := p.window.Close(); err != nil {
			result = errors.Join(result, err)
		}
		p.window = nil
	}

	if p.audio != nil {
		if err := p.audio.Close(); err != nil {
			result = errors.Join(result, err)
		}
		p.audio = nil
	}

	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// PlayFrame plays a decoded frame. For video frames, displays them in the window.
// For audio frames, queues them for playback. Auto-creates window/audio on first frame.
func (p *Player) PlayFrame(ctx *Context, frame *ffmpeg.Frame) error {
	if frame == nil {
		return errors.New("nil frame")
	}

	frameType := frame.Type()
	switch frameType {
	case 2: // VIDEO
		return p.playVideo(ctx, frame)
	case 1: // AUDIO
		return p.playAudio(ctx, frame)
	default:
		// Ignore unknown frame types
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - VIDEO

func (p *Player) playVideo(ctx *Context, frame *ffmpeg.Frame) error {
	width := frame.Width()
	height := frame.Height()
	pixFmt := frame.PixelFormat().String()

	// Create window if needed
	if p.window == nil {
		var err error
		p.window, err = ctx.NewWindow("Media Player", int32(width), int32(height))
		if err != nil {
			return fmt.Errorf("create window: %w", err)
		}
	}

	// Support yuv420p and rgb24 formats
	switch pixFmt {
	case "yuv420p":
		return p.playYUV(frame)
	case "rgb24":
		return p.playRGB(frame)
	default:
		return fmt.Errorf("unsupported pixel format: %s", pixFmt)
	}
}

func (p *Player) playYUV(frame *ffmpeg.Frame) error {
	yPlane := frame.Bytes(0)
	uPlane := frame.Bytes(1)
	vPlane := frame.Bytes(2)
	yStride := frame.Stride(0)
	uStride := frame.Stride(1)
	vStride := frame.Stride(2)

	return p.window.Update(yPlane, uPlane, vPlane, yStride, uStride, vStride)
}

func (p *Player) playRGB(frame *ffmpeg.Frame) error {
	rgbData := frame.Bytes(0)
	stride := frame.Stride(0)

	return p.window.UpdateRGB(rgbData, stride)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - AUDIO

func (p *Player) playAudio(ctx *Context, frame *ffmpeg.Frame) error {
	sampleRate := frame.SampleRate()
	channels := frame.ChannelLayout().NumChannels()
	sampleFmt := frame.SampleFormat().String()

	// Create audio device if needed
	if p.audio == nil {
		var err error
		p.audio, err = ctx.NewAudio(int32(sampleRate), uint8(channels), 4096)
		if err != nil {
			return fmt.Errorf("create audio: %w", err)
		}
		// Start playback
		p.audio.Resume()
	}

	// Only support flt (planar float) and fltp (planar float) formats
	switch sampleFmt {
	case "flt", "fltp":
		return p.queueFloatAudio(frame)
	default:
		return fmt.Errorf("unsupported sample format: %s", sampleFmt)
	}
}

func (p *Player) queueFloatAudio(frame *ffmpeg.Frame) error {
	numSamples := frame.NumSamples()
	channels := frame.ChannelLayout().NumChannels()

	// For planar audio, interleave the channels
	if frame.SampleFormat().String() == "fltp" {
		// Interleave planar audio
		interleavedData := make([]float32, numSamples*channels)

		for ch := 0; ch < channels; ch++ {
			plane := frame.Float32(ch)
			for i := 0; i < numSamples; i++ {
				interleavedData[i*channels+ch] = plane[i]
			}
		}

		// Convert to bytes
		audioBytes := (*[1 << 30]byte)(unsafe.Pointer(&interleavedData[0]))[:len(interleavedData)*4]
		return p.audio.Queue(audioBytes)
	} else {
		// Non-planar float audio - already interleaved
		plane := frame.Float32(0)
		audioBytes := (*[1 << 30]byte)(unsafe.Pointer(&plane[0]))[:len(plane)*4]
		return p.audio.Queue(audioBytes)
	}
}

// Window returns the player's window (may be nil if no video frames yet).
func (p *Player) Window() *Window {
	return p.window
}

// Audio returns the player's audio device (may be nil if no audio frames yet).
func (p *Player) Audio() *Audio {
	return p.audio
}
