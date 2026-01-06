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
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Player combines window and audio for easy playback of decoded frames.
type Player struct {
	window          *Window
	audio           *Audio
	videoResampler  *ffmpeg.Resampler
	audioResampler  *ffmpeg.Resampler
	lastFmt         string
	lastW           int
	lastH           int
	lastPTS         float64
	lastFrameDelay  float64 // Last frame delay in seconds
	frameTimer      float64 // Accumulated time for frame scheduling
	audioClock      float64 // Audio PTS at last queue operation
	audioStarted    bool    // Whether audio playback has been started
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewPlayer creates a new player for displaying video and playing audio.
// It will auto-setup when the first frame is received.
func (c *Context) NewPlayer() *Player {
	return &Player{
		lastPTS:        ffmpeg.TS_UNDEFINED,
		lastFrameDelay: 0.040, // 40ms default (25fps)
		frameTimer:     0.0,   // Will be initialized on first frame
		audioClock:     ffmpeg.TS_UNDEFINED,
	}
}

// Close closes the player and releases all resources.
func (p *Player) Close() error {
	var result error

	if p.videoResampler != nil {
		result = errors.Join(result, p.videoResampler.Close())
		p.videoResampler = nil
	}

	if p.audioResampler != nil {
		result = errors.Join(result, p.audioResampler.Close())
		p.audioResampler = nil
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
	dbg("received frame type=%d", frameType)
	switch frameType {
	case media.VIDEO:
		return p.playVideo(ctx, frame)
	case media.AUDIO:
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

// VideoDelay returns how long to wait before presenting the next frame.
// Implements A/V sync following the FFmpeg tutorial algorithm.
func (p *Player) VideoDelay(frame *ffmpeg.Frame) time.Duration {
	if frame == nil || frame.Type() != media.VIDEO {
		return 0
	}

	pts := frame.Ts()
	if pts == ffmpeg.TS_UNDEFINED {
		return 0
	}

	// Initialize frame timer on first frame
	if p.frameTimer == 0.0 {
		p.frameTimer = float64(time.Now().UnixMicro()) / 1000000.0
		p.lastPTS = pts
		dbg("initialized frameTimer=%.3f firstPTS=%.3f", p.frameTimer, pts)
		return 0
	}

	if p.lastPTS == ffmpeg.TS_UNDEFINED {
		p.lastPTS = pts
		return 0
	}

	// Calculate PTS delay (time between this frame and last frame)
	ptsDelay := pts - p.lastPTS
	if ptsDelay <= 0 || ptsDelay >= 1.0 {
		// If delay is invalid, use last frame's delay
		ptsDelay = p.lastFrameDelay
	}

	// Save for next time
	p.lastFrameDelay = ptsDelay
	p.lastPTS = pts

	// Sync to audio if available
	if p.audio != nil && p.audioClock != ffmpeg.TS_UNDEFINED {
		// Get audio clock (pts - buffered duration)
		queuedBytes := p.audio.QueuedSize()
		bytesPerSec := float64(int(p.audio.spec.Freq) * int(p.audio.spec.Channels) * 4)
		bufferedDuration := float64(queuedBytes) / bytesPerSec
		audioRefClock := p.audioClock - bufferedDuration

		// Calculate how far video is from audio
		audioVideoDiff := pts - audioRefClock

		// Sync threshold: use larger of pts_delay or 10ms
		syncThreshold := ptsDelay
		if syncThreshold < 0.010 {
			syncThreshold = 0.010
		}

		dbg("video PTS=%.3f audio=%.3f diff=%.3f ptsDelay=%.3f sync=%.3f queued=%d",
			pts, audioRefClock, audioVideoDiff, ptsDelay, syncThreshold, queuedBytes)

		// Only sync if difference is reasonable (< 1 second)
		if audioVideoDiff < -syncThreshold && audioVideoDiff > -1.0 {
			// Video is behind audio - don't delay
			dbg("video behind, skip delay")
			ptsDelay = 0
		} else if audioVideoDiff >= syncThreshold && audioVideoDiff < 1.0 {
			// Video is ahead of audio - increase delay
			dbg("video ahead, double delay")
			ptsDelay = 2 * ptsDelay
		}
	}

	// Update frame timer
	p.frameTimer += ptsDelay

	// Calculate actual delay based on real time
	now := float64(time.Now().UnixMicro()) / 1000000.0
	realDelay := p.frameTimer - now

	dbg("frameTimer=%.3f now=%.3f realDelay=%.3f ptsDelay=%.3f", p.frameTimer, now, realDelay, ptsDelay)

	// If we're way behind (>100ms), re-sync by resetting frame timer
	if realDelay < -0.1 {
		dbg("way behind (%.3fs), resync frame timer", realDelay)
		p.frameTimer = now
		realDelay = 0.010
	}

	// Ensure minimum delay
	if realDelay < 0.010 {
		realDelay = 0.010
	}

	return time.Duration(realDelay * float64(time.Second))
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

	dbg("audio frame fmt=%s rate=%d channels=%d samples=%d", sampleFmt, sampleRate, channels, frame.NumSamples())

	// Create audio device if needed
	if p.audio == nil {
		var err error
		p.audio, err = ctx.NewAudio(int32(sampleRate), uint8(channels), 4096)
		if err != nil {
			return fmt.Errorf("create audio: %w", err)
		}
		dbg("audio device created rate=%d channels=%d", sampleRate, channels)
	}

	// Convert to float32 if needed
	if sampleFmt != "flt" && sampleFmt != "fltp" {
		converted, err := p.convertAudio(frame)
		if err != nil {
			return fmt.Errorf("convert audio: %w", err)
		}
		if converted == nil {
			return nil
		}
		frame = converted
		sampleFmt = frame.SampleFormat().String()
	}

	// Queue the audio data
	return p.queueFloatAudio(frame)
}

// convertAudio converts audio to float32 planar format for SDL
func (p *Player) convertAudio(frame *ffmpeg.Frame) (*ffmpeg.Frame, error) {
	if frame == nil {
		return nil, nil
	}

	// Create resampler if needed
	if p.audioResampler == nil {
		chLayout := frame.ChannelLayout()
		channelLayout, err := ff.AVUtil_channel_layout_describe(&chLayout)
		if err != nil {
			return nil, err
		}
		par, err := ffmpeg.NewAudioPar("fltp", channelLayout, frame.SampleRate())
		if err != nil {
			return nil, err
		}
		p.audioResampler, err = ffmpeg.NewResampler(par, true)
		if err != nil {
			return nil, err
		}
		dbg("audio resampler created for conversion to fltp")
	}

	// Resample the frame - collect resampled frame
	var resampled *ffmpeg.Frame
	err := p.audioResampler.Resample(frame, func(f *ffmpeg.Frame) error {
		// Copy the frame since Resample reuses the frame buffer
		var err error
		resampled, err = f.Copy()
		return err
	})
	if err != nil {
		return nil, err
	}

	return resampled, nil
}

func (p *Player) queueFloatAudio(frame *ffmpeg.Frame) error {
	numSamples := frame.NumSamples()
	channels := frame.ChannelLayout().NumChannels()

	// Update audio clock to the PTS of this audio packet
	if pts := frame.Ts(); pts != ffmpeg.TS_UNDEFINED {
		p.audioClock = pts
	}
	// Advance audio clock by the duration of this frame
	duration := float64(numSamples) / float64(frame.SampleRate())
	p.audioClock += duration
	dbg("audio PTS=%.3f -> %.3f (+%.3fs, %d samples)",
		frame.Ts(), p.audioClock, duration, numSamples)

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
		if err := p.audio.Queue(audioBytes); err != nil {
			return err
		}
	} else {
		// Non-planar float audio - already interleaved
		plane := frame.Float32(0)
		audioBytes := (*[1 << 30]byte)(unsafe.Pointer(&plane[0]))[:len(plane)*4]
		if err := p.audio.Queue(audioBytes); err != nil {
			return err
		}
	}

	// Start audio playback once we have ~100ms buffered
	if !p.audioStarted && p.audioClock != ffmpeg.TS_UNDEFINED && p.audioClock >= 0.1 {
		queuedBytes := p.audio.QueuedSize()
		// Only resume if we have actual queued audio (not already playing)
		if queuedBytes > 0 {
			// Check if this is the first time we're starting (queued > 50ms worth)
			bytesPerSec := float64(int(p.audio.spec.Freq) * int(p.audio.spec.Channels) * 4)
			bufferedDuration := float64(queuedBytes) / bytesPerSec
			if bufferedDuration > 0.05 {
				dbg("starting audio playback with %.3fs buffered", bufferedDuration)
				p.audio.Resume()
				p.audioStarted = true
			}
		}
	}

	return nil
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
