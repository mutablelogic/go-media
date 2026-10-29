package segmenter_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	// Packages
	reader "github.com/mutablelogic/go-media/reader"
	segmenter "github.com/mutablelogic/go-media/segmenter"
	assert "github.com/stretchr/testify/assert"
	require "github.com/stretchr/testify/require"
)

func testDataPath(t *testing.T) string {
	wd, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(wd, "..", "etc", "test")
}

// decode returns copies of all the segments
func decode(t *testing.T, seg *segmenter.Segmenter) []segmenter.Segment {
	t.Helper()
	var result []segmenter.Segment
	require.NoError(t, seg.Decode(context.Background(), func(s *segmenter.Segment) error {
		result = append(result, segmenter.Segment{
			Start:    s.Start,
			Duration: s.Duration,
			Samples:  append([]float32(nil), s.Samples...),
			Class:    s.Class,
			Stats:    s.Stats,
		})
		return nil
	}))
	return result
}

// assertContiguous checks that each segment starts where the previous one ends,
// with a duration which matches its samples
func assertContiguous(t *testing.T, segments []segmenter.Segment, sampleRate int) int {
	t.Helper()
	var total int
	var next time.Duration
	for i, s := range segments {
		assert.Equalf(t, next, s.Start, "segment %d start", i)
		assert.Equalf(t, time.Duration(total+len(s.Samples))*time.Second/time.Duration(sampleRate)-s.Start, s.Duration, "segment %d duration", i)
		total += len(s.Samples)
		next = s.Start + s.Duration
	}
	return total
}

func TestSegmenter_Open(t *testing.T) {
	assert := assert.New(t)

	seg, err := segmenter.Open(filepath.Join(testDataPath(t), "sample.mp3"), 16000)
	require.NoError(t, err)
	defer seg.Close()

	assert.Equal(16000, seg.SampleRate())
	assert.Equal(0, seg.Stream())
	assert.Greater(seg.Duration(), time.Duration(0))
}

func TestSegmenter_NewReader(t *testing.T) {
	f, err := os.Open(filepath.Join(testDataPath(t), "sample.mp3"))
	require.NoError(t, err)
	defer f.Close()

	seg, err := segmenter.NewReader(f, 16000)
	require.NoError(t, err)
	defer seg.Close()

	segments := decode(t, seg)
	assert.Equal(t, 1, len(segments))
	assert.Greater(t, assertContiguous(t, segments, 16000), 0)
}

func TestSegmenter_NewReader_RawInput(t *testing.T) {
	f, err := os.Open(filepath.Join(testDataPath(t), "audio_22050_1ch_5m35.s16le.sw"))
	require.NoError(t, err)
	defer f.Close()

	seg, err := segmenter.NewReader(f, 16000,
		segmenter.WithReaderOpt(reader.WithInput("s16le", "sample_rate=22050", "channels=1")),
		segmenter.WithSegments(30*time.Second),
	)
	require.NoError(t, err)
	defer seg.Close()

	segments := decode(t, seg)
	total := assertContiguous(t, segments, 16000)

	// The file is 60s long, despite its name
	assert.Equal(t, 2, len(segments))
	assert.InDelta(t, 60*16000, total, 160)
}

func TestSegmenter_Decode_NoSegmentation(t *testing.T) {
	seg, err := segmenter.Open(filepath.Join(testDataPath(t), "jfk.wav"), 16000)
	require.NoError(t, err)
	defer seg.Close()

	segments := decode(t, seg)
	total := assertContiguous(t, segments, 16000)
	assert.Equal(t, 1, len(segments))
	assert.InDelta(t, seg.Duration().Seconds()*16000, total, 160)
}

func TestSegmenter_Decode_WithSegmentSize(t *testing.T) {
	seg, err := segmenter.Open(filepath.Join(testDataPath(t), "jfk.wav"), 16000, segmenter.WithSegments(time.Second))
	require.NoError(t, err)
	defer seg.Close()

	segments := decode(t, seg)
	total := assertContiguous(t, segments, 16000)
	require.Greater(t, len(segments), 1)
	for i, s := range segments[:len(segments)-1] {
		assert.Equalf(t, 16000, len(s.Samples), "segment %d", i)
	}
	assert.LessOrEqual(t, len(segments[len(segments)-1].Samples), 16000)
	assert.InDelta(t, seg.Duration().Seconds()*16000, total, 160)
}

func TestSegmenter_Decode_SampleRate(t *testing.T) {
	for _, rate := range []int{8000, 22050, 44100, 48000} {
		seg, err := segmenter.Open(filepath.Join(testDataPath(t), "jfk.wav"), rate)
		require.NoError(t, err)
		total := assertContiguous(t, decode(t, seg), rate)
		assert.InDeltaf(t, seg.Duration().Seconds()*float64(rate), total, float64(rate)/100, "rate %d", rate)
		seg.Close()
	}
}

func TestSegmenter_Decode_WithSilence(t *testing.T) {
	// The JFK WAV file has speech with pauses
	seg, err := segmenter.Open(filepath.Join(testDataPath(t), "jfk.wav"), 16000,
		segmenter.WithSegments(5*time.Second),
		segmenter.WithMode(segmenter.ModeSilence, 0, 0),
	)
	require.NoError(t, err)
	defer seg.Close()

	segments := decode(t, seg)
	assertContiguous(t, segments, 16000)
	require.Greater(t, len(segments), 1)

	// A segment which is cut on silence is shorter than the segment size
	var short int
	for _, s := range segments[:len(segments)-1] {
		t.Logf("Segment at %v: %v", s.Start, s.Duration)
		if s.Duration < 5*time.Second {
			short++
		}
	}
	assert.Greater(t, short, 0)

	// Silence is a segment of its own, and the segment after it starts with
	// sound. A segment before silence which is shorter than half the segment
	// size is output with the start of the silence, so the rest of the
	// silence can be shorter than the silence duration
	var silences int
	for i, s := range segments {
		if s.Class == segmenter.ClassSilence {
			silences++
		} else if i > 0 && segments[i-1].Class == segmenter.ClassSilence {
			var sum float32
			for _, v := range s.Samples[:160] { // 10ms
				sum += max(v, -v)
			}
			assert.Greaterf(t, sum/160, float32(1e-4), "segment %d starts too close to silence", i)
		}
	}
	assert.Greater(t, silences, 0)
}

// countingReader counts the bytes read from a reader
type countingReader struct {
	io.Reader
	n int
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.n += n
	return n, err
}

func TestSegmenter_Decode_Latency(t *testing.T) {
	// One second of sound, ten seconds of silence and one second of sound
	data := make([]byte, 12*16000*2)
	for i := range len(data) / 2 {
		if i < 16000 || i >= 11*16000 {
			binary.LittleEndian.PutUint16(data[i*2:], uint16(int16(tone(i)*math.MaxInt16)))
		}
	}
	for _, mode := range []segmenter.Mode{segmenter.ModeSilence, segmenter.ModeVADAggressive} {
		t.Run(mode.String(), func(t *testing.T) {
			r := &countingReader{Reader: bytes.NewReader(data)}
			seg, err := segmenter.NewReader(r, 16000,
				segmenter.WithReaderOpt(reader.WithInput("s16le", "sample_rate=16000", "channels=1")),
				segmenter.WithMode(mode, 0, 0),
			)
			require.NoError(t, err)
			defer seg.Close()

			// The sound is output once the silence has lasted 500ms, long
			// before the silence ends. Voice activity detection reports
			// speech for a little while after it ends
			var read []int
			var segments []segmenter.Segment
			require.NoError(t, seg.Decode(context.Background(), func(s *segmenter.Segment) error {
				read = append(read, r.n)
				segments = append(segments, segmenter.Segment{Start: s.Start, Duration: s.Duration, Class: s.Class})
				return nil
			}))
			require.Equal(t, 3, len(segments))
			t.Log(segments, read)
			assert.GreaterOrEqual(t, segments[0].Duration, time.Second)
			assert.Less(t, segments[0].Duration, 1200*time.Millisecond)
			assert.Less(t, read[0], 4*16000*2)
			assert.Equal(t, segmenter.ClassSilence, segments[1].Class)
			assert.Equal(t, 11*time.Second, segments[1].Start+segments[1].Duration)
			assert.NotEqual(t, segmenter.ClassSilence, segments[2].Class)
		})
	}
}

func TestSegmenter_Decode_Boundaries(t *testing.T) {
	path := filepath.Join(testDataPath(t), "jfk.wav")
	cuts := func(segments []segmenter.Segment) []time.Duration {
		var result []time.Duration
		for _, s := range segments[1:] {
			result = append(result, s.Start)
		}
		return result
	}

	// With silence detection, a cut is at a pause near the boundary. JFK
	// pauses at about 2.3-3.2s and 4.5-5.1s
	seg, err := segmenter.Open(path, 16000,
		segmenter.WithSegments(3*time.Second, 5*time.Second, 9*time.Second),
		segmenter.WithMode(segmenter.ModeSilence, 0, 0),
	)
	require.NoError(t, err)
	segments := decode(t, seg)
	seg.Close()
	assertContiguous(t, segments, 16000)
	t.Log(cuts(segments))
	require.Equal(t, 4, len(segments))
	for i, cut := range cuts(segments) {
		boundary := []time.Duration{3 * time.Second, 5 * time.Second, 9 * time.Second}[i]
		assert.LessOrEqualf(t, (cut - boundary).Abs(), time.Second, "cut %d", i)
	}
	assert.GreaterOrEqual(t, cuts(segments)[0], 2300*time.Millisecond)
	assert.LessOrEqual(t, cuts(segments)[0], 3200*time.Millisecond)
	assert.GreaterOrEqual(t, cuts(segments)[1], 4500*time.Millisecond)
	assert.LessOrEqual(t, cuts(segments)[1], 5100*time.Millisecond)
	for i, s := range segments {
		assert.NotEqualf(t, segmenter.ClassSilence, s.Class, "segment %d", i)
	}

	// Without silence detection, a cut is at the nearest second
	seg, err = segmenter.Open(path, 16000, segmenter.WithSegments(2500*time.Millisecond, 6*time.Second))
	require.NoError(t, err)
	segments = decode(t, seg)
	seg.Close()
	assertContiguous(t, segments, 16000)
	t.Log(cuts(segments))
	require.Equal(t, 3, len(segments))
	assert.LessOrEqual(t, (cuts(segments)[0] - 2500*time.Millisecond).Abs(), 500*time.Millisecond)
	assert.Equal(t, 6*time.Second, cuts(segments)[1])

	// A boundary after the end of the audio isn't cut at
	seg, err = segmenter.Open(path, 16000, segmenter.WithSegments(5*time.Second, 20*time.Second))
	require.NoError(t, err)
	segments = decode(t, seg)
	seg.Close()
	assertContiguous(t, segments, 16000)
	assert.Equal(t, 2, len(segments))
}

func TestSegmenter_Decode_EOF(t *testing.T) {
	seg, err := segmenter.Open(filepath.Join(testDataPath(t), "jfk.wav"), 16000, segmenter.WithSegments(time.Second))
	require.NoError(t, err)
	defer seg.Close()

	var count int
	assert.NoError(t, seg.Decode(context.Background(), func(*segmenter.Segment) error {
		count++
		return io.EOF
	}))
	assert.Equal(t, 1, count)
}

func TestSegmenter_Decode_Cancel(t *testing.T) {
	seg, err := segmenter.Open(filepath.Join(testDataPath(t), "jfk.wav"), 16000)
	require.NoError(t, err)
	defer seg.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.ErrorIs(t, seg.Decode(ctx, func(*segmenter.Segment) error { return nil }), context.Canceled)
}

// cancelReader cancels a context once n bytes have been read
type cancelReader struct {
	io.Reader
	n      int
	cancel context.CancelFunc
}

func (r *cancelReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if r.n -= n; r.n <= 0 {
		r.cancel()
	}
	return n, err
}

func TestSegmenter_Decode_CancelOutputsSamples(t *testing.T) {
	// Twenty seconds of silence, which is cancelled after six seconds have
	// been read, which includes what's read to probe the format
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := &cancelReader{Reader: bytes.NewReader(make([]byte, 20*16000*2)), n: 6 * 16000 * 2, cancel: cancel}
	seg, err := segmenter.NewReader(r, 16000,
		segmenter.WithReaderOpt(reader.WithInput("s16le", "sample_rate=16000", "channels=1")),
		segmenter.WithMode(segmenter.ModeSilence, 0, 0),
	)
	require.NoError(t, err)
	defer seg.Close()

	// The silence so far is output when it's cancelled
	var segments []segmenter.Segment
	err = seg.Decode(ctx, func(s *segmenter.Segment) error {
		segments = append(segments, segmenter.Segment{Start: s.Start, Duration: s.Duration, Class: s.Class})
		return nil
	})
	assert.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, len(segments))
	assert.Equal(t, segmenter.ClassSilence, segments[0].Class)
	assert.Greater(t, segments[0].Duration, time.Second)
	assert.Less(t, segments[0].Duration, 20*time.Second)
}

func TestSegmenter_Decode_Twice(t *testing.T) {
	// The reader can't rewind, so a second decode outputs nothing more
	seg, err := segmenter.Open(filepath.Join(testDataPath(t), "jfk.wav"), 16000)
	require.NoError(t, err)
	defer seg.Close()

	assert.Equal(t, 1, len(decode(t, seg)))
	assert.Equal(t, 0, len(decode(t, seg)))
}

func TestSegmenter_WithStream(t *testing.T) {
	path := filepath.Join(testDataPath(t), "sample.mp4")

	seg, err := segmenter.Open(path, 16000)
	require.NoError(t, err)
	stream := seg.Stream()
	seg.Close()

	seg, err = segmenter.Open(path, 16000, segmenter.WithStream(stream))
	require.NoError(t, err)
	assert.Equal(t, stream, seg.Stream())
	assert.Greater(t, assertContiguous(t, decode(t, seg), 16000), 0)
	seg.Close()

	// Not an audio stream
	_, err = segmenter.Open(path, 16000, segmenter.WithStream(stream+1))
	assert.Error(t, err)
	_, err = segmenter.Open(path, 16000, segmenter.WithStream(99))
	assert.Error(t, err)
}

func TestSegmenter_NoAudio(t *testing.T) {
	_, err := segmenter.Open(filepath.Join(testDataPath(t), "sample.png"), 16000)
	assert.Error(t, err)
}

func TestSegmenter_Invalid(t *testing.T) {
	path := filepath.Join(testDataPath(t), "sample.mp3")

	_, err := segmenter.Open(path, 0)
	assert.Error(t, err)
	_, err = segmenter.Open(path, -1)
	assert.Error(t, err)
	_, err = segmenter.Open(path, 16000, segmenter.WithSegments(10*time.Millisecond))
	assert.Error(t, err)
	_, err = segmenter.Open(path, 16000, segmenter.WithStream(-1))
	assert.Error(t, err)
	_, err = segmenter.Open(path, 16000, segmenter.WithSegments())
	assert.Error(t, err)
	_, err = segmenter.Open(path, 16000, segmenter.WithSegments(2*time.Second, time.Second))
	assert.Error(t, err)
	_, err = segmenter.Open(path, 16000, segmenter.WithSegments(0, time.Second))
	assert.Error(t, err)
	_, err = segmenter.Open("/nonexistent/file.mp3", 16000)
	assert.Error(t, err)

	seg, err := segmenter.Open(path, 16000)
	require.NoError(t, err)
	assert.Error(t, seg.Decode(context.Background(), nil))
	require.NoError(t, seg.Close())
	assert.Error(t, seg.Decode(context.Background(), func(*segmenter.Segment) error { return nil }))
}

func TestSegment_Int16(t *testing.T) {
	s := segmenter.Segment{Samples: []float32{0, 1, -1, 0.5, 2, -2}}
	assert.Equal(t, []int16{0, 32767, -32767, 16384, 32767, -32767}, s.Int16())
}

func TestSegmenter_Stats(t *testing.T) {
	seg, err := segmenter.Open(filepath.Join(testDataPath(t), "jfk.wav"), 16000, segmenter.WithSegments(time.Second))
	require.NoError(t, err)
	defer seg.Close()

	for i, s := range decode(t, seg) {
		var sum, peak float64
		for _, v := range s.Samples {
			sum += float64(v) * float64(v)
			peak = max(peak, float64(max(v, -v)))
		}
		assert.InDeltaf(t, sum/float64(len(s.Samples)), s.Stats.RMS*s.Stats.RMS, 1e-9, "segment %d RMS", i)
		assert.InDeltaf(t, peak, s.Stats.Peak, 1e-9, "segment %d peak", i)

		// No voice activity detection, so it's not classified beyond silence
		assert.Equal(t, 0.0, s.Stats.Speech)
		assert.Equalf(t, segmenter.ClassUnknown, s.Class, "segment %d", i)
	}
}

func TestSegmenter_VAD(t *testing.T) {
	seg, err := segmenter.Open(filepath.Join(testDataPath(t), "jfk.wav"), 16000,
		segmenter.WithSegments(time.Second),
		segmenter.WithMode(segmenter.ModeVADAggressive, 0, 0),
	)
	require.NoError(t, err)
	defer seg.Close()

	segments := decode(t, seg)
	assertContiguous(t, segments, 16000)
	var speech int
	for i, s := range segments {
		t.Logf("Segment at %v: %v speech=%.2f rms=%.3f peak=%.3f", s.Start, s.Class, s.Stats.Speech, s.Stats.RMS, s.Stats.Peak)
		assert.GreaterOrEqualf(t, s.Stats.Speech, 0.0, "segment %d", i)
		assert.LessOrEqualf(t, s.Stats.Speech, 1.0, "segment %d", i)
		if s.Class == segmenter.ClassSpeech {
			speech++
		}
	}

	// The recording is mostly speech
	assert.Greater(t, speech, len(segments)/2)
}

func TestSegmenter_Class(t *testing.T) {
	// Two seconds of samples at 16000 Hz
	tests := []struct {
		name   string
		mode   segmenter.Mode
		sample func(i int) float64
		class  segmenter.Class
	}{
		{"silence", segmenter.ModeVADAggressive, func(int) float64 { return 0 }, segmenter.ClassSilence},
		{"quiet", segmenter.ModeVADAggressive, func(i int) float64 { return 0.005 * math.Sin(float64(i)/10) }, segmenter.ClassSilence},
		{"tone", segmenter.ModeVADAggressive, tone, segmenter.ClassMusic},
		{"chord", segmenter.ModeVADAggressive, chord, segmenter.ClassMusic},
		{"noise", segmenter.ModeVADAggressive, func(i int) float64 { return 0.3 * (2*random(i) - 1) }, segmenter.ClassOther},
		{"silence without VAD", segmenter.ModeSilence, func(int) float64 { return 0 }, segmenter.ClassSilence},
		{"tone without VAD", segmenter.ModeSilence, tone, segmenter.ClassUnknown},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := make([]byte, 2*16000*2)
			for i := range len(data) / 2 {
				binary.LittleEndian.PutUint16(data[i*2:], uint16(int16(test.sample(i)*math.MaxInt16)))
			}
			seg, err := segmenter.NewReader(bytes.NewReader(data), 16000,
				segmenter.WithReaderOpt(reader.WithInput("s16le", "sample_rate=16000", "channels=1")),
				segmenter.WithMode(test.mode, 0, 0),
			)
			require.NoError(t, err)
			defer seg.Close()

			segments := decode(t, seg)
			require.Greater(t, len(segments), 0)
			for _, s := range segments {
				t.Logf("%v %v speech=%.2f rms=%.3f", s.Start, s.Class, s.Stats.Speech, s.Stats.RMS)
				assert.Equal(t, test.class, s.Class)
			}
		})
	}
}

func TestSegmenter_Class_Soundtrack(t *testing.T) {
	// A cartoon soundtrack of music and sound effects, without speech. The
	// sound effects have pauses and a spectrum which changes quickly, like
	// speech, so it's classified as speech, which this only logs
	seg, err := segmenter.Open(filepath.Join(testDataPath(t), "sample.mp3"), 16000,
		segmenter.WithMode(segmenter.ModeVADAggressive, 0, 0),
		segmenter.WithSegments(2*time.Second),
	)
	require.NoError(t, err)
	defer seg.Close()

	for _, s := range decode(t, seg) {
		t.Logf("Segment at %v: %v speech=%.2f low_energy=%.2f zcr_variation=%.2f flux=%.2f", s.Start, s.Class, s.Stats.Speech, s.Stats.LowEnergy, s.Stats.ZCRVariation, s.Stats.Flux)
	}
}

func TestSegmenter_Rhythm(t *testing.T) {
	// Four seconds at 16000 Hz, without silence detection
	for _, test := range []struct {
		name     string
		sample   func(i int) float64
		min, max float64
	}{
		{"clicks at 120 BPM", func(i int) float64 {
			if i%8000 < 160 {
				return 0.5 * (2*random(i) - 1)
			}
			return 0.001 * (2*random(i) - 1)
		}, 0.8, 1},
		{"noise", func(i int) float64 { return 0.3 * (2*random(i) - 1) }, 0, 0.3},
	} {
		t.Run(test.name, func(t *testing.T) {
			data := make([]byte, 4*16000*2)
			for i := range len(data) / 2 {
				binary.LittleEndian.PutUint16(data[i*2:], uint16(int16(test.sample(i)*math.MaxInt16)))
			}
			seg, err := segmenter.NewReader(bytes.NewReader(data), 16000,
				segmenter.WithReaderOpt(reader.WithInput("s16le", "sample_rate=16000", "channels=1")),
			)
			require.NoError(t, err)
			defer seg.Close()

			segments := decode(t, seg)
			require.Equal(t, 1, len(segments))
			t.Logf("rhythm=%.2f", segments[0].Stats.Rhythm)
			assert.GreaterOrEqual(t, segments[0].Stats.Rhythm, test.min)
			assert.LessOrEqual(t, segments[0].Stats.Rhythm, test.max)
		})
	}
}

// tone returns a 440Hz sine wave at 16000 Hz
func tone(i int) float64 {
	return 0.5 * math.Sin(2*math.Pi*440*float64(i)/16000)
}

// chord returns a C major chord with harmonics at 16000 Hz, which slowly
// changes in volume
func chord(i int) float64 {
	t := float64(i) / 16000
	var v float64
	for _, f := range []float64{261.6, 329.6, 392, 523.2} {
		for h := 1.0; h <= 4; h++ {
			v += math.Sin(2*math.Pi*f*h*t) / h
		}
	}
	return 0.1 * v * (0.6 + 0.4*math.Sin(2*math.Pi*0.5*t))
}

// random returns a pseudo-random number in [0, 1) for i
func random(i int) float64 {
	x := uint32(i)*1103515245 + 12345
	x ^= x >> 16
	x *= 0x45d9f3b
	x ^= x >> 16
	return float64(x) / (1 << 32)
}

func TestSegmenter_VAD_Mode(t *testing.T) {
	// A more aggressive mode reports less speech
	prev := -1.0
	for mode := segmenter.ModeVADQuality; mode <= segmenter.ModeVADVeryAggressive; mode++ {
		seg, err := segmenter.Open(filepath.Join(testDataPath(t), "jfk.wav"), 16000, segmenter.WithMode(mode, 0, 0))
		require.NoError(t, err)
		segments := decode(t, seg)
		seg.Close()

		var speech, total float64
		for _, s := range segments {
			speech += s.Stats.Speech * s.Duration.Seconds()
			total += s.Duration.Seconds()
		}
		t.Logf("%v: speech=%.2f", mode, speech/total)
		if prev >= 0 {
			assert.LessOrEqualf(t, speech, prev+0.05, "mode %v", mode)
		}
		prev = speech
	}
}

func TestSegmenter_VAD_Invalid(t *testing.T) {
	path := filepath.Join(testDataPath(t), "jfk.wav")

	seg, err := segmenter.Open(path, 16000, segmenter.WithMode(segmenter.ModeVADAggressive, 0, 0))
	if assert.NoError(t, err) {
		assert.Greater(t, len(decode(t, seg)), 0)
		seg.Close()
	}

	_, err = segmenter.Open(path, 16000, segmenter.WithMode(-1, 0, 0))
	assert.Error(t, err)
	_, err = segmenter.Open(path, 16000, segmenter.WithMode(segmenter.ModeVADVeryAggressive+1, 0, 0))
	assert.Error(t, err)
	_, err = segmenter.Open(path, 16000, segmenter.WithMode(segmenter.ModeVADAggressive, 10*time.Millisecond, 0))
	assert.Error(t, err)
	_, err = segmenter.Open(path, 16000, segmenter.WithMode(segmenter.ModeVADAggressive, 0, -0.1))
	assert.Error(t, err)
	_, err = segmenter.Open(path, 16000, segmenter.WithMode(segmenter.ModeSilence, 0, 1.1))
	assert.Error(t, err)

	// Voice activity detection works at any sample rate, and finds speech
	for _, rate := range []int{8000, 11025, 16000, 22050, 44100, 48000} {
		seg, err := segmenter.Open(path, rate, segmenter.WithMode(segmenter.ModeVADAggressive, 0, 0))
		if !assert.NoErrorf(t, err, "rate %d", rate) {
			continue
		}
		segments := decode(t, seg)
		seg.Close()
		assertContiguous(t, segments, rate)
		var speech time.Duration
		for _, s := range segments {
			speech += time.Duration(s.Stats.Speech * float64(s.Duration))
		}
		t.Logf("rate %d: %d segments, speech %v", rate, len(segments), speech)
		assert.Greaterf(t, speech, 5*time.Second, "rate %d", rate)
	}
}

func TestSegmenter_VAD_Boundaries(t *testing.T) {
	path := filepath.Join(testDataPath(t), "jfk.wav")

	// Segments break at the pauses between speech
	seg, err := segmenter.Open(path, 16000, segmenter.WithMode(segmenter.ModeVADAggressive, 0, 0))
	require.NoError(t, err)
	segments := decode(t, seg)
	seg.Close()
	assertContiguous(t, segments, 16000)
	require.Greater(t, len(segments), 1)
	for i, s := range segments {
		t.Logf("Segment at %v: %v %v speech=%.2f", s.Start, s.Duration, s.Class, s.Stats.Speech)
		if s.Class == segmenter.ClassSilence {
			assert.GreaterOrEqualf(t, s.Duration, segmenter.DefaultSilenceDuration, "segment %d", i)
			continue
		}
		assert.Equalf(t, segmenter.ClassSpeech, s.Class, "segment %d", i)
		if i > 0 {
			// Each segment after silence starts with speech
			var sum float64
			for _, v := range s.Samples[:320] {
				sum += float64(v) * float64(v)
			}
			assert.Greaterf(t, sum/320, 1e-4, "segment %d starts too close to silence", i)
		}
	}

	// No pause is as long as two seconds
	seg, err = segmenter.Open(path, 16000, segmenter.WithMode(segmenter.ModeVADAggressive, 2*time.Second, 0))
	require.NoError(t, err)
	assert.Equal(t, 1, len(decode(t, seg)))
	seg.Close()

	// A silence threshold counts quiet frames as silence too, so never
	// produces fewer segments
	seg, err = segmenter.Open(path, 16000, segmenter.WithMode(segmenter.ModeVADAggressive, 0, 0.05))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(decode(t, seg)), len(segments))
	seg.Close()
}
