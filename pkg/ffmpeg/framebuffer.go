package ffmpeg

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	// Packages
	schema "github.com/mutablelogic/go-media/pkg/ffmpeg/schema"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

///////////////////////////////////////////////////////////////////////////////
// ERRORS

var (
	ErrBufferFull    = errors.New("buffer full")
	ErrOutOfOrder    = errors.New("frame PTS before current position")
	ErrUnknownStream = errors.New("unknown stream")
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// FrameBuffer queues decoded frames/subtitles for consumption for live streaming,
// transcoding and video player.
type FrameBuffer interface {
	// Push adds a frame to the buffer. Pass nil to signal EOF for the stream.
	// Returns ErrBufferFull if duration limit exceeded.
	// Returns ErrOutOfOrder if PTS <= newest PTS in stream.
	// Caller retains ownership of frame, buffer takes a reference.
	Push(*schema.Frame) error

	// Next returns the next frame with PTS > afterT across all streams.
	// Returns io.EOF when all streams closed and drained.
	// Returns nil, nil if no frames available yet.
	// Caller owns returned frame and must free when done.
	Next(afterT int64) (*schema.Frame, error)

	// Flush clears all queues and resets closed state. Used for seeking.
	Flush()

	// Stats returns current buffer statistics.
	Stats() BufferStats
}

// BufferStats provides current buffer state
type BufferStats struct {
	Streams     int
	TotalFrames int
	OldestPTS   int64
	NewestPTS   int64
	Duration    int64
	Full        bool
	AllClosed   bool
}

type framebuffer struct {
	timebase    ff.AVRational
	maxDuration int64
	frames      map[int]*framequeue
}

type framequeue struct {
	mu       sync.Mutex
	timebase ff.AVRational
	closed   atomic.Bool
	oldest   atomic.Int64
	newest   atomic.Int64
	count    atomic.Int32
	frames   []*schema.Frame
	pts      []int64 // normalised PTS, same index as frames
}

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewFrameBuffer creates a buffer with specified timebase and max duration.
// Streams must be provided at construction time and cannot be added later.
func NewFrameBuffer(timebase ff.AVRational, maxDuration time.Duration, streams ...*schema.Stream) (*framebuffer, error) {
	self := new(framebuffer)

	// If there are no streams, return an error
	if len(streams) == 0 {
		return nil, ErrUnknownStream
	}

	// Store timebase and convert duration to timebase units
	self.timebase = timebase
	self.maxDuration = ff.AVUtil_rational_rescale_q(
		maxDuration.Microseconds(),
		ff.AVUtil_rational_d2q(1.0/1000000.0, 0),
		timebase,
	)

	// Create frame queues
	self.frames = make(map[int]*framequeue, len(streams))
	for _, stream := range streams {
		index := stream.Index()
		if _, exists := self.frames[index]; exists {
			return nil, ErrUnknownStream // duplicate stream index
		}
		self.frames[index] = newFrameQueue(stream.TimeBase())
	}

	return self, nil
}

func newFrameQueue(timebase ff.AVRational) *framequeue {
	q := &framequeue{
		timebase: timebase,
		frames:   make([]*schema.Frame, 0, 64),
		pts:      make([]int64, 0, 64),
	}
	q.oldest.Store(-1)
	q.newest.Store(-1)
	return q
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - FRAMEBUFFER

// Push adds a frame to the buffer. Pass nil to signal EOF for the stream.
func (fb *framebuffer) Push(frame *schema.Frame) error {
	// Handle nil frame (EOF signal) - need stream index from somewhere
	if frame == nil {
		return nil // TODO: need way to signal EOF per stream
	}

	index := frame.Stream
	queue, exists := fb.frames[index]
	if !exists {
		return ErrUnknownStream
	}

	// Convert PTS to common timebase
	pts := ff.AVUtil_rational_rescale_q(frame.Pts, queue.timebase, fb.timebase)

	// Check buffer full (no lock needed - uses atomics)
	oldest := fb.oldestPTS()
	if oldest >= 0 && (pts-oldest) > fb.maxDuration {
		return ErrBufferFull
	}

	// Lock and push
	queue.mu.Lock()
	defer queue.mu.Unlock()

	// Check out-of-order
	newest := queue.newest.Load()
	if newest >= 0 && pts <= newest {
		return ErrOutOfOrder
	}

	// Take reference and store
	ref := frame.Ref()
	queue.frames = append(queue.frames, ref)
	queue.pts = append(queue.pts, pts)

	// Update atomics
	if queue.oldest.Load() < 0 {
		queue.oldest.Store(pts)
	}
	queue.newest.Store(pts)
	queue.count.Add(1)

	return nil
}

// Next returns the next frame with PTS >= afterT across all streams.
func (fb *framebuffer) Next(afterT int64) (*schema.Frame, error) {
	var candidate *framequeue
	var candidatePTS int64 = -1

	// First pass: atomics only, no locks
	for _, q := range fb.frames {
		oldest := q.oldest.Load()
		if oldest >= afterT && (candidatePTS < 0 || oldest < candidatePTS) {
			candidate = q
			candidatePTS = oldest
		}
	}

	if candidate != nil {
		// Lock only the winner and pop
		frame, _ := candidate.pop()
		if frame != nil {
			return frame, nil
		}
	}

	// Check if all drained
	if fb.allDrained() {
		return nil, io.EOF
	}

	return nil, nil
}

// Flush clears all queues and resets closed state. Used for seeking.
func (fb *framebuffer) Flush() {
	for _, q := range fb.frames {
		q.flush()
	}
}

// Stats returns current buffer statistics.
func (fb *framebuffer) Stats() BufferStats {
	stats := BufferStats{
		Streams:   len(fb.frames),
		OldestPTS: -1,
		NewestPTS: -1,
		AllClosed: true,
	}

	for _, q := range fb.frames {
		count := int(q.count.Load())
		stats.TotalFrames += count

		oldest := q.oldest.Load()
		if oldest >= 0 && (stats.OldestPTS < 0 || oldest < stats.OldestPTS) {
			stats.OldestPTS = oldest
		}

		newest := q.newest.Load()
		if newest >= 0 && (stats.NewestPTS < 0 || newest > stats.NewestPTS) {
			stats.NewestPTS = newest
		}

		if !q.closed.Load() {
			stats.AllClosed = false
		}
	}

	if stats.OldestPTS >= 0 && stats.NewestPTS >= 0 {
		stats.Duration = stats.NewestPTS - stats.OldestPTS
		stats.Full = stats.Duration >= fb.maxDuration
	}

	return stats
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - FRAMEBUFFER

// oldestPTS returns the oldest PTS across all queues (no locking)
func (fb *framebuffer) oldestPTS() int64 {
	oldest := int64(-1)
	for _, q := range fb.frames {
		qOldest := q.oldest.Load()
		if qOldest >= 0 && (oldest < 0 || qOldest < oldest) {
			oldest = qOldest
		}
	}
	return oldest
}

// allDrained returns true if all streams are closed and empty
func (fb *framebuffer) allDrained() bool {
	for _, q := range fb.frames {
		if !q.closed.Load() || q.count.Load() > 0 {
			return false
		}
	}
	return true
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - FRAMEQUEUE

// pop removes and returns the oldest frame from the queue
func (q *framequeue) pop() (*schema.Frame, int64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.frames) == 0 {
		return nil, -1
	}

	frame := q.frames[0]
	pts := q.pts[0]
	// Stamp the rescaled PTS onto the frame so downstream consumers see common timebase
	frame.Pts = pts

	// Remove from slices
	q.frames = q.frames[1:]
	q.pts = q.pts[1:]

	// Update atomics
	q.count.Add(-1)
	if len(q.frames) == 0 {
		q.oldest.Store(-1)
		q.newest.Store(-1)
	} else {
		q.oldest.Store(q.pts[0])
	}

	return frame, pts
}

// flush clears the queue and resets state
func (q *framequeue) flush() {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Unref all frames
	for _, frame := range q.frames {
		frame.Unref()
	}

	// Reset slices
	q.frames = q.frames[:0]
	q.pts = q.pts[:0]

	// Reset atomics
	q.oldest.Store(-1)
	q.newest.Store(-1)
	q.count.Store(0)
	q.closed.Store(false)
}
