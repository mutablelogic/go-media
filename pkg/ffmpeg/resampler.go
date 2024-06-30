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
	dest  *ff.AVFrame
	force bool
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new audio resampler which will resample the input frame to the
// specified channel layout, sample rate and sample format.
func NewResampler(format ff.AVSampleFormat, opt ...Opt) (*resampler, error) {
	options := newOpts()
	resampler := new(resampler)

	// Apply options
	options.par.SetCodecType(ff.AVMEDIA_TYPE_AUDIO)
	options.par.SetSampleFormat(format)
	options.par.SetChannelLayout(ff.AV_CHANNEL_LAYOUT_MONO)
	options.par.SetSamplerate(44100)
	for _, o := range opt {
		if err := o(options); err != nil {
			return nil, err
		}
	}

	// Check parameters
	if options.par.SampleFormat() == ff.AV_SAMPLE_FMT_NONE {
		return nil, errors.New("invalid sample format parameters")
	}
	ch := options.par.ChannelLayout()
	if !ff.AVUtil_channel_layout_check(&ch) {
		return nil, errors.New("invalid channel layout parameters")
	}

	// Create a destimation frame
	dest := ff.AVUtil_frame_alloc()
	if dest == nil {
		return nil, errors.New("failed to allocate frame")
	}

	// Set parameters - we don't allocate the buffer here,
	// we do that when we have a source frame and know how
	// large the destination frame should be
	dest.SetSampleFormat(options.par.SampleFormat())
	dest.SetSampleRate(options.par.Samplerate())
	if err := dest.SetChannelLayout(options.par.ChannelLayout()); err != nil {
		ff.AVUtil_frame_free(dest)
		return nil, err
	} else {
		resampler.dest = dest
	}

	// Set force flag
	resampler.force = options.force

	// Return success
	return resampler, nil
}

// Release resources
func (r *resampler) Close() error {
	if r.ctx != nil {
		ff.SWResample_free(r.ctx)
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

// Resample the source and return the destination
func (r *resampler) Frame(src *ff.AVFrame) (*ff.AVFrame, error) {
	// Simply return the frame if it matches the destination format
	if src != nil {
		if matchesAudioFormat(src, r.dest) && !r.force {
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
		if err := ff.AVUtil_frame_copy_props(r.dest, src); err != nil {
			return nil, err
		}
	}

	// Get remaining samples
	if src == nil {
		samples := ff.SWResample_get_delay(r.ctx, int64(r.dest.SampleRate()))
		fmt.Println("TODO: remaining samples=", samples)
	}

	// Perform resampling
	if err := ff.SWResample_convert_frame(r.ctx, src, r.dest); err != nil {
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

// Returns true if the sample format, channels and sample rate of the source
// and destination frames match.
func matchesAudioFormat(src, dest *ff.AVFrame) bool {
	if src.SampleFormat() != dest.SampleFormat() {
		return false
	}
	if src.SampleRate() != dest.SampleRate() {
		return false
	}
	a := src.ChannelLayout()
	b := dest.ChannelLayout()
	return ff.AVUtil_channel_layout_compare(&a, &b)
}

func newResampler(dest, src *ff.AVFrame) (*ff.SWRContext, error) {
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

	sample_fmt := dest.SampleFormat()
	sample_rate := dest.SampleRate()
	sample_ch := dest.ChannelLayout()

	// Unreference the current frame
	ff.AVUtil_frame_unref(dest)

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
	if err := ff.AVUtil_frame_get_buffer(dest, false); err != nil {
		ff.SWResample_free(ctx)
		return nil, fmt.Errorf("AVUtil_frame_get_buffer: %w", err)
	}

	// Return success
	return ctx, nil
}
