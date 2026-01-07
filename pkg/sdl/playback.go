//go:build sdl2

package sdl

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	ffmpeg "github.com/mutablelogic/go-media/pkg/ffmpeg"
)

// FrameLoop runs an SDL event that pulls frames from a channel and hands them to a handler on the main thread.
// It keeps posting events with small delays to avoid busy looping while still staying responsive.
type FrameLoop struct {
	ctx        *Context
	handler    func(*ffmpeg.Frame) error
	frameCh    chan *ffmpeg.Frame
	event      uint32
	stopped    uint32
	done       chan struct{}
	doneOnce   sync.Once
	closeOnce  sync.Once
	frameDelay time.Duration
	idleDelay  time.Duration
	retryDelay time.Duration
	delayFn    func(*ffmpeg.Frame) time.Duration
}

// FrameLoopOpt configures a FrameLoop.
type FrameLoopOpt func(*FrameLoop)

// WithFrameDelayFunc uses a custom delay calculator per frame (e.g., based on PTS).
func WithFrameDelayFunc(fn func(*ffmpeg.Frame) time.Duration) FrameLoopOpt {
	return func(l *FrameLoop) { l.delayFn = fn }
}

// NewFrameLoop creates a frame loop with the given buffer size and handler. Use Start to begin posting events.
func NewFrameLoop(ctx *Context, handler func(*ffmpeg.Frame) error, buffer int, opts ...FrameLoopOpt) (*FrameLoop, error) {
	if ctx == nil {
		return nil, errors.New("nil ctx")
	}
	if handler == nil {
		return nil, errors.New("nil handler")
	}
	if buffer <= 0 {
		buffer = 1
	}

	loop := &FrameLoop{
		ctx:        ctx,
		handler:    handler,
		frameCh:    make(chan *ffmpeg.Frame, buffer),
		done:       make(chan struct{}),
		frameDelay: 33 * time.Millisecond,
		idleDelay:  10 * time.Millisecond,
		retryDelay: 1 * time.Millisecond,
	}

	for _, opt := range opts {
		opt(loop)
	}

	loop.event = ctx.Register(func(userInfo interface{}) {
		loop.handleEvent()
	})

	return loop, nil
}

// Start kicks off the loop by posting the first event.
func (l *FrameLoop) Start() {
	l.post(0)
}

// Enqueue adds a frame for processing on the main thread.
func (l *FrameLoop) Enqueue(frame *ffmpeg.Frame) error {
	if frame == nil {
		return errors.New("nil frame")
	}
	if atomic.LoadUint32(&l.stopped) != 0 {
		return errors.New("frame loop stopped")
	}

	l.frameCh <- frame
	return l.ctx.Post(l.event, nil)
}

// CloseInput signals that no more frames will arrive and triggers final processing.
func (l *FrameLoop) CloseInput() {
	l.closeOnce.Do(func() {
		close(l.frameCh)
		l.ctx.Post(l.event, nil)
	})
}

// Stop halts further event posting and closes Done.
func (l *FrameLoop) Stop() {
	if atomic.SwapUint32(&l.stopped, 1) == 0 {
		l.doneOnce.Do(func() { close(l.done) })
	}
}

// Done is closed when the loop stops after input closes.
func (l *FrameLoop) Done() <-chan struct{} {
	return l.done
}

// FrameWriter adapts a FrameLoop to the ffmpeg task writer interface.
type FrameWriter struct {
	loop  *FrameLoop
	stats struct {
		frames int
		video  int
		audio  int
	}
}

// NewFrameWriter creates a writer that forwards frames into the loop.
func NewFrameWriter(loop *FrameLoop) *FrameWriter {
	return &FrameWriter{loop: loop}
}

// Write satisfies io.Writer but discards data; decoding writes frames via WriteFrame.
func (w *FrameWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// WriteFrame queues ffmpeg frames into the loop.
func (w *FrameWriter) WriteFrame(streamIndex int, frame interface{}) error {
	f, ok := frame.(*ffmpeg.Frame)
	if !ok {
		return nil
	}

	// Make a copy because decoder frames are reused; render happens on another goroutine.
	copy, err := f.Copy()
	if err != nil {
		return fmt.Errorf("copy frame: %w", err)
	}

	w.stats.frames++
	switch f.Type() {
	case 1:
		w.stats.audio++
	case 2:
		w.stats.video++
	}

	return w.loop.Enqueue(copy)
}

func (l *FrameLoop) handleEvent() {
	if atomic.LoadUint32(&l.stopped) != 0 {
		return
	}

	select {
	case frame, ok := <-l.frameCh:
		if !ok {
			atomic.StoreUint32(&l.stopped, 1)
			l.doneOnce.Do(func() { close(l.done) })
			return
		}

		// Calculate delay BEFORE handling the frame
		delay := l.frameDelay
		if l.delayFn != nil {
			if d := l.delayFn(frame); d >= 0 {
				delay = d
			}
		}

		// IMPORTANT: Sleep BEFORE displaying the frame to maintain proper timing
		// The VideoDelay function calculates when this frame should be shown
		if delay > 0 {
			dbg("waiting %.3fs before displaying frame", delay.Seconds())
			time.Sleep(delay)
		}

		// Now display the frame at the correct time
		if err := l.handler(frame); err != nil {
			dbg("frame handler error: %v", err)
			l.post(l.retryDelay)
			_ = frame.Close()
			return
		}

		_ = frame.Close()

		// Post next event immediately (no additional delay)
		if err := l.ctx.Post(l.event, nil); err != nil {
			dbg("post event error: %v", err)
		}
	default:
		l.post(l.idleDelay)
	}
}

func (l *FrameLoop) post(delay time.Duration) {
	if atomic.LoadUint32(&l.stopped) != 0 {
		return
	}

	go func() {
		if delay > 0 {
			dbg("sleeping for %.3fs before next event", delay.Seconds())
			time.Sleep(delay)
		}
		if err := l.ctx.Post(l.event, nil); err != nil {
			fmt.Println("post event:", err)
		}
	}()
}
