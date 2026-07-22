package schema

import (
	"context"

	// Packages
	frame "github.com/mutablelogic/go-media/frame"
	profile "github.com/mutablelogic/go-media/profile/schema"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Source produces frames for one or more streams until ctx is cancelled or
// the source naturally ends. It is the abstract input side of an encoding
// job — a live capture, a file, or a fallback generator like a test card —
// so a job can be built once and swap sources without restarting the encoder.
type Source interface {
	// Streams describes each stream's raw format, in the same order
	// Encoder.Add / Writer.Create expect — stream ID equals slice index. A
	// source's own Profile typically resolves to a raw codec (pcm_f32le,
	// rawvideo, ...) rather than a compressed one.
	Streams() []profile.Profile

	// NextFrame blocks until the next frame is ready, ctx is cancelled, or
	// the source has no more frames (io.EOF). Frames are paced by the
	// source itself — callers just pull in a loop.
	NextFrame(ctx context.Context) (*frame.Frame, error)

	// Close releases resources the source holds (file handles, devices, ...).
	Close() error
}
