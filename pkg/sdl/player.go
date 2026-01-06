package sdl

import (
	"errors"
	"fmt"
	"os"
	"time"
	"unsafe"

	// Packages
	"github.com/mutablelogic/go-media"
	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Player combines window and audio for easy playback of decoded frames.
type Player struct {
	window         *Window
	audio          *Audio
	videoResampler *ffmpeg.Resampler
	lastFmt        string
	lastW          int
	lastH          int
	lastPTS        float64
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewPlayer creates a new player for displaying video and playing audio.
// It will auto-setup when the first frame is received.
func (c *Context) NewPlayer() *Player {
	return &Player{lastPTS: ffmpeg.TS_UNDEFINED}
}

// Close closes the player and releases all resources.
func (p *Player) Close() error {
	var result error

	if p.videoResampler != nil {
		result = errors.Join(result, p.videoResampler.Close())
		p.videoResampler = nil
	}

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
// PUBLIC METHODS - ACCESSORS

// SetWindow sets the window for video playback
func (p *Player) SetWindow(w *Window) {
	p.window = w
}

// Window returns the current window, if any
func (p *Player) Window() *Window {
	return p.window
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
		// Silently ignore unknown frame types
		return nil
	}
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - VIDEO

func (p *Player) playVideo(ctx *Context, frame *ffmpeg.Frame) error {
	width := frame.Width()
	height := frame.Height()
	pixFmt := frame.PixelFormat().String()
	if pixFmt != p.lastFmt || width != p.lastW || height != p.lastH {
		dbg("video frame fmt=%s size=%dx%d", pixFmt, width, height)
		p.lastFmt, p.lastW, p.lastH = pixFmt, width, height
	}

	// Create window if needed
	if p.window == nil {
		var err error
		p.window, err = ctx.NewWindow("Media Player", int32(width), int32(height))
		if err != nil {
			return fmt.Errorf("create window: %w", err)
		}
		dbg("window created %dx%d", width, height)
	}

	// Support yuv420p and rgb24 formats
	switch pixFmt {
	case "yuv420p":
		return p.playYUV(frame)
	case "rgb24":
		return p.playRGB(frame)
	default:
		dbg("convert from %s -> yuv420p", pixFmt)
		converted, err := p.convertVideo(frame, "yuv420p")
		if err != nil {
			return fmt.Errorf("convert video: %w", err)
		}
		if converted == nil {
			return nil
		}
		return p.playYUV(converted)
	}
}

// convertVideo converts an incoming frame to a target pixel format for display.
// It caches a resampler so repeated frames avoid reallocation.
func (p *Player) convertVideo(frame *ffmpeg.Frame, targetPixFmt string) (*ffmpeg.Frame, error) {
	if frame == nil {
		return nil, nil
	}

	if p.videoResampler == nil {
		par, err := ffmpeg.NewVideoPar(targetPixFmt, fmt.Sprintf("%dx%d", frame.Width(), frame.Height()), 0)
		if err != nil {
			return nil, err
		}
		r, err := ffmpeg.NewResampler(par, true)
		if err != nil {
			return nil, err
		}
		p.videoResampler = r
	}

	var out *ffmpeg.Frame
	err := p.videoResampler.Resample(frame, func(f *ffmpeg.Frame) error {
		out = f
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (p *Player) playYUV(frame *ffmpeg.Frame) error {
	yPlane := frame.Bytes(0)
	uPlane := frame.Bytes(1)
	vPlane := frame.Bytes(2)

	// Skip frames with empty planes (shouldn't happen for valid video frames)
	if len(yPlane) == 0 || len(uPlane) == 0 || len(vPlane) == 0 {
		dbg("yuv planes empty: y=%d u=%d v=%d", len(yPlane), len(uPlane), len(vPlane))
		return nil
	}

	yStride := frame.Stride(0)
	uStride := frame.Stride(1)
	vStride := frame.Stride(2)

	if err := p.window.Update(yPlane, uPlane, vPlane, yStride, uStride, vStride, int32(frame.Width()), int32(frame.Height())); err != nil {
		dbg("window update YUV failed: %v", err)
		return err
	}

	if err := p.window.Render(); err != nil {
		dbg("window render failed: %v", err)
		return err
	}

	return nil
}

// VideoDelay returns how long to wait before presenting the next frame based on PTS.
// If PTS is undefined, returns 0 to present immediately.
func (p *Player) VideoDelay(frame *ffmpeg.Frame) time.Duration {
	if frame == nil || frame.Type() != media.VIDEO {
		return 0
	}

	pts := frame.Ts()
	if pts == ffmpeg.TS_UNDEFINED {
		return 0
	}

	if p.lastPTS == ffmpeg.TS_UNDEFINED {
		p.lastPTS = pts
		return 0
	}

	delta := pts - p.lastPTS
	if delta < 0 {
		delta = 0
	}
	// Clamp to avoid very long sleeps on stalled timestamps
	if delta > 0.25 {
		delta = 0.25
	}
	p.lastPTS = pts

	return time.Duration(delta * float64(time.Second))
}

func (p *Player) playRGB(frame *ffmpeg.Frame) error {
	rgbData := frame.Bytes(0)
	stride := frame.Stride(0)

	if err := p.window.UpdateRGB(rgbData, stride, int32(frame.Width()), int32(frame.Height())); err != nil {
		dbg("window update RGB failed: %v", err)
		return err
	}

	if err := p.window.Render(); err != nil {
		dbg("window render failed: %v", err)
		return err
	}

	return nil
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

func dbg(format string, args ...interface{}) {
	if os.Getenv("GOMEDIA_SDL_DEBUG") == "" {
		return
	}
	fmt.Fprintf(os.Stderr, "[sdl] "+format+"\n", args...)
}

// Audio returns the player's audio device (may be nil if no audio frames yet).
func (p *Player) Audio() *Audio {
	return p.audio
}
