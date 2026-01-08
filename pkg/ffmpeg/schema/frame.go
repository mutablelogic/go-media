package schema

import (
	"sync/atomic"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type Frame struct {
	*ff.AVFrame    `json:"-"`
	*ff.AVSubtitle `json:"-"`

	// JSON Fields
	Stream int   `json:"stream,omitempty"` // Stream index
	Pts    int64 `json:"pts,omitempty"`    // Presentation timestamp

	// Subtitle reference counting (AVFrame uses native refcounting)
	refcount *atomic.Int32 `json:"-"`
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewFrame(frame *ff.AVFrame, stream int) *Frame {
	if frame == nil {
		return nil
	}
	return &Frame{
		AVFrame: frame,
		Stream:  stream,
		Pts:     frame.Pts(),
	}
}

func NewSubtitle(subtitle *ff.AVSubtitle, stream int) *Frame {
	if subtitle == nil {
		return nil
	}
	rc := &atomic.Int32{}
	rc.Store(1) // Initial reference count
	return &Frame{
		AVSubtitle: subtitle,
		Stream:     stream,
		Pts:        subtitle.PTS(),
		refcount:   rc,
	}
}

////////////////////////////////////////////////////////////////////////////////
// METHODS

// Ref increments the reference count and returns a new reference.
// For AVFrame: uses av_frame_ref() for efficient buffer sharing.
// For AVSubtitle: uses manual reference counting.
func (f *Frame) Ref() *Frame {
	switch {
	case f.AVFrame != nil:
		// AVFrame: allocate new frame and reference original's buffers
		newFrame := ff.AVUtil_frame_alloc()
		if newFrame == nil {
			return nil
		}
		if err := ff.AVUtil_frame_ref(newFrame, f.AVFrame); err != nil {
			ff.AVUtil_frame_free(newFrame)
			return nil
		}
		return &Frame{
			AVFrame: newFrame,
			Stream:  f.Stream,
			Pts:     f.Pts,
		}

	case f.AVSubtitle != nil:
		// AVSubtitle: manual reference counting (no native ref in FFmpeg)
		if f.refcount != nil {
			f.refcount.Add(1)
		}
		return f // Return same pointer with incremented refcount

	default:
		return nil
	}
}

// Unref decrements the reference count and frees resources when count reaches zero.
func (f *Frame) Unref() {
	switch {
	case f.AVFrame != nil:
		// AVFrame: use native unref
		ff.AVUtil_frame_unref(f.AVFrame)

	case f.AVSubtitle != nil && f.refcount != nil:
		// AVSubtitle: manual reference counting
		if f.refcount.Add(-1) <= 0 {
			ff.AVSubtitle_free(f.AVSubtitle)
		}
	}
}
