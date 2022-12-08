package swresample

import (
	"runtime"

	// Package imports
	"github.com/hashicorp/go-multierror"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

// Package imports

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
	runtime.SetFinalizer(r, finalizer)
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

func finalizer(r *swcontext) {
	if r.ctx != nil {
		panic("swresample: context not closed")
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func toChannelLayout(v ChannelLayout) ffmpeg.AVChannelLayout {
	switch v {
	case CHANNEL_LAYOUT_MONO:
		return ffmpeg.AV_CHANNEL_LAYOUT_MONO
	case CHANNEL_LAYOUT_STEREO:
		return ffmpeg.AV_CHANNEL_LAYOUT_STEREO
	case CHANNEL_LAYOUT_2POINT1:
		return ffmpeg.AV_CHANNEL_LAYOUT_2POINT1
	case CHANNEL_LAYOUT_2_1:
		return ffmpeg.AV_CHANNEL_LAYOUT_2_1
	case CHANNEL_LAYOUT_SURROUND:
		return ffmpeg.AV_CHANNEL_LAYOUT_SURROUND
	case CHANNEL_LAYOUT_3POINT1:
		return ffmpeg.AV_CHANNEL_LAYOUT_3POINT1
	case CHANNEL_LAYOUT_4POINT0:
		return ffmpeg.AV_CHANNEL_LAYOUT_4POINT0
	case CHANNEL_LAYOUT_4POINT1:
		return ffmpeg.AV_CHANNEL_LAYOUT_4POINT1
	case CHANNEL_LAYOUT_2_2:
		return ffmpeg.AV_CHANNEL_LAYOUT_2_2
	case CHANNEL_LAYOUT_QUAD:
		return ffmpeg.AV_CHANNEL_LAYOUT_QUAD
	case CHANNEL_LAYOUT_5POINT0:
		return ffmpeg.AV_CHANNEL_LAYOUT_5POINT0
	case CHANNEL_LAYOUT_5POINT1:
		return ffmpeg.AV_CHANNEL_LAYOUT_5POINT1
	case CHANNEL_LAYOUT_5POINT0_BACK:
		return ffmpeg.AV_CHANNEL_LAYOUT_5POINT0_BACK
	case CHANNEL_LAYOUT_5POINT1_BACK:
		return ffmpeg.AV_CHANNEL_LAYOUT_5POINT1_BACK
	case CHANNEL_LAYOUT_6POINT0:
		return ffmpeg.AV_CHANNEL_LAYOUT_6POINT0
	case CHANNEL_LAYOUT_6POINT0_FRONT:
		return ffmpeg.AV_CHANNEL_LAYOUT_6POINT0_FRONT
	case CHANNEL_LAYOUT_HEXAGONAL:
		return ffmpeg.AV_CHANNEL_LAYOUT_HEXAGONAL
	case CHANNEL_LAYOUT_6POINT1:
		return ffmpeg.AV_CHANNEL_LAYOUT_6POINT1
	case CHANNEL_LAYOUT_6POINT1_BACK:
		return ffmpeg.AV_CHANNEL_LAYOUT_6POINT1_BACK
	case CHANNEL_LAYOUT_6POINT1_FRONT:
		return ffmpeg.AV_CHANNEL_LAYOUT_6POINT1_FRONT
	case CHANNEL_LAYOUT_7POINT0:
		return ffmpeg.AV_CHANNEL_LAYOUT_7POINT0
	case CHANNEL_LAYOUT_7POINT0_FRONT:
		return ffmpeg.AV_CHANNEL_LAYOUT_7POINT0_FRONT
	case CHANNEL_LAYOUT_7POINT1:
		return ffmpeg.AV_CHANNEL_LAYOUT_7POINT1
	case CHANNEL_LAYOUT_7POINT1_WIDE:
		return ffmpeg.AV_CHANNEL_LAYOUT_7POINT1_WIDE
	case CHANNEL_LAYOUT_7POINT1_WIDE_BACK:
		return ffmpeg.AV_CHANNEL_LAYOUT_7POINT1_WIDE_BACK
	case CHANNEL_LAYOUT_OCTAGONAL:
		return ffmpeg.AV_CHANNEL_LAYOUT_OCTAGONAL
	case CHANNEL_LAYOUT_STEREO_DOWNMIX:
		return ffmpeg.AV_CHANNEL_LAYOUT_STEREO_DOWNMIX
	case CHANNEL_LAYOUT_22POINT2:
		return ffmpeg.AV_CHANNEL_LAYOUT_22POINT2
	case CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER:
		return ffmpeg.AV_CHANNEL_LAYOUT_AMBISONIC_FIRST_ORDER
	default:
		return ffmpeg.AV_CHANNEL_LAYOUT_MONO
	}
}

func toSampleFormat(v SampleFormat) ffmpeg.AVSampleFormat {
	switch v {
	case SAMPLE_FORMAT_NONE:
		return ffmpeg.AV_SAMPLE_FMT_NONE
	case SAMPLE_FORMAT_U8:
		return ffmpeg.AV_SAMPLE_FMT_U8
	case SAMPLE_FORMAT_S16:
		return ffmpeg.AV_SAMPLE_FMT_S16
	case SAMPLE_FORMAT_S32:
		return ffmpeg.AV_SAMPLE_FMT_S32
	case SAMPLE_FORMAT_FLT:
		return ffmpeg.AV_SAMPLE_FMT_FLT
	case SAMPLE_FORMAT_DBL:
		return ffmpeg.AV_SAMPLE_FMT_DBL
	case SAMPLE_FORMAT_U8P:
		return ffmpeg.AV_SAMPLE_FMT_U8P
	case SAMPLE_FORMAT_S16P:
		return ffmpeg.AV_SAMPLE_FMT_S16P
	case SAMPLE_FORMAT_S32P:
		return ffmpeg.AV_SAMPLE_FMT_S32P
	case SAMPLE_FORMAT_FLTP:
		return ffmpeg.AV_SAMPLE_FMT_FLTP
	case SAMPLE_FORMAT_DBLP:
		return ffmpeg.AV_SAMPLE_FMT_DBLP
	case SAMPLE_FORMAT_S64:
		return ffmpeg.AV_SAMPLE_FMT_S64
	case SAMPLE_FORMAT_S64P:
		return ffmpeg.AV_SAMPLE_FMT_S64P
	default:
		return ffmpeg.AV_SAMPLE_FMT_NB
	}
}
