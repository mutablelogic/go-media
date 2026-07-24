package frame

import (
	// Packages
	gomedia "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// Frame is the decoded (or to-be-encoded) output for a single stream. It is
// implemented by AudioFrame and VideoFrame (both backed by an AVFrame - the
// same underlying FFmpeg struct for either kind) and SubtitleFrame (backed
// by an AVSubtitle, which FFmpeg represents and codes completely separately
// from AVFrame, via its own legacy encode/decode API).
type Frame interface {
	Stream() int  // The stream this frame belongs to
	Close() error // Release the frame's resources
}

// mediaFrame is the shared implementation behind AudioFrame and VideoFrame.
type mediaFrame struct {
	*ff.AVFrame
	stream int
}

// AudioFrame is a decoded (or to-be-encoded) audio frame.
type AudioFrame struct{ mediaFrame }

// VideoFrame is a decoded (or to-be-encoded) video frame.
type VideoFrame struct{ mediaFrame }

// SubtitleFrame is a decoded (or to-be-encoded) subtitle.
type SubtitleFrame struct {
	*ff.AVSubtitle
	stream int
}

var (
	_ Frame = (*AudioFrame)(nil)
	_ Frame = (*VideoFrame)(nil)
	_ Frame = (*SubtitleFrame)(nil)
)

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func newMediaFrame(stream int) (mediaFrame, error) {
	f := ff.AVUtil_frame_alloc()
	if f == nil {
		return mediaFrame{}, gomedia.ErrInternalError.With("failed to allocate frame")
	}
	return mediaFrame{AVFrame: f, stream: stream}, nil
}

// NewAudioFrame allocates an audio frame for the given stream.
func NewAudioFrame(stream int) (*AudioFrame, error) {
	m, err := newMediaFrame(stream)
	if err != nil {
		return nil, err
	}
	return &AudioFrame{m}, nil
}

// NewVideoFrame allocates a video frame for the given stream.
func NewVideoFrame(stream int) (*VideoFrame, error) {
	m, err := newMediaFrame(stream)
	if err != nil {
		return nil, err
	}
	return &VideoFrame{m}, nil
}

// NewSubtitleFrame wraps an already-populated AVSubtitle - either decoded by
// Decoder, or built by the caller for Encoder.Encode - for the given stream.
func NewSubtitleFrame(stream int, sub *ff.AVSubtitle) *SubtitleFrame {
	return &SubtitleFrame{AVSubtitle: sub, stream: stream}
}

func (f *mediaFrame) Stream() int { return f.stream }

// Close releases the frame's resources.
func (f *mediaFrame) Close() error {
	if f.AVFrame != nil {
		ff.AVUtil_frame_free(f.AVFrame)
		f.AVFrame = nil
	}
	return nil
}

func (f *SubtitleFrame) Stream() int { return f.stream }

// Close releases the subtitle's resources.
func (f *SubtitleFrame) Close() error {
	if f.AVSubtitle != nil {
		ff.AVSubtitle_free(f.AVSubtitle)
		f.AVSubtitle = nil
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// AllocateBuffers allocates sample/pixel data buffers for the frame based on
// its currently-set format fields (sample format, channel layout and number
// of samples for audio; width, height and pixel format for video).
func (f *mediaFrame) AllocateBuffers() error {
	return ff.AVUtil_frame_get_buffer(f.AVFrame, true)
}

// MakeWritable ensures the frame's data buffers are exclusively owned before
// mutating them in place. Needed before reusing the same Frame for a second
// call to Encoder.Encode, since the encoder may retain a reference to the
// buffer from the first call.
func (f *mediaFrame) MakeWritable() error {
	return ff.AVUtil_frame_make_writable(f.AVFrame)
}

// IncPts advances the frame's Pts by v.
func (f *mediaFrame) IncPts(v int64) {
	f.SetPts(f.Pts() + v)
}
