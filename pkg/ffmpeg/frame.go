package ffmpeg

import (
	"encoding/json"
	"errors"
	"fmt"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Frame ff.AVFrame

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	PTS_UNDEFINED = ff.AV_NOPTS_VALUE
	TS_UNDEFINED  = -1.0
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new frame and optionally set audio or video parameters
func NewFrame(par *Par) (*Frame, error) {
	frame := ff.AVUtil_frame_alloc()
	if frame == nil {
		return nil, errors.New("failed to allocate frame")
	}

	// If parameters are nil, then return the frame
	if par == nil {
		return (*Frame)(frame), nil
	}

	// Set parameters
	switch par.CodecType() {
	case ff.AVMEDIA_TYPE_AUDIO:
		frame.SetSampleFormat(par.SampleFormat())
		if err := frame.SetChannelLayout(par.ChannelLayout()); err != nil {
			ff.AVUtil_frame_free(frame)
			return nil, err
		}
		frame.SetSampleRate(par.Samplerate())
		frame.SetNumSamples(par.FrameSize())
		frame.SetTimeBase(ff.AVUtil_rational(1, par.Samplerate()))
	case ff.AVMEDIA_TYPE_VIDEO:
		frame.SetPixFmt(par.PixelFormat())
		frame.SetWidth(par.Width())
		frame.SetHeight(par.Height())
		frame.SetSampleAspectRatio(par.SampleAspectRatio())
		frame.SetTimeBase(par.timebase) // Also sets framerate
	default:
		ff.AVUtil_frame_free(frame)
		return nil, errors.New("invalid codec type")
	}

	// Clear Pts
	frame.SetPts(ff.AV_NOPTS_VALUE)

	// Return success
	return (*Frame)(frame), nil
}

// Release frame resources
func (frame *Frame) Close() error {
	ff.AVUtil_frame_free((*ff.AVFrame)(frame))
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (frame *Frame) String() string {
	data, _ := json.MarshalIndent((*ff.AVFrame)(frame), "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - FRAME

// Allocate buffers for the frame
func (frame *Frame) AllocateBuffers() error {
	return ff.AVUtil_frame_get_buffer((*ff.AVFrame)(frame), false)
}

// Return true if the frame has allocated buffers
func (frame *Frame) IsAllocated() bool {
	return ff.AVUtil_frame_is_allocated((*ff.AVFrame)(frame))
}

// Make the frame writable
func (frame *Frame) MakeWritable() error {
	return ff.AVUtil_frame_make_writable((*ff.AVFrame)(frame))
}

// Make a copy of the frame, which should be released by the caller
func (frame *Frame) Copy() (*Frame, error) {
	copy := ff.AVUtil_frame_alloc()
	if copy == nil {
		return nil, errors.New("failed to allocate frame")
	}

	switch frame.Type() {
	case media.AUDIO:
		copy.SetSampleFormat(frame.SampleFormat())
		copy.SetChannelLayout(frame.ChannelLayout())
		copy.SetSampleRate(frame.SampleRate())
		copy.SetNumSamples(frame.NumSamples())
	case media.VIDEO:
		copy.SetPixFmt(frame.PixelFormat())
		copy.SetWidth(frame.Width())
		copy.SetHeight(frame.Height())
		copy.SetSampleAspectRatio(frame.SampleAspectRatio())
	default:
		ff.AVUtil_frame_free(copy)
		return nil, errors.New("invalid codec type")
	}
	if frame.IsAllocated() {
		if err := ff.AVUtil_frame_get_buffer(copy, false); err != nil {
			ff.AVUtil_frame_free(copy)
			return nil, fmt.Errorf("AVUtil_frame_get_buffer: %w", err)
		}
		if err := ff.AVUtil_frame_copy(copy, (*ff.AVFrame)(frame)); err != nil {
			ff.AVUtil_frame_free(copy)
			return nil, fmt.Errorf("AVUtil_frame_copy: %w", err)
		}
	}
	if err := ff.AVUtil_frame_copy_props(copy, (*ff.AVFrame)(frame)); err != nil {
		ff.AVUtil_frame_free(copy)
		return nil, fmt.Errorf("AVUtil_frame_copy_props: %w", err)
	}

	return (*Frame)(copy), nil
}

// Unreference frame buffers
func (frame *Frame) Unref() {
	ff.AVUtil_frame_unref((*ff.AVFrame)(frame))
}

// Copy properties from another frame
func (frame *Frame) CopyPropsFromFrame(other *Frame) error {
	return ff.AVUtil_frame_copy_props((*ff.AVFrame)(frame), (*ff.AVFrame)(other))
}

// Return frame type - AUDIO or VIDEO. Other types are not yet
// identified and returned as UNKNOWN
func (frame *Frame) Type() media.Type {
	switch {
	case frame.SampleRate() > 0 && frame.SampleFormat() != ff.AV_SAMPLE_FMT_NONE:
		return media.AUDIO
	case frame.Width() > 0 && frame.Height() > 0 && frame.PixelFormat() != ff.AV_PIX_FMT_NONE:
		return media.VIDEO
	default:
		return media.UNKNOWN
	}
}

// Return plane data as a float32 slice
func (frame *Frame) Float32(plane int) []float32 {
	ctx := (*ff.AVFrame)(frame)
	return ctx.Float32(plane)
}

// Return plane data as a byte slice
func (frame *Frame) Bytes(plane int) []byte {
	return (*ff.AVFrame)(frame).Bytes(plane)
}

// Return the stride for a plane (number of bytes in a row)
func (frame *Frame) Stride(plane int) int {
	return (*ff.AVFrame)(frame).Linesize(plane)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - AUDIO PARAMETERS

func (frame *Frame) SampleFormat() ff.AVSampleFormat {
	return (*ff.AVFrame)(frame).SampleFormat()
}

func (frame *Frame) ChannelLayout() ff.AVChannelLayout {
	return (*ff.AVFrame)(frame).ChannelLayout()
}

func (frame *Frame) SampleRate() int {
	return (*ff.AVFrame)(frame).SampleRate()
}

func (frame *Frame) FrameSize() int {
	return (*ff.AVFrame)(frame).NumSamples()
}

func (frame *Frame) NumSamples() int {
	return (*ff.AVFrame)(frame).NumSamples()
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - VIDEO PARAMETERS

func (frame *Frame) PixelFormat() ff.AVPixelFormat {
	return (*ff.AVFrame)(frame).PixFmt()
}

func (frame *Frame) Width() int {
	return (*ff.AVFrame)(frame).Width()
}

func (frame *Frame) Height() int {
	return (*ff.AVFrame)(frame).Height()
}

func (frame *Frame) SampleAspectRatio() ff.AVRational {
	return (*ff.AVFrame)(frame).SampleAspectRatio()
}

func (frame *Frame) FrameRate() ff.AVRational {
	return ff.AVUtil_rational_invert((*ff.AVFrame)(frame).TimeBase())
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - TIME PARAMETERS

func (frame *Frame) TimeBase() ff.AVRational {
	return (*ff.AVFrame)(frame).TimeBase()
}

// Return the presentation timestamp in int64
func (frame *Frame) Pts() int64 {
	return (*ff.AVFrame)(frame).Pts()
}

// Set the presentation timestamp in int64
func (frame *Frame) SetPts(v int64) {
	(*ff.AVFrame)(frame).SetPts(v)
}

// Increment presentation timestamp
func (frame *Frame) IncPts(v int64) {
	(*ff.AVFrame)(frame).SetPts((*ff.AVFrame)(frame).Pts() + v)
}

// Return the timestamp in seconds, or TS_UNDEFINED if the timestamp
// is undefined or timebase is not set
func (frame *Frame) Ts() float64 {
	ctx := (*ff.AVFrame)(frame)
	pts := ctx.Pts()
	if pts == ff.AV_NOPTS_VALUE {
		return TS_UNDEFINED
	}
	tb := ctx.TimeBase()
	if tb.Num() == 0 || tb.Den() == 0 {
		return TS_UNDEFINED
	}
	return ff.AVUtil_rational_q2d(tb) * float64(pts)
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Returns true if a AUDIO or VIDEO frame matches the other frame, for
// resampling and resizing purposes. Note does not check number of samples
// or framerate
func (frame *Frame) matchesResampleResize(other *Frame) bool {
	// Match types
	if frame.Type() != other.Type() {
		return false
	}
	switch frame.Type() {
	case media.AUDIO:
		if frame.SampleFormat() != other.SampleFormat() {
			return false
		}
		cha, chb := frame.ChannelLayout(), other.ChannelLayout()
		if !ff.AVUtil_channel_layout_compare(&cha, &chb) {
			return false
		}
		if frame.SampleRate() != other.SampleRate() {
			return false
		}
		return true
	case media.VIDEO:
		if frame.PixelFormat() != other.PixelFormat() {
			return false
		}
		if frame.Width() != other.Width() || frame.Height() != other.Height() {
			return false
		}
		// We don't need to check the SampleAspectRatio, TimeBase or FrameRate
		// for the purposes of resampling or resizing
		return true
	default:
		return false
	}
}
