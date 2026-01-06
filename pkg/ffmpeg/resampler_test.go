package ffmpeg

import (
	"testing"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg80"
	assert "github.com/stretchr/testify/assert"
)

// helper: make video frame with allocated buffers and black fill
func makeVideoFrame(t *testing.T, par *Par, pts int64) *Frame {
	t.Helper()
	f, err := NewFrame(par)
	if err != nil {
		t.Fatalf("NewFrame: %v", err)
	}
	if err := f.AllocateBuffers(); err != nil {
		t.Fatalf("AllocateBuffers: %v", err)
	}
	fillVideoBlackYUV420P(f)
	f.SetPts(pts)
	return f
}

func Test_resampler_audio_passthrough(t *testing.T) {
	assert := assert.New(t)

	dstPar, err := NewAudioPar("fltp", "stereo", 48000)
	if !assert.NoError(err) {
		t.FailNow()
	}

	r, err := NewResampler(dstPar, false)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	src, err := NewFrame(dstPar)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer src.Close()
	(*ff.AVFrame)(src).SetNumSamples(512)
	(*ff.AVFrame)(src).SetTimeBase(dstPar.timebase)
	if !assert.NoError(src.AllocateBuffers()) {
		t.FailNow()
	}
	fillAudioSilenceFLTP(src)
	src.SetPts(123)

	var out *Frame
	err = r.Resample(src, func(f *Frame) error {
		out = f
		return nil
	})
	assert.NoError(err)
	assert.EqualValues(123, out.Pts())
	assert.Same(src, out) // passthrough when formats match and force=false
}

func Test_resampler_audio_resample_rate(t *testing.T) {
	assert := assert.New(t)

	dstPar, err := NewAudioPar("fltp", "stereo", 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}
	srcPar, err := NewAudioPar("fltp", "stereo", 48000)
	if !assert.NoError(err) {
		t.FailNow()
	}

	r, err := NewResampler(dstPar, false)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	src, err := NewFrame(srcPar)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer src.Close()
	(*ff.AVFrame)(src).SetNumSamples(1024)
	(*ff.AVFrame)(src).SetTimeBase(srcPar.timebase)
	if !assert.NoError(src.AllocateBuffers()) {
		t.FailNow()
	}
	fillAudioSilenceFLTP(src)
	src.SetPts(321)

	var out *Frame
	err = r.Resample(src, func(f *Frame) error {
		out = f
		return nil
	})
	assert.NoError(err)
	if assert.NotNil(out) {
		assert.EqualValues(dstPar.SampleRate(), out.SampleRate())
		assert.NotZero(out.NumSamples())
		assert.EqualValues(ff.AVUtil_rational_rescale_q(src.Pts(), src.TimeBase(), out.TimeBase()), out.Pts())
	}
}

func Test_resampler_video_passthrough(t *testing.T) {
	assert := assert.New(t)

	dstPar, err := NewVideoPar("yuv420p", "320x240", 25.0)
	if !assert.NoError(err) {
		t.FailNow()
	}

	r, err := NewResampler(dstPar, false)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	src := makeVideoFrame(t, dstPar, 456)
	defer src.Close()

	var out *Frame
	err = r.Resample(src, func(f *Frame) error {
		out = f
		return nil
	})
	assert.NoError(err)
	// When passthrough, out is the same object as src, so don't close it separately
	assert.Same(src, out)
}

func Test_resampler_video_scale(t *testing.T) {
	assert := assert.New(t)

	dstPar, err := NewVideoPar("yuv420p", "160x120", 25.0)
	if !assert.NoError(err) {
		t.FailNow()
	}
	srcPar, err := NewVideoPar("yuv420p", "320x240", 25.0)
	if !assert.NoError(err) {
		t.FailNow()
	}

	r, err := NewResampler(dstPar, false)
	if !assert.NoError(err) {
		t.FailNow()
	}
	defer r.Close()

	src := makeVideoFrame(t, srcPar, 789)
	defer src.Close()

	var out *Frame
	err = r.Resample(src, func(f *Frame) error {
		out = f
		return nil
	})
	assert.NoError(err)
	if assert.NotNil(out) {
		assert.Equal(dstPar.Width(), out.Width())
		assert.Equal(dstPar.Height(), out.Height())
		assert.Equal(dstPar.PixelFormat(), out.PixelFormat())
	}
}
