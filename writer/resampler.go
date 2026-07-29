package writer

import (
	// Packages
	gomedia "github.com/mutablelogic/go-media"
	frame "github.com/mutablelogic/go-media/frame"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

// resampler converts frames arriving for one stream into the exact format
// its codec was opened with, so Writer.Encode never hands Encoder.Encode a
// mismatched frame.
type Resampler struct {
	audio *audioResampler
	video *videoRescaler
}

// audioResampler wraps libswresample to convert sample format/rate/channel
// layout and, if frameSize > 0, to re-chunk output into exactly frameSize
// samples per frame (all but the last).
type audioResampler struct {
	par       *ff.AVCodecParameters
	frameSize int // 0 = codec accepts any frame size

	ctx      *ff.SWRContext
	dest     *frame.AudioFrame // frameSize<=0: swr's direct output target. frameSize>0: accumulation buffer holding pending samples not yet emitted
	pts      int64
	capacity int // dest's buffer capacity, in samples

	// frameSize > 0 only: swr output is re-chunked to exactly frameSize
	// samples per frame, which may take several process() calls to fill -
	// scratch is swr's per-call output target, copied sample-by-sample into
	// dest's [pending, pending+n) range until dest fills to a full chunk.
	scratch *frame.AudioFrame
	pending int

	// srcFmt/srcRate/srcLayout are what ctx is currently configured to
	// convert FROM - compared against each incoming frame to detect an
	// upstream source change (e.g. splicing in a different feed mid-stream)
	// so ctx can be drained and rebuilt for the new source rather than
	// silently relying on swr's own internal reconfigure-on-format-change,
	// which isn't guaranteed to preserve samples still buffered in the old
	// context's resampling delay line.
	//
	// Comparing srcLayout needs care: audioResampler (this struct) also
	// holds live pointer fields (ctx/dest/scratch), so &r.srcLayout is an
	// interior pointer into an allocation that legitimately contains Go
	// pointers elsewhere. Go's cgo pointer check resolves an interior
	// pointer to its whole containing allocation and scans all of it, so
	// passing &r.srcLayout straight into a cgo call panics on those
	// unrelated neighbors - not on srcLayout's own (harmless) content. Copy
	// it into an independent local first (see sourceChanged) before taking
	// its address.
	srcFmt    ff.AVSampleFormat
	srcRate   int
	srcLayout ff.AVChannelLayout
}

// videoRescaler wraps libswscale to convert pixel format and/or dimensions.
// It also owns the output pts, counting frames at a constant timebase
// rather than trusting src's own pts/timebase
type videoRescaler struct {
	par      *ff.AVCodecParameters
	flags    ff.SWSFlag
	timebase ff.AVRational

	ctx    *ff.SWSContext
	dest   *frame.VideoFrame
	srcFmt ff.AVPixelFormat
	srcW   int
	srcH   int
	pts    int64
}

//////////////////////////////////////////////////////////////////////////////
// GLOBALS

// defaultResampleCapacity sizes the destination buffer for an audio stream
// whose codec accepts any frame size (FrameSize == 0) - there's no fixed
// target to size it to, so just pick something.
const defaultResampleCapacity = 8192

//////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewResampler creates a resampler that converts audio or video frames into
// par's target format. For any other codec type (e.g. subtitle, data), it returns
// a passthrough Resampler. frameSize only applies to audio (see audioResampler);
// timebase only applies to video (see videoRescaler) - it's the fixed output
// timebase (typically 1/framerate) frames are counted out in, so switching
// source mid-stream can't hand the encoder a pts in the wrong timebase or one
// that jumps backwards relative to the previous source.
func NewResampler(par *ff.AVCodecParameters, frameSize int, timebase ff.AVRational) (*Resampler, error) {
	if par == nil {
		return nil, gomedia.ErrBadParameter.With("resampler: nil codec parameters")
	}
	switch par.CodecType() {
	case ff.AVMEDIA_TYPE_AUDIO:
		return newAudioResampler(par, frameSize)
	case ff.AVMEDIA_TYPE_VIDEO:
		return newVideoRescaler(par, timebase)
	default:
		return &Resampler{}, nil
	}
}

func newAudioResampler(par *ff.AVCodecParameters, frameSize int) (*Resampler, error) {
	if par.CodecType() != ff.AVMEDIA_TYPE_AUDIO {
		return nil, gomedia.ErrBadParameter.With("resampler: invalid target codec type")
	}
	if par.SampleFormat() == ff.AV_SAMPLE_FMT_NONE {
		return nil, gomedia.ErrBadParameter.With("resampler: invalid target sample format")
	}
	if par.SampleRate() <= 0 {
		return nil, gomedia.ErrBadParameter.With("resampler: invalid target sample rate")
	}
	layout := par.ChannelLayout()
	if !ff.AVUtil_channel_layout_check(&layout) {
		return nil, gomedia.ErrBadParameter.With("resampler: invalid target channel layout")
	}
	return &Resampler{audio: &audioResampler{par: par, frameSize: frameSize}}, nil
}

func newVideoRescaler(par *ff.AVCodecParameters, timebase ff.AVRational) (*Resampler, error) {
	if par.CodecType() != ff.AVMEDIA_TYPE_VIDEO {
		return nil, gomedia.ErrBadParameter.With("rescaler: invalid target codec type")
	}
	if par.PixelFormat() == ff.AV_PIX_FMT_NONE {
		return nil, gomedia.ErrBadParameter.With("rescaler: invalid target pixel format")
	}
	if par.Width() <= 0 || par.Height() <= 0 {
		return nil, gomedia.ErrBadParameter.With("rescaler: invalid target width/height")
	}
	if timebase.Num() <= 0 || timebase.Den() <= 0 {
		return nil, gomedia.ErrBadParameter.With("rescaler: invalid target timebase")
	}
	return &Resampler{video: &videoRescaler{par: par, flags: ff.SWS_BILINEAR, timebase: timebase}}, nil
}

func (r *audioResampler) Close() error {
	if r.ctx != nil {
		ff.SWResample_free(r.ctx)
		r.ctx = nil
	}
	var err error
	if r.dest != nil {
		err = r.dest.Close()
		r.dest = nil
	}
	if r.scratch != nil {
		if serr := r.scratch.Close(); err == nil {
			err = serr
		}
		r.scratch = nil
	}
	return err
}

func (r *videoRescaler) Close() error {
	if r.ctx != nil {
		ff.SWScale_free_context(r.ctx)
		r.ctx = nil
	}
	var err error
	if r.dest != nil {
		err = r.dest.Close()
		r.dest = nil
	}
	return err
}

func (r *Resampler) Close() error {
	switch {
	case r.audio != nil:
		return r.audio.Close()
	case r.video != nil:
		return r.video.Close()
	default:
		return nil
	}
}

//////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Process converts src and calls fn for every resulting frame. Pass
// src==nil to flush - audio may have buffered samples still to emit; video
// and passthrough have nothing to flush and this is a no-op for them.
func (r *Resampler) Process(src frame.Frame, fn func(frame.Frame) error) error {
	if fn == nil {
		return gomedia.ErrBadParameter.With("resampler: nil callback function")
	}

	switch {
	case r.audio != nil:
		audioSrc, ok := src.(*frame.AudioFrame)
		if src != nil && !ok {
			return gomedia.ErrBadParameter.Withf("resampler: expected an audio frame, got %T", src)
		}
		return r.audio.process(audioSrc, func(f *frame.AudioFrame) error {
			return fn(f)
		})

	case r.video != nil:
		videoSrc, ok := src.(*frame.VideoFrame)
		if src != nil && !ok {
			return gomedia.ErrBadParameter.Withf("resampler: expected a video frame, got %T", src)
		}
		return r.video.process(videoSrc, func(f *frame.VideoFrame) error {
			return fn(f)
		})

	default:
		// Passthrough - nothing to flush, and any non-nil frame goes
		// straight through unmodified.
		if src == nil {
			return nil
		}
		return fn(src)
	}
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - AUDIO

// matches reports whether src can go straight to the encoder untouched.
func (r *audioResampler) matches(src *frame.AudioFrame) bool {
	if src.SampleFormat() != r.par.SampleFormat() || src.SampleRate() != r.par.SampleRate() {
		return false
	}
	srcLayout, dstLayout := src.ChannelLayout(), r.par.ChannelLayout()
	if !ff.AVUtil_channel_layout_compare(&srcLayout, &dstLayout) {
		return false
	}
	return r.frameSize <= 0 || src.NumSamples() == r.frameSize
}

// sourceChanged reports whether src's format no longer matches what ctx was
// built for.
func (r *audioResampler) sourceChanged(src *frame.AudioFrame) bool {
	if r.ctx == nil {
		return false
	}
	if src.SampleFormat() != r.srcFmt || src.SampleRate() != r.srcRate {
		return true
	}
	// Copy r.srcLayout into an independent local before taking its address -
	// see the field's doc comment for why &r.srcLayout itself isn't safe.
	layout, stored := src.ChannelLayout(), r.srcLayout
	return !ff.AVUtil_channel_layout_compare(&layout, &stored)
}

// drain empties whatever ctx currently has buffered, without pushing new
// input - used when switching to a new source mid-stream, so anything still
// in flight in the outgoing context's resampling delay line is captured
// rather than lost. Deliberately not a "flush": pending (a partially
// accumulated chunk) is left exactly as it is, so the new source's samples
// continue filling it rather than being cut short into a shorter chunk here.
func (r *audioResampler) drain(fn func(*frame.AudioFrame) error) error {
	if r.frameSize <= 0 {
		return r.processUnbounded(nil, fn)
	}
	return r.processChunked(nil, false, fn)
}

// process converts src (nil flushes) and calls fn for every resulting frame.
func (r *audioResampler) process(src *frame.AudioFrame, fn func(*frame.AudioFrame) error) error {
	// The fast path only applies with nothing buffered - if a chunk is
	// still pending, this frame must go through accumulation too, or it
	// would be emitted out of order, ahead of the older buffered samples.
	if src != nil && r.pending == 0 && r.matches(src) {
		// Own the pts even here - trusting src's own pts would break
		// continuity the moment the source switches to one numbering its
		// samples differently (or not needing conversion, while a previous
		// source did).
		n := src.NumSamples()
		src.SetPts(r.pts)
		r.pts += int64(n)
		return fn(src)
	}

	if src != nil && r.sourceChanged(src) {
		if err := r.drain(fn); err != nil {
			return err
		}
		ff.SWResample_free(r.ctx)
		r.ctx = nil
	}

	if r.ctx == nil {
		if src == nil {
			return nil // never initialized - nothing buffered to flush
		}
		if err := r.init(src); err != nil {
			return err
		}
	}

	var in *ff.AVFrame
	if src != nil {
		in = src.AVFrame
	}
	if r.frameSize <= 0 {
		return r.processUnbounded(in, fn)
	}
	return r.processChunked(in, src == nil, fn)
}

// processUnbounded is used when the codec accepts any frame size - swr's
// output is passed straight through, one frame per convert call, with no
// re-chunking.
func (r *audioResampler) processUnbounded(in *ff.AVFrame, fn func(*frame.AudioFrame) error) error {
	// First iteration converts in (nil on a flush); every further iteration
	// pulls only from swr's internal buffer, so a single call that produces
	// more than one dest-buffer's worth of output correctly yields >1
	// output frame instead of silently dropping the rest.
	for first := true; ; first = false {
		if !first {
			in = nil
		}

		// Offer the full capacity every call - swr shrinks AVFrame's
		// NumSamples to the actual count written, so it must be reset
		// before each call or output would shrink monotonically. This must
		// come before MakeWritable below: a clone allocates sized to
		// whatever NumSamples currently is, so it has to already be back
		// at full capacity, not left at a previous call's shrunk count.
		r.dest.SetNumSamples(r.capacity)

		// dest may still be referenced by the encoder from a previous
		// fn(r.dest) call - avcodec_send_frame is documented to possibly
		// keep a reference rather than copy, so overwriting it in place
		// here could corrupt a frame the encoder hasn't finished with yet.
		// MakeWritable clones a fresh buffer instead when that's the case.
		if err := r.dest.MakeWritable(); err != nil {
			return err
		}
		if err := ff.SWResample_convert_frame(r.ctx, in, r.dest.AVFrame); err != nil {
			return err
		}

		n := r.dest.NumSamples()
		if n == 0 {
			return nil
		}
		r.dest.SetPts(r.pts)
		r.pts += int64(n)
		if err := fn(r.dest); err != nil {
			return err
		}
	}
}

// processChunked is used when the codec requires exactly frameSize samples
// per frame. swr's output rarely lines up with frameSize on its own, so it's
// converted into a scratch buffer and accumulated into dest, emitting a
// frame each time dest fills to exactly frameSize - across as many calls as
// it takes. On flush, any samples short of a full chunk are emitted as a
// final, shorter frame.
func (r *audioResampler) processChunked(in *ff.AVFrame, flush bool, fn func(*frame.AudioFrame) error) error {
	for first := true; ; first = false {
		if !first {
			in = nil
		}

		r.scratch.SetNumSamples(defaultResampleCapacity)
		if err := ff.SWResample_convert_frame(r.ctx, in, r.scratch.AVFrame); err != nil {
			return err
		}

		n := r.scratch.NumSamples()
		if n == 0 {
			break
		}
		if err := r.accumulate(n, fn); err != nil {
			return err
		}
	}

	if !flush || r.pending == 0 {
		return nil
	}

	// Flush: release the remainder, short of a full chunk.
	r.dest.SetNumSamples(r.pending)
	r.dest.SetPts(r.pts)
	r.pts += int64(r.pending)
	err := fn(r.dest)
	r.pending = 0
	r.dest.SetNumSamples(r.capacity)
	return err
}

// accumulate copies n newly-converted samples from r.scratch into r.dest at
// the current pending offset, emitting (and resetting) every time dest
// fills to a full frameSize chunk - a single scratch drain can complete more
// than one pending chunk, so this loops until every sample is placed.
func (r *audioResampler) accumulate(n int, fn func(*frame.AudioFrame) error) error {
	for off := 0; off < n; {
		take := min(n-off, r.frameSize-r.pending)

		// dest may still be referenced by the encoder from a previous
		// fn(r.dest) call - see processUnbounded for why this must happen
		// before writing into it. dest's NumSamples never changes here
		// (always r.capacity), so no reordering concern like
		// processUnbounded's - just needs to happen before the copy below.
		if err := r.dest.MakeWritable(); err != nil {
			return err
		}
		copyAudioSamples(r.dest, r.pending, r.scratch, off, take)
		r.pending += take
		off += take

		if r.pending == r.frameSize {
			r.dest.SetPts(r.pts)
			r.pts += int64(r.pending)
			if err := fn(r.dest); err != nil {
				return err
			}
			r.pending = 0
		}
	}
	return nil
}

// copyAudioSamples copies n samples of each plane from src (starting at
// srcOffset samples in) to dst (starting at dstOffset samples in). dst and
// src must share the same sample format and channel layout.
func copyAudioSamples(dst *frame.AudioFrame, dstOffset int, src *frame.AudioFrame, srcOffset, n int) {
	if n <= 0 {
		return
	}
	bytesPerSample := ff.AVUtil_get_bytes_per_sample(dst.SampleFormat())
	stride := bytesPerSample
	planes := 1
	if ff.AVUtil_sample_fmt_is_planar(dst.SampleFormat()) {
		planes = dst.NumChannels()
	} else {
		stride = bytesPerSample * dst.NumChannels()
	}
	for p := 0; p < planes; p++ {
		srcBytes, dstBytes := src.Bytes(p), dst.Bytes(p)
		so, do := srcOffset*stride, dstOffset*stride
		copy(dstBytes[do:do+n*stride], srcBytes[so:so+n*stride])
	}
}

func (r *audioResampler) init(src *frame.AudioFrame) error {
	ctx := ff.SWResample_alloc()
	if ctx == nil {
		return gomedia.ErrInternalError.With("resampler: failed to allocate swr context")
	}

	srcLayout, dstLayout := src.ChannelLayout(), r.par.ChannelLayout()
	if err := ff.SWResample_set_opts(ctx,
		dstLayout, r.par.SampleFormat(), r.par.SampleRate(),
		srcLayout, src.SampleFormat(), src.SampleRate(),
	); err != nil {
		ff.SWResample_free(ctx)
		return err
	}
	if err := ff.SWResample_init(ctx); err != nil {
		ff.SWResample_free(ctx)
		return err
	}

	// dest/scratch only need allocating once - their format is always the
	// fixed target (par), which never changes across a source switch; only
	// ctx (the source side) does. Reusing them keeps any still-pending
	// partially accumulated chunk intact across the switch.
	if r.dest == nil {
		capacity := r.frameSize
		if capacity <= 0 {
			capacity = defaultResampleCapacity
		}

		dest, err := newAudioFrameBuffer(src.Stream(), r.par.SampleFormat(), r.par.SampleRate(), dstLayout, capacity)
		if err != nil {
			ff.SWResample_free(ctx)
			return err
		}

		// Chunked mode re-chunks swr's output into exactly frameSize samples
		// per frame, which needs a separate scratch buffer for swr to
		// convert into - see accumulate.
		var scratch *frame.AudioFrame
		if r.frameSize > 0 {
			scratch, err = newAudioFrameBuffer(src.Stream(), r.par.SampleFormat(), r.par.SampleRate(), dstLayout, defaultResampleCapacity)
			if err != nil {
				dest.Close()
				ff.SWResample_free(ctx)
				return err
			}
		}

		r.dest, r.scratch, r.capacity = dest, scratch, capacity
	}

	r.ctx = ctx
	r.srcFmt, r.srcRate, r.srcLayout = src.SampleFormat(), src.SampleRate(), srcLayout
	return nil
}

// newAudioFrameBuffer allocates an audio frame with sample buffers for
// capacity samples in the given format/rate/layout.
func newAudioFrameBuffer(stream int, format ff.AVSampleFormat, rate int, layout ff.AVChannelLayout, capacity int) (*frame.AudioFrame, error) {
	f, err := frame.NewAudioFrame(stream)
	if err != nil {
		return nil, err
	}
	f.SetSampleFormat(format)
	f.SetSampleRate(rate)
	if err := f.SetChannelLayout(layout); err != nil {
		f.Close()
		return nil, err
	}
	f.SetNumSamples(capacity)
	if err := f.AllocateBuffers(); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

//////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - VIDEO

func (r *videoRescaler) matches(src *frame.VideoFrame) bool {
	return src.PixFmt() == r.par.PixelFormat() &&
		src.Width() == r.par.Width() && src.Height() == r.par.Height()
}

func (r *videoRescaler) process(src *frame.VideoFrame, fn func(*frame.VideoFrame) error) error {
	if src == nil {
		return nil
	}
	if r.matches(src) {
		// Own the pts even on the fast path - trusting src's own pts/timebase
		// here would break continuity the moment the source switches to one
		// numbering its frames differently (or not needing conversion, while
		// a previous source did).
		src.SetPts(r.pts)
		src.SetTimeBase(r.timebase)
		r.pts++
		return fn(src)
	}

	if r.ctx == nil || r.srcFmt != src.PixFmt() || r.srcW != src.Width() || r.srcH != src.Height() {
		if r.ctx != nil {
			ff.SWScale_free_context(r.ctx)
			r.ctx = nil
		}
		if r.dest == nil {
			dest, err := frame.NewVideoFrame(src.Stream())
			if err != nil {
				return err
			}
			dest.SetPixFmt(r.par.PixelFormat())
			dest.SetWidth(r.par.Width())
			dest.SetHeight(r.par.Height())
			if err := dest.AllocateBuffers(); err != nil {
				dest.Close()
				return err
			}
			r.dest = dest
		}
		ctx := ff.SWScale_get_context(
			src.Width(), src.Height(), src.PixFmt(),
			r.par.Width(), r.par.Height(), r.par.PixelFormat(),
			r.flags, nil, nil, nil,
		)
		if ctx == nil {
			return gomedia.ErrInternalError.With("resampler: failed to allocate swscale context")
		}
		r.ctx, r.srcFmt, r.srcW, r.srcH = ctx, src.PixFmt(), src.Width(), src.Height()
	}

	if err := ff.AVUtil_frame_copy_props(r.dest.AVFrame, src.AVFrame); err != nil {
		return err
	}

	// dest may still be referenced by the encoder from a previous fn(r.dest)
	// call - see audioResampler.processUnbounded for why this must happen
	// before writing into it. dest's width/height/format never change once
	// allocated, so there's no reordering concern like processUnbounded's.
	if err := r.dest.MakeWritable(); err != nil {
		return err
	}
	if err := ff.SWScale_scale_frame(r.ctx, r.dest.AVFrame, src.AVFrame, false); err != nil {
		return err
	}

	// Override whatever pts/timebase frame_copy_props just copied from src -
	// see the fast path above for why.
	r.dest.SetPts(r.pts)
	r.dest.SetTimeBase(r.timebase)
	r.pts++
	return fn(r.dest)
}
