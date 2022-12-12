package media

import (
	// Packages
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	ffmpeg "github.com/mutablelogic/go-media/sys/ffmpeg51"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
	. "github.com/mutablelogic/go-media"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type resampler struct {
	ctx  *ffmpeg.SWRContext
	src  AudioFormat
	dest AudioFormat
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new resampler for a frame
func NewResampler(src, dest AudioFormat) *resampler {
	this := new(resampler)

	// source needs to have all parameters set
	if src.Format == SAMPLE_FORMAT_NONE || src.Layout == CHANNEL_LAYOUT_NONE || src.Rate == 0 {
		return nil
	}

	// copy parameters over from src to dest
	if dest.Rate == 0 {
		dest.Rate = src.Rate
	}
	if dest.Format == SAMPLE_FORMAT_NONE {
		dest.Format = src.Format
	}
	if dest.Layout == CHANNEL_LAYOUT_NONE {
		dest.Layout = src.Layout
	}

	// Allocate context
	if ctx := ffmpeg.SWR_alloc(); ctx == nil {
		return nil
	} else {
		this.ctx = ctx
	}

	// Set context source
	var result error
	if err := this.set("in", src); err != nil {
		result = multierror.Append(result, err)
	}
	if err := this.set("out", dest); err != nil {
		result = multierror.Append(result, err)
	}
	if result != nil {
		ffmpeg.SWR_free(this.ctx)
		this.ctx = nil
		return nil
	}

	// Initialize the context
	if err := ffmpeg.SWR_init(this.ctx); err != nil {
		ffmpeg.SWR_free(this.ctx)
		this.ctx = nil
		return nil
	}

	// Set parameters
	this.src = src
	this.dest = dest

	// Return success
	return this
}

func (resampler *resampler) Close() error {
	var result error

	// Release context
	if resampler.ctx != nil {
		ffmpeg.SWR_free(resampler.ctx)
	}

	// Blank out other fields
	resampler.ctx = nil
	resampler.src.Format = SAMPLE_FORMAT_NONE
	resampler.dest.Format = SAMPLE_FORMAT_NONE

	// Return success
	return result
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (resampler *resampler) String() string {
	str := "<media.resampler"
	if resampler.src.Format != SAMPLE_FORMAT_NONE {
		str += fmt.Sprint(" src=", resampler.src)
	}
	if resampler.dest.Format != SAMPLE_FORMAT_NONE {
		str += fmt.Sprint(" dest=", resampler.dest)
	}
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (resampler *resampler) Resample(src, dest Frame) error {
	if resampler.ctx == nil {
		return ErrInternalAppError.With("resampler")
	}
	if src == nil || !src.Flags().Is(MEDIA_FLAG_AUDIO) {
		return ErrBadParameter.With("src")
	}
	if dest == nil || !dest.Flags().Is(MEDIA_FLAG_AUDIO) {
		return ErrBadParameter.With("dest")
	}

	// Do conversion
	return ffmpeg.SWR_convert_frame(resampler.ctx, dest.(*frame).ctx, src.(*frame).ctx)
}

func (resampler *resampler) Flush() error {
	if resampler.ctx == nil {
		return ErrInternalAppError.With("resampler")
	}

	// Do flush
	return ffmpeg.SWR_convert_frame(resampler.ctx, nil, nil)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// Set the context options. Prefix should be "in" or "out" depending on whether
// the source or destination format is being set.
func (resampler *resampler) set(prefix string, f AudioFormat) error {
	var result error
	if resampler.ctx == nil {
		return ErrInternalAppError.With("Context is closed")
	}
	if f.Rate != 0 {
		if err := resampler.ctx.AVUtil_av_opt_set_int(prefix+"_sample_rate", int64(f.Rate)); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if f.Format != SAMPLE_FORMAT_NONE {
		if fmt := toSampleFormat(f.Format); fmt == ffmpeg.AV_SAMPLE_FMT_NB {
			result = multierror.Append(result, ErrBadParameter.With("Invalid sample format", f.Format))
		} else if err := resampler.ctx.AVUtil_av_opt_set_sample_fmt(prefix+"_sample_fmt", fmt); err != nil {
			result = multierror.Append(result, err)
		}
	}
	if f.Layout != CHANNEL_LAYOUT_NONE {
		l := toChannelLayout(f.Layout)
		if err := resampler.ctx.AVUtil_av_opt_set_chlayout(prefix+"_channel_layout", &l); err != nil {
			result = multierror.Append(result, ErrBadParameter.With("Invalid channel layout:", f.Layout))
		}
	}

	// Return any errors
	return result
}
