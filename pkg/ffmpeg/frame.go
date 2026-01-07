package ffmpeg

import (
	"encoding/json"
	"errors"
	"fmt"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
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
		frame.SetPts(int64(ff.AV_NOPTS_VALUE))
		return (*Frame)(frame), nil
	}

	// Set parameters based on codec type
	switch par.CodecType() {
	case ff.AVMEDIA_TYPE_AUDIO:
		if err := (*Frame)(frame).setAudioParams(par); err != nil {
			ff.AVUtil_frame_free(frame)
			return nil, err
		}
	case ff.AVMEDIA_TYPE_VIDEO:
		(*Frame)(frame).setVideoParams(par)
	default:
		ff.AVUtil_frame_free(frame)
		return nil, errors.New("invalid codec type")
	}

	// Clear Pts
	frame.SetPts(int64(ff.AV_NOPTS_VALUE))

	// Return success
	return (*Frame)(frame), nil
}

// Helper to set audio parameters on a frame
func (frame *Frame) setAudioParams(par *Par) error {
	ctx := (*ff.AVFrame)(frame)
	ctx.SetSampleFormat(par.SampleFormat())
	if err := ctx.SetChannelLayout(par.ChannelLayout()); err != nil {
		return err
	}
	ctx.SetSampleRate(par.SampleRate())
	// Note: NumSamples is not set here - it should be set when allocating buffers
	ctx.SetTimeBase(ff.AVUtil_rational(1, par.SampleRate()))
	return nil
}

// Helper to set video parameters on a frame
func (frame *Frame) setVideoParams(par *Par) {
	ctx := (*ff.AVFrame)(frame)
	ctx.SetPixFmt(par.PixelFormat())
	ctx.SetWidth(par.Width())
	ctx.SetHeight(par.Height())
	ctx.SetSampleAspectRatio(par.SampleAspectRatio())
	// Use FrameRate to calculate timebase
	if framerate := par.FrameRate(); framerate > 0 {
		ctx.SetTimeBase(ff.AVUtil_rational_invert(ff.AVUtil_rational_d2q(framerate, 1<<24)))
	}
}

// Release frame resources
func (frame *Frame) Close() error {
	if frame != nil {
		ff.AVUtil_frame_free((*ff.AVFrame)(frame))
	}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (frame *Frame) String() string {
	data, _ := json.MarshalIndent((*ff.AVFrame)(frame), "", "  ")
	return string(data)
}

// MarshalJSON implements json.Marshaler by delegating to the underlying AVFrame
func (frame *Frame) MarshalJSON() ([]byte, error) {
	return json.Marshal((*ff.AVFrame)(frame))
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
	if frame == nil {
		return nil, errors.New("frame is nil")
	}

	copy := ff.AVUtil_frame_alloc()
	if copy == nil {
		return nil, errors.New("failed to allocate frame")
	}

	// Copy parameters based on frame type
	frameType := frame.Type()
	if err := frame.copyParameters(copy, frameType); err != nil {
		ff.AVUtil_frame_free(copy)
		return nil, err
	}

	// Allocate and copy buffer data if original has buffers
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

	// Copy properties (timestamps, metadata, side data, etc.)
	if err := ff.AVUtil_frame_copy_props(copy, (*ff.AVFrame)(frame)); err != nil {
		ff.AVUtil_frame_free(copy)
		return nil, fmt.Errorf("AVUtil_frame_copy_props: %w", err)
	}

	return (*Frame)(copy), nil
}

// Helper to copy frame parameters based on type
func (frame *Frame) copyParameters(dst *ff.AVFrame, frameType media.Type) error {
	switch frameType {
	case media.AUDIO:
		dst.SetSampleFormat(frame.SampleFormat())
		dst.SetChannelLayout(frame.ChannelLayout())
		dst.SetSampleRate(frame.SampleRate())
		dst.SetNumSamples(frame.NumSamples())
		dst.SetTimeBase(frame.TimeBase())
		return nil
	case media.VIDEO:
		dst.SetPixFmt(frame.PixelFormat())
		dst.SetWidth(frame.Width())
		dst.SetHeight(frame.Height())
		dst.SetSampleAspectRatio(frame.SampleAspectRatio())
		dst.SetTimeBase(frame.TimeBase())
		return nil
	default:
		return errors.New("invalid codec type")
	}
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
	if frame == nil {
		return media.UNKNOWN
	}
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

// Set plane data from float32 slice
func (frame *Frame) SetFloat32(plane int, data []float32) error {
	if frame.Type() != media.AUDIO {
		return errors.New("frame is not an audio frame")
	}

	// If the number of samples is not the same, the re-allocate the frame
	ctx := (*ff.AVFrame)(frame)
	if len(data) != frame.NumSamples() {
		ctx.SetNumSamples(len(data))
		if err := ff.AVUtil_frame_get_buffer(ctx, false); err != nil {
			ff.AVUtil_frame_unref(ctx)
			return err
		}
	}

	// Copy data
	copy(ctx.Float32(plane), data)

	// Return success
	return nil
}

// Return plane data as a  int16 slice
func (frame *Frame) Int16(plane int) []int16 {
	ctx := (*ff.AVFrame)(frame)
	return ctx.Int16(plane)
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
	if pts == int64(ff.AV_NOPTS_VALUE) {
		return TS_UNDEFINED
	}
	tb := ctx.TimeBase()
	if tb.Num() == 0 || tb.Den() == 0 {
		return TS_UNDEFINED
	}
	return ff.AVUtil_rational_q2d(tb) * float64(pts)
}

// Set timestamp in seconds
func (frame *Frame) SetTs(secs float64) {
	ctx := (*ff.AVFrame)(frame)
	tb := ctx.TimeBase()
	if secs == TS_UNDEFINED || tb.Num() == 0 || tb.Den() == 0 {
		frame.SetPts(int64(ff.AV_NOPTS_VALUE))
		return
	}
	ctx.SetPts(int64(secs / ff.AVUtil_rational_q2d(tb)))
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - COMPARISON

// Returns true if this frame can be resampled/resized to match the other frame.
// Checks format compatibility (sample format, pixel format, channel layout, sample rate)
// but ignores sample count and framerate.
func (frame *Frame) MatchesFormat(other *Frame) bool {
	if frame == nil || other == nil {
		return false
	}

	// Must be same type
	if frame.Type() != other.Type() {
		return false
	}

	switch frame.Type() {
	case media.AUDIO:
		return frame.matchesAudioFormat(other)
	case media.VIDEO:
		return frame.matchesVideoFormat(other)
	default:
		return false
	}
}

// Helper to check audio format compatibility
func (frame *Frame) matchesAudioFormat(other *Frame) bool {
	if frame.SampleFormat() != other.SampleFormat() {
		return false
	}
	if frame.SampleRate() != other.SampleRate() {
		return false
	}
	cha, chb := frame.ChannelLayout(), other.ChannelLayout()
	return ff.AVUtil_channel_layout_compare(&cha, &chb)
}

// Helper to check video format compatibility
func (frame *Frame) matchesVideoFormat(other *Frame) bool {
	if frame.PixelFormat() != other.PixelFormat() {
		return false
	}
	if frame.Width() != other.Width() || frame.Height() != other.Height() {
		return false
	}
	// SampleAspectRatio, TimeBase, and FrameRate don't affect resizing
	return true
}
