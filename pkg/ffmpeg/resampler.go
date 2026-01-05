package ffmpeg

import (
	"errors"
	"fmt"

	// Packages
	media "github.com/mutablelogic/go-media"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// Resampler handles audio resampling or video rescaling for a single stream.
type Resampler struct {
	t     media.Type
	audio *audioResampler
	video *videoRescaler
}

type audioResampler struct {
	ctx   *ff.SWRContext
	dest  *Frame
	force bool
}

type videoRescaler struct {
	ctx   *ff.SWSContext
	dest  *Frame
	flags ff.SWSFlag
	force bool

	srcFmt ff.AVPixelFormat
	srcW   int
	srcH   int
}

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Create a new resampler which will resample or rescale frames to the
// specified parameters
func NewResampler(par *Par, force bool) (*Resampler, error) {
	if par == nil {
		return nil, errors.New("par is nil")
	}
	r := &Resampler{t: par.Type()}
	switch r.t {
	case media.AUDIO:
		a, err := newAudioResampler(par, force)
		if err != nil {
			return nil, err
		}
		r.audio = a
	case media.VIDEO:
		v, err := newVideoRescaler(par, force)
		if err != nil {
			return nil, err
		}
		r.video = v
	default:
		return nil, fmt.Errorf("unsupported type: %v", par.Type())
	}
	return r, nil
}

// Release resources
func (r *Resampler) Close() error {
	var err error
	if r.audio != nil {
		err = errors.Join(err, r.audio.Close())
	}
	if r.video != nil {
		err = errors.Join(err, r.video.Close())
	}
	r.audio = nil
	r.video = nil
	return err
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Resample converts one frame and emits zero or more frames via fn.
// Pass src==nil to flush. fn may receive nil to signal end-of-batch.
func (r *Resampler) Resample(src *Frame, fn func(*Frame) error) error {
	if fn == nil {
		return media.ErrBadParameter.With("nil callback")
	}
	switch r.t {
	case media.AUDIO:
		return r.audio.process(src, fn)
	case media.VIDEO:
		return r.video.process(src, fn)
	default:
		return fmt.Errorf("unsupported type: %v", r.t)
	}
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - AUDIO

func newAudioResampler(par *Par, force bool) (*audioResampler, error) {
	if par.CodecType() != ff.AVMEDIA_TYPE_AUDIO {
		return nil, errors.New("invalid codec type")
	}
	if par.SampleFormat() == ff.AV_SAMPLE_FMT_NONE {
		return nil, errors.New("invalid sample format")
	}
	if par.SampleRate() <= 0 {
		return nil, errors.New("invalid sample rate")
	}
	ch := par.ChannelLayout()
	if !ff.AVUtil_channel_layout_check(&ch) {
		return nil, errors.New("invalid channel layout")
	}

	dest, err := NewFrame(par)
	if err != nil {
		return nil, err
	}

	return &audioResampler{dest: dest, force: force}, nil
}

func (r *audioResampler) Close() error {
	if r.ctx != nil {
		ff.SWResample_free(r.ctx)
	}
	err := r.dest.Close()
	r.ctx = nil
	r.dest = nil
	return err
}

func (r *audioResampler) process(src *Frame, fn func(*Frame) error) error {
	if src != nil && src.Type() != media.AUDIO {
		return errors.New("frame type mismatch")
	}

	// Fast path when formats already match and not forced.
	if src != nil && !r.force && src.MatchesFormat(r.dest) {
		return fn(src)
	}

	if r.ctx == nil {
		if src == nil {
			return nil
		}
		ctx, err := newSWR(r.dest, src)
		if err != nil {
			return err
		}
		r.ctx = ctx
	}

	numSamples, err := r.outputSamples(src)
	if err != nil || numSamples == 0 {
		return err
	}

	if err := r.ensureCapacity(numSamples); err != nil {
		return err
	}

	if err := ff.SWResample_convert_frame(r.ctx, (*ff.AVFrame)(src), (*ff.AVFrame)(r.dest)); err != nil {
		return fmt.Errorf("SWResample_convert_frame: %w", err)
	}

	if src != nil && src.Pts() != int64(ff.AV_NOPTS_VALUE) {
		r.dest.SetPts(ff.AVUtil_rational_rescale_q(src.Pts(), src.TimeBase(), r.dest.TimeBase()))
	}

	if r.dest.NumSamples() == 0 {
		return nil
	}
	return fn(r.dest)
}

func (r *audioResampler) outputSamples(src *Frame) (int, error) {
	if src == nil {
		return int(ff.SWResample_get_delay(r.ctx, int64(r.dest.SampleRate()))), nil
	}
	delay := ff.SWResample_get_delay(r.ctx, int64(src.SampleRate())) + int64(src.NumSamples())
	samples := ff.AVUtil_rescale_rnd(delay, int64(r.dest.SampleRate()), int64(src.SampleRate()), ff.AV_ROUND_UP)
	if samples < 0 {
		return 0, errors.New("av_rescale_rnd error")
	}
	return int(samples), nil
}

func (r *audioResampler) ensureCapacity(numSamples int) error {
	if r.dest.NumSamples() >= numSamples {
		return nil
	}
	fmt := r.dest.SampleFormat()
	rate := r.dest.SampleRate()
	layout := r.dest.ChannelLayout()
	tb := r.dest.TimeBase()
	pts := r.dest.Pts()

	r.dest.Unref()
	(*ff.AVFrame)(r.dest).SetSampleFormat(fmt)
	(*ff.AVFrame)(r.dest).SetSampleRate(rate)
	(*ff.AVFrame)(r.dest).SetChannelLayout(layout)
	(*ff.AVFrame)(r.dest).SetNumSamples(numSamples)
	(*ff.AVFrame)(r.dest).SetTimeBase(tb)
	(*ff.AVFrame)(r.dest).SetPts(pts)

	return r.dest.AllocateBuffers()
}

func newSWR(dest, src *Frame) (*ff.SWRContext, error) {
	ctx := ff.SWResample_alloc()
	if ctx == nil {
		return nil, errors.New("failed to allocate resampler")
	}

	if err := ff.SWResample_set_opts(ctx,
		dest.ChannelLayout(), dest.SampleFormat(), dest.SampleRate(),
		src.ChannelLayout(), src.SampleFormat(), src.SampleRate(),
	); err != nil {
		ff.SWResample_free(ctx)
		return nil, fmt.Errorf("SWResample_set_opts: %w", err)
	}
	if err := ff.SWResample_init(ctx); err != nil {
		ff.SWResample_free(ctx)
		return nil, fmt.Errorf("SWResample_init: %w", err)
	}
	return ctx, nil
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - VIDEO

func newVideoRescaler(par *Par, force bool) (*videoRescaler, error) {
	if par.CodecType() != ff.AVMEDIA_TYPE_VIDEO {
		return nil, errors.New("invalid codec type")
	}
	if par.PixelFormat() == ff.AV_PIX_FMT_NONE {
		return nil, errors.New("invalid pixel format")
	}
	if par.Width() == 0 || par.Height() == 0 {
		return nil, errors.New("invalid width/height")
	}

	dest, err := NewFrame(par)
	if err != nil {
		return nil, err
	}
	if err := dest.AllocateBuffers(); err != nil {
		return nil, errors.Join(err, dest.Close())
	}

	return &videoRescaler{dest: dest, flags: ff.SWS_BILINEAR, force: force}, nil
}

func (r *videoRescaler) Close() error {
	if r.ctx != nil {
		ff.SWScale_free_context(r.ctx)
	}
	err := r.dest.Close()
	r.ctx = nil
	r.dest = nil
	return err
}

func (r *videoRescaler) process(src *Frame, fn func(*Frame) error) error {
	if src != nil && src.Type() != media.VIDEO {
		return errors.New("frame type mismatch")
	}
	if src == nil {
		return fn(nil)
	}
	if !r.force && src.MatchesFormat(r.dest) {
		return fn(src)
	}

	if r.ctx == nil || r.srcFmt != src.PixelFormat() || r.srcW != src.Width() || r.srcH != src.Height() {
		if r.ctx != nil {
			ff.SWScale_free_context(r.ctx)
		}
		ctx := ff.SWScale_get_context(
			src.Width(), src.Height(), src.PixelFormat(),
			r.dest.Width(), r.dest.Height(), r.dest.PixelFormat(),
			r.flags, nil, nil, nil,
		)
		if ctx == nil {
			return errors.New("failed to allocate swscale context")
		}
		r.ctx = ctx
		r.srcFmt = src.PixelFormat()
		r.srcW = src.Width()
		r.srcH = src.Height()
	}

	if err := r.dest.CopyPropsFromFrame(src); err != nil {
		return err
	}
	if err := ff.SWScale_scale_frame(r.ctx, (*ff.AVFrame)(r.dest), (*ff.AVFrame)(src), false); err != nil {
		return err
	}
	return fn(r.dest)
}
