package frame

import (
	// Packages
	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// Frame wraps an AVFrame with the stream ID it belongs to, so it can be
// routed to the right codec by Encoder.Encode.
type Frame struct {
	*ff.AVFrame
	StreamID int
}

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewFrame allocates a frame for the given stream.
func NewFrame(streamID int) (*Frame, error) {
	frame := ff.AVUtil_frame_alloc()
	if frame == nil {
		return nil, gomedia.ErrInternalError.With("failed to allocate frame")
	}
	return &Frame{AVFrame: frame, StreamID: streamID}, nil
}

// Close releases the frame's resources.
func (f *Frame) Close() error {
	if f.AVFrame != nil {
		ff.AVUtil_frame_free(f.AVFrame)
		f.AVFrame = nil
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AllocateBuffers allocates sample/pixel data buffers for the frame based on
// its currently-set format fields (sample format, channel layout and number
// of samples for audio; width, height and pixel format for video).
func (f *Frame) AllocateBuffers() error {
	return ff.AVUtil_frame_get_buffer(f.AVFrame, true)
}

// MakeWritable ensures the frame's data buffers are exclusively owned before
// mutating them in place. Needed before reusing the same Frame for a second
// call to Encoder.Encode, since the encoder may retain a reference to the
// buffer from the first call.
func (f *Frame) MakeWritable() error {
	return ff.AVUtil_frame_make_writable(f.AVFrame)
}

// IncPts advances the frame's Pts by v.
func (f *Frame) IncPts(v int64) {
	f.SetPts(f.Pts() + v)
}
