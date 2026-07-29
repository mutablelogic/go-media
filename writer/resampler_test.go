package writer

import (
	"testing"

	// Packages
	frame "github.com/mutablelogic/go-media/frame"
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
)

//////////////////////////////////////////////////////////////////////////////
// HELPERS

func audioCodecParameters(t *testing.T, sampleFmt string, sampleRate int, layout string) *ff.AVCodecParameters {
	t.Helper()
	par := ff.AVCodec_parameters_alloc()
	if par == nil {
		t.Fatal("AVCodec_parameters_alloc: nil")
	}
	t.Cleanup(func() { ff.AVCodec_parameters_free(par) })

	par.SetCodecType(ff.AVMEDIA_TYPE_AUDIO)
	par.SetSampleFormat(ff.AVUtil_get_sample_fmt(sampleFmt))
	par.SetSampleRate(sampleRate)

	var ch ff.AVChannelLayout
	if err := ff.AVUtil_channel_layout_from_string(&ch, layout); err != nil {
		t.Fatalf("AVUtil_channel_layout_from_string: %v", err)
	}
	if err := par.SetChannelLayout(ch); err != nil {
		t.Fatalf("SetChannelLayout: %v", err)
	}
	return par
}

func videoCodecParameters(t *testing.T, pixFmt string, width, height int) *ff.AVCodecParameters {
	t.Helper()
	par := ff.AVCodec_parameters_alloc()
	if par == nil {
		t.Fatal("AVCodec_parameters_alloc: nil")
	}
	t.Cleanup(func() { ff.AVCodec_parameters_free(par) })

	par.SetCodecType(ff.AVMEDIA_TYPE_VIDEO)
	par.SetPixelFormat(ff.AVUtil_get_pix_fmt(pixFmt))
	par.SetWidth(width)
	par.SetHeight(height)
	return par
}

// newS16MonoFrame builds a mono s16 audio frame carrying exactly samples.
func newS16MonoFrame(t *testing.T, stream, rate int, samples []int16) *frame.AudioFrame {
	t.Helper()
	f, err := frame.NewAudioFrame(stream)
	if err != nil {
		t.Fatalf("NewAudioFrame: %v", err)
	}
	f.SetSampleFormat(ff.AVUtil_get_sample_fmt("s16"))
	f.SetSampleRate(rate)

	var ch ff.AVChannelLayout
	if err := ff.AVUtil_channel_layout_from_string(&ch, "mono"); err != nil {
		t.Fatalf("AVUtil_channel_layout_from_string: %v", err)
	}
	if err := f.SetChannelLayout(ch); err != nil {
		t.Fatalf("SetChannelLayout: %v", err)
	}
	f.SetNumSamples(len(samples))
	if err := f.AllocateBuffers(); err != nil {
		t.Fatalf("AllocateBuffers: %v", err)
	}
	copy(f.Int16(0), samples)
	return f
}

func newYUV420PFrame(t *testing.T, stream, width, height int) *frame.VideoFrame {
	t.Helper()
	f, err := frame.NewVideoFrame(stream)
	if err != nil {
		t.Fatalf("NewVideoFrame: %v", err)
	}
	f.SetPixFmt(ff.AVUtil_get_pix_fmt("yuv420p"))
	f.SetWidth(width)
	f.SetHeight(height)
	if err := f.AllocateBuffers(); err != nil {
		t.Fatalf("AllocateBuffers: %v", err)
	}
	return f
}

//////////////////////////////////////////////////////////////////////////////
// AUDIO - frameSize accumulation across process() calls (bug #2)

// Feeds a sequence of small input frames (never a multiple of frameSize)
// carrying a globally increasing sample counter, so any dropped, duplicated
// or reordered sample is immediately detectable. Regression test for the
// re-chunking bug: process() used to emit whatever swr produced on every
// call, rather than buffering short calls until a full chunk was ready.
func TestAudioResampler_ChunkedAccumulatesAcrossCalls(t *testing.T) {
	const rate = 8000
	const frameSize = 100
	const inputChunk = 30
	const numInputs = 5 // 150 samples total - 1 full chunk + a 50-sample remainder

	par := audioCodecParameters(t, "s16", rate, "mono")
	r, err := NewResampler(par, frameSize, ff.AVRational{})
	if err != nil {
		t.Fatalf("NewResampler: %v", err)
	}
	defer r.audio.Close()

	var next int16
	var got []int16
	var gotSizes []int

	emit := func(f frame.Frame) error {
		af := f.(*frame.AudioFrame)
		got = append(got, af.Int16(0)...)
		gotSizes = append(gotSizes, af.NumSamples())
		return nil
	}

	for i := 0; i < numInputs; i++ {
		samples := make([]int16, inputChunk)
		for j := range samples {
			samples[j] = next
			next++
		}
		src := newS16MonoFrame(t, 0, rate, samples)
		err := r.Process(src, emit)
		src.Close()
		if err != nil {
			t.Fatalf("Process(%d): %v", i, err)
		}
	}
	if err := r.Process(nil, emit); err != nil {
		t.Fatalf("Process(flush): %v", err)
	}

	const total = numInputs * inputChunk
	if len(got) != total {
		t.Fatalf("got %d samples total, want %d", len(got), total)
	}
	for i, v := range got {
		if int(v) != i {
			t.Fatalf("sample %d = %d, want %d (dropped, duplicated, or reordered sample)", i, v, i)
		}
	}

	if len(gotSizes) == 0 {
		t.Fatal("expected at least one emitted chunk")
	}
	for i, n := range gotSizes {
		if i < len(gotSizes)-1 {
			if n != frameSize {
				t.Fatalf("chunk %d has %d samples, want exactly %d", i, n, frameSize)
			}
		} else if want := total % frameSize; n != want {
			t.Fatalf("final chunk has %d samples, want %d", n, want)
		}
	}
}

// Regression test for the pending==0 guard on the matches() fast path: a
// frame that already matches the target format/rate/layout/frameSize must
// still go through accumulation (not bypass straight to fn) if a previous
// short call left samples pending - otherwise it would be emitted out of
// order, ahead of the older buffered samples.
func TestAudioResampler_MatchingFrameDoesNotBypassPending(t *testing.T) {
	const rate = 8000
	const frameSize = 100

	par := audioCodecParameters(t, "s16", rate, "mono")
	r, err := NewResampler(par, frameSize, ff.AVRational{})
	if err != nil {
		t.Fatalf("NewResampler: %v", err)
	}
	defer r.audio.Close()

	var chunks [][]int16
	emit := func(f frame.Frame) error {
		af := f.(*frame.AudioFrame)
		chunk := make([]int16, af.NumSamples())
		copy(chunk, af.Int16(0))
		chunks = append(chunks, chunk)
		return nil
	}

	// Frame A: 30 samples (0..29) - leaves 30 samples pending, no emit yet.
	samplesA := make([]int16, 30)
	for i := range samplesA {
		samplesA[i] = int16(i)
	}
	srcA := newS16MonoFrame(t, 0, rate, samplesA)
	if err := r.Process(srcA, emit); err != nil {
		t.Fatalf("Process(A): %v", err)
	}
	srcA.Close()
	if len(chunks) != 0 {
		t.Fatalf("Process(A): got %d chunks, want 0 (nothing should be emitted yet)", len(chunks))
	}

	// Frame B: exactly frameSize samples (1000..1099), same format/rate -
	// matches() would fast-path this straight to fn if not for the
	// pending-samples guard.
	samplesB := make([]int16, frameSize)
	for i := range samplesB {
		samplesB[i] = int16(1000 + i)
	}
	srcB := newS16MonoFrame(t, 0, rate, samplesB)
	if err := r.Process(srcB, emit); err != nil {
		t.Fatalf("Process(B): %v", err)
	}
	srcB.Close()

	if err := r.Process(nil, emit); err != nil {
		t.Fatalf("Process(flush): %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("got %d chunks, want 2", len(chunks))
	}
	if len(chunks[0]) != frameSize {
		t.Fatalf("chunk 0 has %d samples, want %d", len(chunks[0]), frameSize)
	}
	for i := 0; i < 30; i++ {
		if chunks[0][i] != int16(i) {
			t.Fatalf("chunk 0[%d] = %d, want %d (A's pending samples must come first)", i, chunks[0][i], i)
		}
	}
	for i := 30; i < frameSize; i++ {
		want := int16(1000 + (i - 30))
		if chunks[0][i] != want {
			t.Fatalf("chunk 0[%d] = %d, want %d (B must fill the rest of the chunk in order)", i, chunks[0][i], want)
		}
	}
	if len(chunks[1]) != 30 {
		t.Fatalf("chunk 1 (flush) has %d samples, want 30", len(chunks[1]))
	}
	for i, v := range chunks[1] {
		want := int16(1000 + 70 + i)
		if v != want {
			t.Fatalf("chunk 1[%d] = %d, want %d", i, v, want)
		}
	}
}

// Sanity check for the frameSize<=0 ("codec accepts any frame size") path:
// output is passed through with no re-chunking, and the total sample count
// is preserved across a flush.
func TestAudioResampler_UnboundedPreservesSampleCount(t *testing.T) {
	const rate = 8000
	par := audioCodecParameters(t, "s16", rate, "mono")
	r, err := NewResampler(par, 0, ff.AVRational{})
	if err != nil {
		t.Fatalf("NewResampler: %v", err)
	}
	defer r.audio.Close()

	var total int
	emit := func(f frame.Frame) error {
		total += f.(*frame.AudioFrame).NumSamples()
		return nil
	}

	samples := make([]int16, 500)
	for i := range samples {
		samples[i] = int16(i)
	}
	src := newS16MonoFrame(t, 0, rate, samples)
	defer src.Close()

	if err := r.Process(src, emit); err != nil {
		t.Fatalf("Process: %v", err)
	}
	if err := r.Process(nil, emit); err != nil {
		t.Fatalf("Process(flush): %v", err)
	}
	if total != len(samples) {
		t.Fatalf("got %d samples total, want %d", total, len(samples))
	}
}

// Regression test for owning pts across a source switch: the resampler must
// never trust a source frame's own pts, since a second source can restart
// its own numbering (or use a different rate/format) independently of the
// first - only a resampler-owned, cumulative-sample-count pts stays
// continuous across the switch. Exercises both the fast (format-matches)
// path and the full conversion path.
func TestAudioResampler_OwnsPtsAcrossSourceSwitch(t *testing.T) {
	const rate = 8000
	par := audioCodecParameters(t, "s16", rate, "mono")
	r, err := NewResampler(par, 0, ff.AVRational{})
	if err != nil {
		t.Fatalf("NewResampler: %v", err)
	}
	defer r.audio.Close()

	var gotPts []int64
	var gotN []int
	emit := func(f frame.Frame) error {
		af := f.(*frame.AudioFrame)
		gotPts = append(gotPts, af.Pts())
		gotN = append(gotN, af.NumSamples())
		return nil
	}

	// Source A: already matches the target format exactly (fast path) -
	// carries a large, unrelated pts.
	srcA := newS16MonoFrame(t, 0, rate, make([]int16, 50))
	srcA.SetPts(123456)
	if err := r.Process(srcA, emit); err != nil {
		t.Fatalf("Process(A): %v", err)
	}
	srcA.Close()

	// Source B: switches to double the sample rate (forces real conversion)
	// and restarts its own pts numbering at an unrelated value.
	srcB := newS16MonoFrame(t, 0, rate*2, make([]int16, 100))
	srcB.SetPts(7)
	if err := r.Process(srcB, emit); err != nil {
		t.Fatalf("Process(B): %v", err)
	}
	srcB.Close()

	if err := r.Process(nil, emit); err != nil {
		t.Fatalf("Process(flush): %v", err)
	}

	if len(gotPts) == 0 {
		t.Fatal("expected at least one emitted frame")
	}
	want := int64(0)
	for i, pts := range gotPts {
		if pts != want {
			t.Fatalf("frame %d: pts = %d, want %d (must be cumulative sample count, ignoring src's own pts)", i, pts, want)
		}
		want += int64(gotN[i])
	}
}

// Regression test for sourceChanged/drain/teardown: switching to a source
// with a different sample rate or channel layout mid-stream must drain the
// outgoing context, rebuild ctx for the new source, and keep chunking
// exactly frameSize samples per frame throughout, with no crash - this
// exercises the exact cgo channel-layout-comparison path that originally
// panicked during development ("Go pointer to unpinned Go pointer") when
// sourceChanged compared a live frame's layout against one stored as a raw
// AVChannelLayout struct field - and confirms the pending accumulator
// survives the switch intact rather than being reset or corrupted.
func TestAudioResampler_SourceChangeDrainsAndReinitializes(t *testing.T) {
	const frameSize = 100
	par := audioCodecParameters(t, "s16", 8000, "mono")
	r, err := NewResampler(par, frameSize, ff.AVRational{})
	if err != nil {
		t.Fatalf("NewResampler: %v", err)
	}
	defer r.audio.Close()

	var chunkSizes []int
	emit := func(f frame.Frame) error {
		chunkSizes = append(chunkSizes, f.(*frame.AudioFrame).NumSamples())
		return nil
	}

	// Source A: matches the target format but not frameSize, so it's forced
	// through real conversion (initializing ctx for the first time) and
	// leaves a partial chunk pending.
	srcA := newS16MonoFrame(t, 0, 8000, make([]int16, 30))
	if err := r.Process(srcA, emit); err != nil {
		t.Fatalf("Process(A): %v", err)
	}
	srcA.Close()
	if r.audio.pending != 30 {
		t.Fatalf("pending after A = %d, want 30", r.audio.pending)
	}

	// Source B: a different sample rate - forces sourceChanged, which
	// drains and rebuilds ctx. pending must be untouched by the switch.
	srcB := newS16MonoFrame(t, 0, 16000, make([]int16, 200))
	if err := r.Process(srcB, emit); err != nil {
		t.Fatalf("Process(B): %v", err)
	}
	srcB.Close()

	// Source C: back to A's rate, but a different channel layout this time -
	// exercises the layout side of sourceChanged specifically.
	srcC, err := frame.NewAudioFrame(0)
	if err != nil {
		t.Fatalf("NewAudioFrame: %v", err)
	}
	srcC.SetSampleFormat(ff.AVUtil_get_sample_fmt("s16"))
	srcC.SetSampleRate(8000)
	var stereo ff.AVChannelLayout
	if err := ff.AVUtil_channel_layout_from_string(&stereo, "stereo"); err != nil {
		t.Fatalf("AVUtil_channel_layout_from_string: %v", err)
	}
	if err := srcC.SetChannelLayout(stereo); err != nil {
		t.Fatalf("SetChannelLayout: %v", err)
	}
	srcC.SetNumSamples(60)
	if err := srcC.AllocateBuffers(); err != nil {
		t.Fatalf("AllocateBuffers: %v", err)
	}
	if err := r.Process(srcC, emit); err != nil {
		t.Fatalf("Process(C): %v", err)
	}
	srcC.Close()

	if err := r.Process(nil, emit); err != nil {
		t.Fatalf("Process(flush): %v", err)
	}

	if len(chunkSizes) == 0 {
		t.Fatal("expected at least one emitted chunk")
	}
	for i, n := range chunkSizes {
		if i < len(chunkSizes)-1 {
			if n != frameSize {
				t.Fatalf("chunk %d has %d samples, want exactly %d", i, n, frameSize)
			}
		} else if n <= 0 || n > frameSize {
			t.Fatalf("final chunk has %d samples, want 1..%d", n, frameSize)
		}
	}
	if r.audio.pending != 0 {
		t.Fatalf("pending after flush = %d, want 0", r.audio.pending)
	}
}

func TestAudioResampler_Close_Idempotent(t *testing.T) {
	par := audioCodecParameters(t, "s16", 8000, "mono")
	r, err := NewResampler(par, 100, ff.AVRational{})
	if err != nil {
		t.Fatalf("NewResampler: %v", err)
	}

	src := newS16MonoFrame(t, 0, 8000, make([]int16, 10))
	if err := r.Process(src, func(frame.Frame) error { return nil }); err != nil {
		t.Fatalf("Process: %v", err)
	}
	src.Close()

	if err := r.audio.Close(); err != nil {
		t.Fatalf("Close (first): %v", err)
	}
	if err := r.audio.Close(); err != nil {
		t.Fatalf("Close (second): %v", err)
	}
}

//////////////////////////////////////////////////////////////////////////////
// VIDEO - context replaced (not dangling) on format/size change (bug #1)

// Regression test for the videoRescaler.process free/realloc path: the old
// swscale context must never be left set once freed. Repeatedly swapping
// between sizes exercises that path on every call and confirms the context
// is always replaced with a fresh one rather than reused after being freed.
// (Forcing SWScale_get_context itself to fail isn't reproducible through
// valid public inputs, since NewResampler/newVideoRescaler already reject
// any parameters that would make allocation fail.)
func TestVideoRescaler_ContextReplacedOnSizeChange(t *testing.T) {
	par := videoCodecParameters(t, "yuv420p", 160, 120)
	r, err := NewResampler(par, 0, ff.AVUtil_rational(1, 25))
	if err != nil {
		t.Fatalf("NewResampler: %v", err)
	}
	defer r.video.Close()

	// Note: contexts aren't compared by pointer across iterations - the
	// allocator can legitimately hand back the same address once freed, so
	// pointer equality wouldn't distinguish "replaced" from "dangling".
	// What matters is that every swap leaves ctx non-nil and valid to use.
	sizes := [][2]int{{320, 240}, {640, 480}, {160, 120}, {800, 600}, {320, 240}}
	for _, dims := range sizes {
		src := newYUV420PFrame(t, 0, dims[0], dims[1])
		if err := r.Process(src, func(frame.Frame) error { return nil }); err != nil {
			src.Close()
			t.Fatalf("Process(%dx%d): %v", dims[0], dims[1], err)
		}
		src.Close()

		if r.video.ctx == nil {
			t.Fatalf("ctx is nil after successful rescale to %dx%d", dims[0], dims[1])
		}
	}
}

// Regression test for owning pts across a source switch: neither the fast
// (format-matches) path nor the full rescale path may leak a source frame's
// own pts/timebase through, since a second source can restart its own
// numbering (or use a different timebase) independently of the first - only
// a resampler-owned, one-tick-per-frame pts at the fixed target timebase
// stays continuous and correctly-scaled across the switch.
func TestVideoRescaler_OwnsPtsAcrossSourceSwitch(t *testing.T) {
	const targetW, targetH = 160, 120
	target := ff.AVUtil_rational(1, 25)

	par := videoCodecParameters(t, "yuv420p", targetW, targetH)
	r, err := NewResampler(par, 0, target)
	if err != nil {
		t.Fatalf("NewResampler: %v", err)
	}
	defer r.video.Close()

	var gotPts []int64
	var gotTb []ff.AVRational
	emit := func(f frame.Frame) error {
		vf := f.(*frame.VideoFrame)
		gotPts = append(gotPts, vf.Pts())
		gotTb = append(gotTb, vf.TimeBase())
		return nil
	}

	// Source A: a different size (forces rescaling), carrying an unrelated
	// pts/timebase.
	srcA := newYUV420PFrame(t, 0, 320, 240)
	srcA.SetPts(999999)
	srcA.SetTimeBase(ff.AVUtil_rational(1, 90000))
	if err := r.Process(srcA, emit); err != nil {
		t.Fatalf("Process(A): %v", err)
	}
	srcA.Close()

	// Source B: yet another size, restarting its own pts numbering with a
	// different timebase again - simulates switching input sources.
	srcB := newYUV420PFrame(t, 0, 640, 480)
	srcB.SetPts(0)
	srcB.SetTimeBase(ff.AVUtil_rational(1, 1000))
	if err := r.Process(srcB, emit); err != nil {
		t.Fatalf("Process(B): %v", err)
	}
	srcB.Close()

	// Source C: already matches the target format/size exactly (fast path).
	srcC := newYUV420PFrame(t, 0, targetW, targetH)
	srcC.SetPts(555)
	srcC.SetTimeBase(ff.AVUtil_rational(1, 48000))
	if err := r.Process(srcC, emit); err != nil {
		t.Fatalf("Process(C): %v", err)
	}
	srcC.Close()

	if len(gotPts) != 3 {
		t.Fatalf("got %d emitted frames, want 3", len(gotPts))
	}
	for i, pts := range gotPts {
		if pts != int64(i) {
			t.Fatalf("frame %d: pts = %d, want %d (must ignore src's own pts)", i, pts, i)
		}
		if gotTb[i] != target {
			t.Fatalf("frame %d: timebase = %v, want %v (must ignore src's own timebase)", i, gotTb[i], target)
		}
	}
}

func TestVideoRescaler_Close_Idempotent(t *testing.T) {
	par := videoCodecParameters(t, "yuv420p", 160, 120)
	r, err := NewResampler(par, 0, ff.AVUtil_rational(1, 25))
	if err != nil {
		t.Fatalf("NewResampler: %v", err)
	}

	src := newYUV420PFrame(t, 0, 320, 240)
	if err := r.Process(src, func(frame.Frame) error { return nil }); err != nil {
		t.Fatalf("Process: %v", err)
	}
	src.Close()

	if err := r.video.Close(); err != nil {
		t.Fatalf("Close (first): %v", err)
	}
	if err := r.video.Close(); err != nil {
		t.Fatalf("Close (second): %v", err)
	}
}
