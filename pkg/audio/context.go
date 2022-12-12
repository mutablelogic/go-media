package audio

import (
	"runtime"

	// Package imports
	multierror "github.com/hashicorp/go-multierror"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type swcontext struct {
	// Context for resampling
	ctx *ffmpeg.SWRContext
	src AudioFormat

	// Destination audio format
	dest AudioFormat

	// Frame to contain converted data
	frame AudioFrame
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new empty context object
func NewContext(src AudioFrame, dest AudioFormat) (*swcontext, error) {
	var result error

	r := new(swcontext)
	defer func() {
		if result != nil && r.ctx != nil {
			r.ctx.SWR_free()
			r.ctx = nil
		}
	}()

	// Set finalizer to panic if Close() is not called
	runtime.SetFinalizer(r, swcontext_finalizer)

	// Allocate context
	r.ctx = ffmpeg.SWR_alloc()

	// Check in parameter
	if in == nil || in.Samples() == 0 {
		r.ctx.SWR_free()
		r.ctx = nil
		return nil, ErrBadParameter.With("in")
	} else if err := r.setIn(in.AudioFormat()); err != nil {
		r.ctx.SWR_free()
		r.ctx = nil
		return nil, err
	}

	// Copy in parameters to out format
	if dest.Rate == 0 {
		dest.Rate = src.AudioFormat().Rate
	}
	if dest.Format == SAMPLE_FORMAT_NONE {
		dest.Format = src.AudioFormat().Format
	}
	if dest.Layout == CHANNEL_LAYOUT_NONE {
		dest.Layout = src.AudioFormat().Layout
	}
	if err := r.setOut(out); err != nil {
		r.ctx.SWR_free()
		r.ctx = nil
		return nil, err
	}

	// Initialize the context
	if err := r.ctx.SWR_init(); err != nil {
		result = err
		return nil, result
	}

	// Return success
	return r, nil
}

// Free resources associated with the context
func (r *swcontext) Close() error {
	var result error

	// Free frame
	if r.frame != nil {
		if err := r.frame.Close(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	// Free context
	if r.ctx != nil {
		r.ctx.SWR_free()
		r.ctx = nil
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Set the input audio format
func (r *swcontext) setIn(f AudioFormat) error {
	var result error
	if r.ctx == nil {
		return ErrInternalAppError.With("Context is closed")
	}
	if f.Rate != 0 {
		if err := r.ctx.AVUtil_av_opt_set_int("in_sample_rate", int64(f.Rate)); err != nil {
			result = multierror.Append(result, err)
		} else {
			r.src.Rate = f.Rate
		}
	}
	if f.Format != SAMPLE_FORMAT_NONE {
		if fmt := toSampleFormat(f.Format); fmt == ffmpeg.AV_SAMPLE_FMT_NB {
			result = multierror.Append(result, ErrBadParameter.With("Invalid sample format", f.Format))
		} else if err := r.ctx.AVUtil_av_opt_set_sample_fmt("in_sample_fmt", fmt); err != nil {
			result = multierror.Append(result, err)
		} else {
			r.src.Format = f.Format
		}
	}
	if f.Layout != CHANNEL_LAYOUT_NONE {
		l := toChannelLayout(f.Layout)
		if err := r.ctx.AVUtil_av_opt_set_chlayout("in_channel_layout", &l); err != nil {
			result = multierror.Append(result, ErrBadParameter.With("Invalid channel layout:", f.Layout))
		} else {
			r.src.Layout = f.Layout
		}
	}

	// Return any errors
	return result
}

// Set the output audio format
func (r *swcontext) setOut(f AudioFormat) error {
	var result error
	if r.ctx == nil {
		return ErrInternalAppError.With("Context is closed")
	}
	if f.Rate != 0 {
		if err := r.ctx.AVUtil_av_opt_set_int("out_sample_rate", int64(f.Rate)); err != nil {
			result = multierror.Append(result, err)
		} else {
			r.dest.Rate = f.Rate
		}
	}
	if f.Format != SAMPLE_FORMAT_NONE {
		if fmt := toSampleFormat(f.Format); fmt == ffmpeg.AV_SAMPLE_FMT_NB {
			result = multierror.Append(result, ErrBadParameter.With("Invalid sample format", f.Format))
		} else if err := r.ctx.AVUtil_av_opt_set_sample_fmt("out_sample_fmt", fmt); err != nil {
			result = multierror.Append(result, err)
		} else {
			r.dest.Format = f.Format
		}
	}
	if f.Layout != CHANNEL_LAYOUT_NONE {
		l := toChannelLayout(f.Layout)
		if err := r.ctx.AVUtil_av_opt_set_chlayout("out_channel_layout", &l); err != nil {
			result = multierror.Append(result, ErrBadParameter.With("Invalid channel layout:", f.Layout))
		} else {
			r.dest.Layout = f.Layout
		}
	}

	// Return any errors
	return result
}

func swcontext_finalizer(r *swcontext) {
	if r.ctx != nil {
		panic("swresample: context not closed")
	}
}
