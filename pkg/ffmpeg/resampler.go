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
		return nil, errors.New("invalid sample format")
	}
	if par.Samplerate() <= 0 {
		return nil, errors.New("invalid sample rate")
	}
	ch := par.ChannelLayout()
	if !ff.AVUtil_channel_layout_check(&ch) {
		return nil, errors.New("invalid channel layout")
	}

	// Create a destimation frame
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

	// Output buffer size
	var num_samples int
	if src == nil {
		num_samples = int(ff.SWResample_get_delay(r.ctx, int64(r.dest.SampleRate())))
	} else {
		delay := ff.SWResample_get_delay(r.ctx, int64(src.SampleRate())) + int64(src.NumSamples())
		num_samples = int(ff.AVUtil_rescale_rnd(delay, int64(r.dest.SampleRate()), int64(src.SampleRate()), ff.AV_ROUND_UP))
	}
	if num_samples < 0 {
		return nil, errors.New("av_rescale_rnd error")
	}
	if num_samples == 0 {
		return nil, nil
	}

	// Check buffer
	// TODO UGLY CODE ALERT
	if r.dest.NumSamples() < num_samples {
		sample_fmt := r.dest.SampleFormat()
		sample_rate := r.dest.SampleRate()
		sample_ch := r.dest.ChannelLayout()
		r.dest.Unref()
		(*ff.AVFrame)(r.dest).SetSampleFormat(sample_fmt)
		(*ff.AVFrame)(r.dest).SetSampleRate(sample_rate)
		(*ff.AVFrame)(r.dest).SetChannelLayout(sample_ch)
		(*ff.AVFrame)(r.dest).SetNumSamples(num_samples)
		if err := r.dest.AllocateBuffers(); err != nil {
			ff.SWResample_free(r.ctx)
			return nil, err
		}
	}

	// Perform resampling
	if err := ff.SWResample_convert_frame(r.ctx, (*ff.AVFrame)(src), (*ff.AVFrame)(r.dest)); err != nil {
		return nil, fmt.Errorf("SWResample_convert_frame: %w", err)
	}

	// Return the destination frame or nil
	return r.dest, nil
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

	// Return success
	return ctx, nil
}
