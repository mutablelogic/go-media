package ffmpeg

import (
	"errors"
	"fmt"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type resampler struct {
	ctx   *ff.SWRContext
	dest  *Frame
	force bool
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new audio resampler which will resample the input frame to the
// specified channel layout, sample rate and sample format.
func NewResampler(par *Par, force bool) (*resampler, error) {
	resampler := new(resampler)

	// Check parameters
	if par == nil || par.CodecType() != ff.AVMEDIA_TYPE_AUDIO {
		return nil, errors.New("invalid codec type")
	}
	if par.SampleFormat() == ff.AV_SAMPLE_FMT_NONE {
		return nil, errors.New("invalid sample format parameters")
	}
	ch := par.ChannelLayout()
	if !ff.AVUtil_channel_layout_check(&ch) {
		return nil, errors.New("invalid channel layout parameters")
	}

	// Create a destimation frame
	// Set parameters - we don't allocate the buffer here,
	// we do that when we have a source frame and know how
	// large the destination frame should be
	dest, err := NewFrame(par)
	if err != nil {
		return nil, err
	}

	// Set parameters
	resampler.dest = dest
	resampler.force = force

	// Return success
	return resampler, nil
}

// Release resources
func (r *resampler) Close() error {
	if r.ctx != nil {
		ff.SWResample_free(r.ctx)
		r.ctx = nil
	}
	result := r.dest.Close()
	r.dest = nil
	return result
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Resample the source and return the destination
func (r *resampler) Frame(src *Frame) (*Frame, error) {
	// Simply return the frame if it matches the destination format
	if src != nil {
		if src.matchesResampleResize(r.dest) && !r.force {
			return src, nil
		}
	}

	// Create a resampling context
	if r.ctx == nil {
		if src == nil {
			return nil, nil
		}
		ctx, err := newResampler(r.dest, src)
		if err != nil {
			return nil, err
		} else {
			r.ctx = ctx
		}
	}

	// Copy parameters from the source frame
	if src != nil {
		if err := r.dest.CopyPropsFromFrame(src); err != nil {
			return nil, err
		}
	}

	// Get remaining samples
	if src == nil {
		if samples := ff.SWResample_get_delay(r.ctx, int64(r.dest.SampleRate())); samples > 0 {
			fmt.Println("TODO: SWResample_get_delay remaining samples=", samples)
		}
	}

	// Perform resampling
	if err := ff.SWResample_convert_frame(r.ctx, (*ff.AVFrame)(src), (*ff.AVFrame)(r.dest)); err != nil {
		return nil, fmt.Errorf("SWResample_convert_frame: %w", err)
	}

	// Return the destination frame or nil
	if r.dest.NumSamples() == 0 {
		return nil, nil
	} else {
		return r.dest, nil
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func newResampler(dest, src *Frame) (*ff.SWRContext, error) {
	// Create a new resampler
	ctx := ff.SWResample_alloc()
	if ctx == nil {
		return nil, errors.New("failed to allocate resampler")
	}

	// Set options to covert from the codec frame to the decoder frame
	if err := ff.SWResample_set_opts(ctx,
		dest.ChannelLayout(), dest.SampleFormat(), dest.SampleRate(), // destination
		src.ChannelLayout(), src.SampleFormat(), src.SampleRate(), // source
	); err != nil {
		ff.SWResample_free(ctx)
		return nil, fmt.Errorf("SWResample_set_opts: %w", err)
	}

	// Initialize the resampling context
	if err := ff.SWResample_init(ctx); err != nil {
		ff.SWResample_free(ctx)
		return nil, fmt.Errorf("SWResample_init: %w", err)
	}

	// Make a copy of the destination frame parameters
	sample_fmt := dest.SampleFormat()
	sample_rate := dest.SampleRate()
	sample_ch := dest.ChannelLayout()

	// Unreference the current frame
	dest.Unref()

	// Set the number of samples
	if dest_samples, err := ff.SWResample_get_out_samples(ctx, src.NumSamples()); err != nil {
		ff.SWResample_free(ctx)
		return nil, fmt.Errorf("SWResample_get_out_samples: %w", err)
	} else if dest_samples == 0 {
		ff.SWResample_free(ctx)
		return nil, fmt.Errorf("SWResample_get_out_samples: number of samples is zero")
	} else {
		dest.SetSampleFormat(sample_fmt)
		dest.SetSampleRate(sample_rate)
		dest.SetChannelLayout(sample_ch)
		dest.SetNumSamples(dest_samples)
	}

	// Create buffers
	if err := dest.AllocateBuffers(); err != nil {
		ff.SWResample_free(ctx)
		return nil, err
	}

	// Return success
	return ctx, nil
}
