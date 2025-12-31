package ffmpeg

import (
	"errors"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg71"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type rescaler struct {
	ctx   *ff.SWSContext
	flags ff.SWSFlag
	force bool
	dest  *Frame

	src_pix_fmt ff.AVPixelFormat
	src_width   int
	src_height  int
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new rescaler which will rescale the input frame to the
// specified format, width and height.
func NewRescaler(par *Par, force bool) (*rescaler, error) {
	rescaler := new(rescaler)

	// Check parameters
	if par == nil || par.CodecType() != ff.AVMEDIA_TYPE_VIDEO {
		return nil, errors.New("invalid codec type")
	}
	if par.PixelFormat() == ff.AV_PIX_FMT_NONE {
		return nil, errors.New("invalid pixel format parameters")
	}
	if par.Width() == 0 || par.Height() == 0 {
		return nil, errors.New("invalid width or height parameters")
	}

	// Create a destimation frame
	dest, err := NewFrame(par)
	if err != nil {
		return nil, err
	}

	// Set parameters
	rescaler.dest = dest
	rescaler.force = force
	rescaler.flags = ff.SWS_POINT

	// Allocate buffer
	if err := dest.AllocateBuffers(); err != nil {
		return nil, errors.Join(err, dest.Close())
	} else {
		rescaler.dest = dest

	}

	// Return success
	return rescaler, nil
}

// Release resources
func (r *rescaler) Close() error {
	if r.ctx != nil {
		ff.SWScale_free_context(r.ctx)
	}
	result := r.dest.Close()
	r.dest = nil
	r.ctx = nil
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Scale the source image and return the destination image
func (r *rescaler) Frame(src *Frame) (*Frame, error) {
	// If source is null then return null (no flushing needed)
	// or return the same frame if it matches the destination format
	// and force is not set
	if src == nil {
		return nil, nil
	} else if src.matchesResampleResize(r.dest) && !r.force {
		return src, nil
	}

	// Allocate a context
	if r.ctx == nil || r.src_pix_fmt != src.PixelFormat() || r.src_width != src.Width() || r.src_height != src.Height() {
		// Release existing scaling context, if any
		if r.ctx != nil {
			ff.SWScale_free_context(r.ctx)
		}
		// Create a new scaling context
		ctx := ff.SWScale_get_context(
			src.Width(), src.Height(), src.PixelFormat(), // source
			r.dest.Width(), r.dest.Height(), r.dest.PixelFormat(), // destination
			r.flags, nil, nil, nil,
		)
		if ctx == nil {
			return nil, errors.New("failed to allocate swscale context")
		} else {
			r.ctx = ctx
			r.src_pix_fmt = src.PixelFormat()
			r.src_width = src.Width()
			r.src_height = src.Height()
		}
	}

	// Copy parameters from the source frame
	if err := r.dest.CopyPropsFromFrame(src); err != nil {
		return nil, err
	}

	// Rescale the image
	if err := ff.SWScale_scale_frame(r.ctx, (*ff.AVFrame)(r.dest), (*ff.AVFrame)(src), false); err != nil {
		return nil, err
	}

	// Return the destination frame
	return r.dest, nil
}
