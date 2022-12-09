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
	ctx *ffmpeg.SWRContext
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new empty context object
func NewContext() *swcontext {
	r := new(swcontext)
	r.ctx = ffmpeg.SWR_alloc()
	runtime.SetFinalizer(r, swcontext_finalizer)
	return r
}

// Free resources associated with the context
func (r *swcontext) Close() error {
	var result error

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

// Set the input audio format
func (r *swcontext) SetIn(f AudioFormat) error {
	var result error
	if r.ctx == nil {
		return ErrInternalAppError.With("Context is closed")
	}
	if f.Rate != 0 {
		if err := r.ctx.AVUtil_av_opt_set_int("in_sample_rate", int64(f.Rate)); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if f.Format != SAMPLE_FORMAT_NONE {
		if fmt := toSampleFormat(f.Format); fmt == ffmpeg.AV_SAMPLE_FMT_NB {
			result = multierror.Append(result, ErrBadParameter.With("Invalid sample format", f.Format))
		} else if err := r.ctx.AVUtil_av_opt_set_sample_fmt("in_sample_fmt", fmt); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if f.Layout != CHANNEL_LAYOUT_NONE {
		l := toChannelLayout(f.Layout)
		if err := r.ctx.AVUtil_av_opt_set_chlayout("in_channel_layout", &l); err != nil {
			result = multierror.Append(result, ErrBadParameter.With("Invalid channel layout:", f.Layout))
		}
	}

	// Return any errors
	return result
}

// Set the output audio format
func (r *swcontext) SetOut(f AudioFormat) error {
	var result error
	if r.ctx == nil {
		return ErrInternalAppError.With("Context is closed")
	}
	if f.Rate != 0 {
		if err := r.ctx.AVUtil_av_opt_set_int("out_sample_rate", int64(f.Rate)); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if f.Format != SAMPLE_FORMAT_NONE {
		if fmt := toSampleFormat(f.Format); fmt == ffmpeg.AV_SAMPLE_FMT_NB {
			result = multierror.Append(result, ErrBadParameter.With("Invalid sample format", f.Format))
		} else if err := r.ctx.AVUtil_av_opt_set_sample_fmt("out_sample_fmt", fmt); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if f.Layout != CHANNEL_LAYOUT_NONE {
		l := toChannelLayout(f.Layout)
		if err := r.ctx.AVUtil_av_opt_set_chlayout("out_channel_layout", &l); err != nil {
			result = multierror.Append(result, ErrBadParameter.With("Invalid channel layout:", f.Layout))
		}
	}

	// Return any errors
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (r *swcontext) initialize() error {
	// Check parameters
	if r.ctx == nil {
		return ErrInternalAppError.With("Context is closed")
	} else if r.ctx.SWR_is_initialized() {
		return ErrOutOfOrder.With("Context already initialized")
	}
	return r.ctx.SWR_init()
}

func swcontext_finalizer(r *swcontext) {
	if r.ctx != nil {
		panic("swresample: context not closed")
	}
}
