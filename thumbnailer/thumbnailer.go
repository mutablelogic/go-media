package thumbnailer

import (
	"cmp"
	"context"
	"errors"
	"image"
	"io"
	"slices"
	"time"

	// Packages
	gomedia "github.com/mutablelogic/go-media"
	frame "github.com/mutablelogic/go-media/frame"
	reader "github.com/mutablelogic/go-media/reader"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	writer "github.com/mutablelogic/go-media/writer"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Thumbnailer finds the scenes and thumbnails of a video
type Thumbnailer struct {
	opts
	reader        *reader.Reader
	stream        int
	timebase      ff.AVRational
	width, height int           // size of the video as it's displayed, with square pixels
	displayed     bool          // whether width and height are from a frame, rather than as stored
	start         time.Duration // time of the stream's first frame, which times are from
	sar           ff.AVRational // sample aspect ratio the container or codec gives, or 0:1
	analysis      *Analysis
}

// Analysis is the frames sampled from a video, and its scenes
type Analysis struct {
	Mode     Mode          // How frames were sampled, ModeKeyframes or ModeSeek
	Duration time.Duration // Duration of the video, or zero if it's unknown
	Samples  []Sample      // Frames sampled, in time order
	Scenes   []Scene       // Scenes, in time order
}

// Thumbnail is a frame of a video, as an image
type Thumbnail struct {
	Time  time.Duration // Time of the frame
	Scene int           // Index of the scene it's in
	Stats Stats         // How good a thumbnail it makes
	Image image.Image   // The frame, at the thumbnail size
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// measureWidth is the width frames are measured at
	measureWidth = 320

	// probeDuration is how much of the video's packets are read to find how
	// often it has keyframes
	probeDuration = 30 * time.Second

	// Keyframes are sampled when there are between these numbers each
	// second, and at least minKeyframes in the video, and the video is
	// seeked otherwise
	minKeyframeRate = 0.1
	maxKeyframeRate = 4
	minKeyframes    = 8

	// minSeekInterval is the minimum time between frames sampled by seeking
	minSeekInterval = 500 * time.Millisecond
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Open opens a video from a file path or URL.
func Open(path string, opt ...Opt) (*Thumbnailer, error) {
	o, err := applyOpts(opt...)
	if err != nil {
		return nil, err
	}
	r, err := reader.Open(path, o.readerOpts...)
	if err != nil {
		return nil, err
	}
	t := &Thumbnailer{opts: *o, reader: r}

	// Choose the video stream
	var video []int
	for index, s := range r.Streams() {
		if s.Par().CodecType() == ff.AVMEDIA_TYPE_VIDEO {
			video = append(video, index)
		}
	}
	slices.Sort(video)
	switch {
	case t.opts.stream >= 0 && !slices.Contains(video, t.opts.stream):
		return nil, errors.Join(gomedia.ErrBadParameter.Withf("stream %d is not a video stream", t.opts.stream), r.Close())
	case t.opts.stream >= 0:
		t.stream = t.opts.stream
	case len(video) == 0:
		return nil, errors.Join(gomedia.ErrNotFound.With("no video stream"), r.Close())
	default:
		t.stream = video[0]
	}
	s := r.Streams()[t.stream]
	t.timebase = *s.TimeBase()
	t.start = r.StartTime(t.stream)
	t.sar = r.SampleAspectRatio(t.stream)
	t.width, t.height = s.Par().Width(), s.Par().Height()
	if t.width <= 0 || t.height <= 0 {
		return nil, errors.Join(gomedia.ErrBadParameter.With("video has no size"), r.Close())
	}
	return t, nil
}

// Close releases the video.
func (t *Thumbnailer) Close() error {
	var result error
	if t.reader != nil {
		result = t.reader.Close()
		t.reader = nil
	}
	return result
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Analyse samples frames from the video, measures them, and finds its
// scenes. It's only done once, and Thumbnails and Poster call it.
func (t *Thumbnailer) Analyse(ctx context.Context) (*Analysis, error) {
	if t.analysis != nil {
		return t.analysis, nil
	}
	if t.reader == nil {
		return nil, gomedia.ErrInternalError.With("thumbnailer is closed")
	}
	result := &Analysis{Mode: t.mode, Duration: t.reader.Duration()}

	// Start from the beginning, as Frame can have moved the reader
	if err := t.seek(0); err != nil {
		return nil, err
	}

	// Choose how to sample frames, from how often the video has keyframes
	if result.Mode == ModeAuto {
		mode, err := t.chooseMode(ctx, result.Duration)
		if err != nil {
			return nil, err
		}
		result.Mode = mode
	}

	// Sample the frames
	var err error
	switch result.Mode {
	case ModeKeyframes:
		result.Samples, err = t.sampleKeyframes(ctx)
	default:
		result.Samples, err = t.sampleSeek(ctx, result.Duration)
	}
	if err != nil {
		return nil, err
	}
	if len(result.Samples) == 0 {
		return nil, gomedia.ErrNotFound.With("no frames in the video")
	}

	result.Scenes = detect(result.Samples, result.Duration)
	t.analysis = result
	return result, nil
}

// Thumbnails returns up to n thumbnails of the video, in time order, each
// the best frame of a different scene, from the scenes which make the best
// thumbnails and don't look alike.
func (t *Thumbnailer) Thumbnails(ctx context.Context, n int) ([]Thumbnail, error) {
	a, err := t.Analyse(ctx)
	if err != nil {
		return nil, err
	}
	var result []Thumbnail
	for _, i := range choose(a.Samples, a.Scenes, n, a.Duration) {
		img, err := t.frameAt(ctx, a.Samples[i].Time)
		if err != nil {
			return nil, err
		}
		result = append(result, Thumbnail{Time: a.Samples[i].Time, Scene: sceneOf(a.Scenes, i), Stats: a.Samples[i].Stats, Image: img})
	}
	return result, nil
}

// Poster returns the one thumbnail which best represents the video.
func (t *Thumbnailer) Poster(ctx context.Context) (*Thumbnail, error) {
	thumbnails, err := t.Thumbnails(ctx, 1)
	if err != nil {
		return nil, err
	}
	if len(thumbnails) == 0 {
		return nil, gomedia.ErrNotFound.With("no thumbnail in the video")
	}
	return &thumbnails[0], nil
}

// Frame returns the frame at a time, or the first one after it, at the
// thumbnail size, such as the best frame of a scene.
func (t *Thumbnailer) Frame(ctx context.Context, at time.Duration) (image.Image, error) {
	return t.frameAt(ctx, at)
}

// Size returns the size of the video as it's displayed, with square pixels,
// which is the size it's stored at with its frames' sample aspect ratio
// applied, once a frame has been decoded, such as by Analyse, or the size
// it's stored at before then.
func (t *Thumbnailer) Size() (int, int) {
	return t.width, t.height
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// chooseMode returns ModeKeyframes when the start of the video has
// keyframes at a useful rate, and ModeSeek otherwise, such as when every
// frame is a keyframe, or they're too far apart
func (t *Thumbnailer) chooseMode(ctx context.Context, duration time.Duration) (Mode, error) {
	var keyframes []time.Duration
	var last time.Duration
	if err := t.reader.Decode(ctx, func(packet *ff.AVPacket) error {
		if packet.StreamIndex() != t.stream {
			return nil
		}
		ts, ok := t.time(packet.Pts())
		if !ok {
			ts, ok = t.time(packet.Dts())
		}
		if !ok {
			return nil
		}
		last = max(last, ts)
		if packet.Flags()&ff.AV_PKT_FLAG_KEY != 0 {
			keyframes = append(keyframes, ts)
		}
		if ts >= probeDuration {
			return io.EOF
		}
		return nil
	}, nil); err != nil {
		return ModeAuto, err
	}
	if err := t.seek(0); err != nil {
		return ModeAuto, err
	}

	// The rate of keyframes, and how many there are expected to be
	span := last.Seconds()
	if span <= 0 || len(keyframes) == 0 {
		return ModeSeek, nil
	}
	rate := float64(len(keyframes)) / span
	expected := float64(len(keyframes))
	if duration > 0 {
		expected = rate * duration.Seconds()
	}
	if rate >= minKeyframeRate && rate <= maxKeyframeRate && expected >= minKeyframes {
		return ModeKeyframes, nil
	}
	return ModeSeek, nil
}

// sampleKeyframes measures every keyframe in the video, which is decoded on
// its own, so it's fast
func (t *Thumbnailer) sampleKeyframes(ctx context.Context) ([]Sample, error) {
	var result []Sample
	err := t.decode(ctx, measureWidth, 0, "skip_frame=nokey", func(ts time.Duration, ok bool, img *image.YCbCr) error {
		if !ok {
			return nil
		}
		stats, hist := measure(img)
		result = append(result, Sample{Time: ts, Keyframe: true, Stats: stats, hist: hist})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := t.seek(0); err != nil {
		return nil, err
	}
	slices.SortFunc(result, func(a, b Sample) int { return cmp.Compare(a.Time, b.Time) })
	return result, nil
}

// sampleSeek measures frames at evenly spaced times, seeking to each, or
// every frame at least minSeekInterval apart if the duration is unknown
func (t *Thumbnailer) sampleSeek(ctx context.Context, duration time.Duration) ([]Sample, error) {
	var result []Sample
	add := func(ts time.Duration, img *image.YCbCr) {
		stats, hist := measure(img)
		result = append(result, Sample{Time: ts, Stats: stats, hist: hist})
	}

	// Without a duration, decode the whole video
	if duration <= 0 {
		next := time.Duration(0)
		err := t.decode(ctx, measureWidth, 0, "", func(ts time.Duration, ok bool, img *image.YCbCr) error {
			if ok && ts >= next {
				add(ts, img)
				next = ts + minSeekInterval
			}
			return nil
		})
		return result, err
	}

	// Seek to the middle of each interval, and measure the first frame at
	// or after it. A frame already sampled, as seeking goes back to the
	// keyframe before, is skipped, and a frame without a time is taken to be
	// at the time sought. A video shorter than the first interval's middle
	// is sampled from its start
	interval := max(minSeekInterval, duration/time.Duration(t.maxSamples))
	first := interval / 2
	if first >= duration {
		first = 0
	}
	for at := first; at < duration; at += interval {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := t.seek(at); err != nil {
			return nil, err
		}
		if err := t.decode(ctx, measureWidth, 0, "", func(ts time.Duration, ok bool, img *image.YCbCr) error {
			if !ok {
				ts = at
			}
			if ts < at || (len(result) > 0 && ts <= result[len(result)-1].Time) {
				return nil
			}
			add(ts, img)
			return io.EOF
		}); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// frameAt returns the frame at a time, at the thumbnail size, by seeking to
// it and decoding from the keyframe before it. A frame without a time is
// taken to be at the time sought
func (t *Thumbnailer) frameAt(ctx context.Context, at time.Duration) (image.Image, error) {
	if t.reader == nil {
		return nil, gomedia.ErrInternalError.With("thumbnailer is closed")
	}
	if err := t.seek(at); err != nil {
		return nil, err
	}
	var result image.Image
	if err := t.decode(ctx, t.opts.width, t.opts.height, "", func(ts time.Duration, ok bool, img *image.YCbCr) error {
		if ok && ts < at {
			return nil
		}
		result = copyYCbCr(img)
		return io.EOF
	}); err != nil {
		return nil, err
	}
	if result == nil {
		return nil, gomedia.ErrNotFound.Withf("no frame at %v", at)
	}
	return result, nil
}

// decode decodes the video stream from where the reader is, with codec
// options, and calls fn with each frame's time and the frame as a 4:2:0
// image of a size with square pixels (keeping the displayed aspect ratio
// when either is zero, and the displayed size when both are), until fn
// returns io.EOF or the video ends. The image is only valid during the
// call. The displayed size is from the first frame's sample aspect ratio,
// so the frames are scaled from then. Frames before the first keyframe are
// skipped, as after seeking a format without an index, such as MPEG-TS, the
// decoder can start between keyframes, and those frames are corrupt
func (t *Thumbnailer) decode(ctx context.Context, width, height int, codecOpts string, fn func(time.Duration, bool, *image.YCbCr) error) error {
	par := ff.AVCodec_parameters_alloc()
	if par == nil {
		return gomedia.ErrInternalError.With("failed to allocate codec parameters")
	}
	defer ff.AVCodec_parameters_free(par)
	var resampler *writer.Resampler
	defer func() {
		if resampler != nil {
			resampler.Close()
		}
	}()
	scaler := func(src *frame.VideoFrame) (*writer.Resampler, error) {
		if resampler != nil {
			return resampler, nil
		}
		if !t.displayed {
			// The container's or codec's sample aspect ratio is used over
			// the frame's, as FFmpeg does
			if t.sar.Num() > 0 && t.sar.Den() > 0 {
				src.SetSampleAspectRatio(t.sar)
			}
			t.width, t.height = src.DisplaySize()
			t.displayed = true
		}
		w, h := t.size(width, height)
		par.SetCodecType(ff.AVMEDIA_TYPE_VIDEO)
		par.SetPixelFormat(ff.AV_PIX_FMT_YUV420P)
		par.SetWidth(w)
		par.SetHeight(h)
		r, err := writer.NewResampler(par, 0, ff.AVUtil_rational(1, 25))
		resampler = r
		return r, err
	}

	var opts []string
	if codecOpts != "" {
		opts = append(opts, codecOpts)
	}
	keyframe := false
	decoder, err := reader.NewDecoder(func(f frame.Frame) error {
		src, ok := f.(*frame.VideoFrame)
		if !ok {
			return nil
		}
		if keyframe = keyframe || src.IsKeyFrame(); !keyframe {
			return nil
		}
		ts, ok := t.time(src.BestEffortTs())
		if !ok {
			ts, ok = t.time(src.Pts())
		}
		r, err := scaler(src)
		if err != nil {
			return err
		}
		return r.Process(src, func(f frame.Frame) error {
			img, err := f.(*frame.VideoFrame).Image()
			if err != nil {
				return err
			}
			return fn(ts, ok, img.(*image.YCbCr))
		})
	})
	if err != nil {
		return err
	}
	if err := decoder.Add(t.stream, opts...); err != nil {
		return err
	}
	return t.reader.Decode(ctx, nil, decoder)
}

// time returns a timestamp in the stream's timebase as a time from the
// stream's first frame, which can be negative, such as for frames an edit
// list skips, and false if it's unknown
func (t *Thumbnailer) time(ts int64) (time.Duration, bool) {
	if ts == int64(ff.AV_NOPTS_VALUE) {
		return 0, false
	}
	return time.Duration(ff.AVUtil_rational_rescale_q(ts, t.timebase, ff.AVUtil_rational(1, int(time.Second)))) - t.start, true
}

// seek seeks the reader to a time from the stream's first frame
func (t *Thumbnailer) seek(at time.Duration) error {
	return t.reader.Seek(t.stream, t.start+at)
}

// size returns a size, keeping the video's displayed aspect ratio when
// either is zero, rounded to even numbers for 4:2:0, and no larger than the
// video as it's displayed
func (t *Thumbnailer) size(width, height int) (int, int) {
	switch {
	case width == 0 && height == 0:
		width, height = t.width, t.height
	case height == 0:
		width = min(width, t.width)
		height = (width*t.height + t.width/2) / t.width
	case width == 0:
		height = min(height, t.height)
		width = (height*t.width + t.height/2) / t.height
	}
	return max(2, width&^1), max(2, height&^1)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE FUNCTIONS

// copyYCbCr returns a copy of a 4:2:0 image
func copyYCbCr(src *image.YCbCr) *image.YCbCr {
	dst := image.NewYCbCr(image.Rect(0, 0, src.Rect.Dx(), src.Rect.Dy()), src.SubsampleRatio)
	for y := range dst.Rect.Dy() {
		copy(dst.Y[y*dst.YStride:y*dst.YStride+dst.Rect.Dx()], src.Y[src.YOffset(src.Rect.Min.X, src.Rect.Min.Y+y):])
	}
	cw, ch := (dst.Rect.Dx()+1)/2, (dst.Rect.Dy()+1)/2
	for y := range ch {
		i := src.COffset(src.Rect.Min.X, src.Rect.Min.Y+2*y)
		copy(dst.Cb[y*dst.CStride:y*dst.CStride+cw], src.Cb[i:])
		copy(dst.Cr[y*dst.CStride:y*dst.CStride+cw], src.Cr[i:])
	}
	return dst
}

// sceneOf returns the index of the scene a sample is in
func sceneOf(scenes []Scene, sample int) int {
	for i := len(scenes) - 1; i >= 0; i-- {
		if scenes[i].First <= sample {
			return i
		}
	}
	return 0
}
