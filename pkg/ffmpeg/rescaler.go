package ffmpeg

import (
	"errors"

	// Packages

	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type rescaler struct {
	opts

	src_pix_fmt ff.AVPixelFormat
	src_width   int
	src_height  int
	ctx         *ff.SWSContext
	flags       ff.SWSFlag
	dest        *ff.AVFrame
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new rescaler which will rescale the input frame to the
// specified format, width and height.
func NewRescaler(format ff.AVPixelFormat, opt ...Opt) (*rescaler, error) {
	rescaler := new(rescaler)

	// Apply options
	rescaler.pix_fmt = format
	rescaler.width = 640
	rescaler.height = 480
	for _, o := range opt {
		if err := o(&rescaler.opts); err != nil {
			return nil, err
		}
	}

	// Check parameters
	if rescaler.pix_fmt == ff.AV_PIX_FMT_NONE {
		return nil, errors.New("invalid parameters")
	}

	// Create a destimation frame
	dest := ff.AVUtil_frame_alloc()
	if dest == nil {
		return nil, errors.New("failed to allocate frame")
	}

	// Set parameters
	dest.SetPixFmt(rescaler.pix_fmt)
	dest.SetWidth(rescaler.width)
	dest.SetHeight(rescaler.height)

	// Allocate buffer
	if err := ff.AVUtil_frame_get_buffer(dest, false); err != nil {
		ff.AVUtil_frame_free(dest)
		return nil, err
	} else {
		rescaler.dest = dest
		rescaler.flags = ff.SWS_POINT
	}

	// Return success
	return rescaler, nil
}

// Release resources
func (r *rescaler) Close() error {
	if r.ctx != nil {
		ff.SWScale_free_context(r.ctx)
		r.ctx = nil
	}
	if r.dest != nil {
		ff.AVUtil_frame_free(r.dest)
		r.dest = nil
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Scale the source image and return the destination image
func (r *rescaler) Frame(src *ff.AVFrame) (*ff.AVFrame, error) {
	// If source is null then return null (no flushing)
	if src == nil {
		return nil, nil
	}

	// Simply return the frame if it matches the destination format
	if matchesVideoFormat(src, r.dest) && !r.force {
		return src, nil
	}

	// Allocate a context
	if r.ctx == nil || r.src_pix_fmt != src.PixFmt() || r.src_width != src.Width() || r.src_height != src.Height() {
		// Release existing scaling context, if any
		if r.ctx != nil {
			ff.SWScale_free_context(r.ctx)
		}
		// Create a new scaling context
		ctx := ff.SWScale_get_context(
			src.Width(), src.Height(), src.PixFmt(), // source
			r.dest.Width(), r.dest.Height(), r.dest.PixFmt(), // destination
			r.flags, nil, nil, nil,
		)
		if ctx == nil {
			return nil, errors.New("failed to allocate swscale context")
		} else {
			r.ctx = ctx
			r.src_pix_fmt = src.PixFmt()
			r.src_width = src.Width()
			r.src_height = src.Height()
		}
	}

	// Rescale the image
	if err := ff.SWScale_scale_frame(r.ctx, r.dest, src, false); err != nil {
		return nil, err
	}

	// Copy parameters from the source frame
	if err := ff.AVUtil_frame_copy_props(r.dest, src); err != nil {
		return nil, err
	}

	// Return the destination frame
	return r.dest, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Returns true if the pixel format, width and height of the source
// and destination frames match
func matchesVideoFormat(src, dest *ff.AVFrame) bool {
	if src.PixFmt() == dest.PixFmt() && src.Width() == dest.Width() && src.Height() == dest.Height() {
		return true
	}
	return false
}
