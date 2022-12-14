package media

import (
	// Packages
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type scaler struct {
	ctx *ffmpeg.SWSContext
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new video scalar for a source and destination frame. If the destination
// is nil then use the sample parameters as the source.
func NewScalar(src, dest *ffmpeg.AVFrame) *scaler {
	this := new(scaler)

	// if source frame is nil or not a video frame, return nil
	if src == nil || src.PixelFormat() == ffmpeg.AV_PIX_FMT_NONE {
		return nil
	}

	// if destination frame is nil, use source frame as scalar
	if dest == nil {
		dest = src
	}

	// Allocate context - use BILINEAR scaling?
	// TODO allow other scaling algorithms and also filters
	if ctx := ffmpeg.SWS_get_context(
		src.Width(), src.Height(), src.PixelFormat(),
		dest.Width(), dest.Height(), dest.PixelFormat(),
		ffmpeg.SWS_BILINEAR, nil, nil, nil); ctx == nil {
		return nil
	} else {
		this.ctx = ctx
	}

	// Return success
	return this
}

// Release scalar resources
func (scaler *scaler) Close() error {
	var result error

	// Release context
	if scaler.ctx != nil {
		ffmpeg.SWS_free_context(scaler.ctx)
	}

	// Blank out other fields
	scaler.ctx = nil

	// Return success
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (scaler *scaler) String() string {
	str := "<media.scalar"
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Perform the scaling operation on one frame
func (scaler *scaler) Scale(src, dest *ffmpeg.AVFrame) error {
	if src == nil || dest == nil || src.PixelFormat() == ffmpeg.AV_PIX_FMT_NONE || dest.PixelFormat() == ffmpeg.AV_PIX_FMT_NONE {
		return ErrBadParameter.With("Scale")
	}
	if !ffmpeg.SWS_is_supported_input(src.PixelFormat()) {
		return ErrBadParameter.With("Unsupported source pixel format", src.PixelFormat())
	}
	if !ffmpeg.SWS_is_supported_output(dest.PixelFormat()) {
		return ErrBadParameter.With("Unsupported destination pixel format", dest.PixelFormat())
	}
	return ffmpeg.SWS_scale_frame(scaler.ctx, src, dest)
}
