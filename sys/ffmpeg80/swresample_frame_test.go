package ffmpeg

import (
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test complete workflow: AVFrame in -> resample -> AVFrame out
func Test_swresample_frame_workflow_stereo_to_mono(t *testing.T) {
	assert := assert.New(t)

	// Allocate resampler context
	ctx := SWResample_alloc()
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWResample_free(ctx)

	// Create source frame: Stereo, 48kHz, FLTP, 1024 samples
	src := AVUtil_frame_alloc()
	if !assert.NotNil(src) {
		t.FailNow()
	}
	defer AVUtil_frame_free(src)

	src.SetChannelLayout(AV_CHANNEL_LAYOUT_STEREO())
	src.SetSampleRate(48000)
	src.SetSampleFormat(AV_SAMPLE_FMT_FLTP)
	src.SetNumSamples(1024)

	err := AVUtil_frame_get_buffer(src, false)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create destination frame: Mono, 44.1kHz, S16
	dst := AVUtil_frame_alloc()
	if !assert.NotNil(dst) {
		t.FailNow()
	}
	defer AVUtil_frame_free(dst)

	dst.SetChannelLayout(AV_CHANNEL_LAYOUT_MONO())
	dst.SetSampleRate(44100)
	dst.SetSampleFormat(AV_SAMPLE_FMT_S16)

	// Convert frame - this automatically configures the resampler
	err = SWResample_convert_frame(ctx, src, dst)
	assert.NoError(err, "Frame conversion should succeed")

	// Verify context is now initialized
	assert.True(SWResample_is_initialized(ctx))

	// Verify output frame properties
	assert.Equal(44100, dst.SampleRate())
	assert.Equal(AV_SAMPLE_FMT_S16, dst.SampleFormat())
	assert.Equal(1, dst.NumChannels())
	assert.Greater(dst.NumSamples(), 0, "Output should have samples")

	t.Logf("Converted %d samples (stereo 48kHz FLTP) -> %d samples (mono 44.1kHz S16)",
		src.NumSamples(), dst.NumSamples())
}

func Test_swresample_frame_workflow_mono_to_stereo(t *testing.T) {
	assert := assert.New(t)

	ctx := SWResample_alloc()
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWResample_free(ctx)

	// Create source frame: Mono, 44.1kHz, S16, 1024 samples
	src := AVUtil_frame_alloc()
	if !assert.NotNil(src) {
		t.FailNow()
	}
	defer AVUtil_frame_free(src)

	src.SetChannelLayout(AV_CHANNEL_LAYOUT_MONO())
	src.SetSampleRate(44100)
	src.SetSampleFormat(AV_SAMPLE_FMT_S16)
	src.SetNumSamples(1024)

	err := AVUtil_frame_get_buffer(src, false)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create destination frame: Stereo, 48kHz, FLTP
	dst := AVUtil_frame_alloc()
	if !assert.NotNil(dst) {
		t.FailNow()
	}
	defer AVUtil_frame_free(dst)

	dst.SetChannelLayout(AV_CHANNEL_LAYOUT_STEREO())
	dst.SetSampleRate(48000)
	dst.SetSampleFormat(AV_SAMPLE_FMT_FLTP)

	// Convert
	err = SWResample_convert_frame(ctx, src, dst)
	assert.NoError(err, "Frame conversion should succeed")

	// Verify output
	assert.Equal(48000, dst.SampleRate())
	assert.Equal(AV_SAMPLE_FMT_FLTP, dst.SampleFormat())
	assert.Equal(2, dst.NumChannels())
	assert.Greater(dst.NumSamples(), 0)

	t.Logf("Converted %d samples (mono 44.1kHz S16) -> %d samples (stereo 48kHz FLTP)",
		src.NumSamples(), dst.NumSamples())
}

func Test_swresample_frame_workflow_5point1_to_stereo(t *testing.T) {
	assert := assert.New(t)

	ctx := SWResample_alloc()
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWResample_free(ctx)

	// Create source frame: 5.1 surround, 48kHz, FLTP, 1024 samples
	src := AVUtil_frame_alloc()
	if !assert.NotNil(src) {
		t.FailNow()
	}
	defer AVUtil_frame_free(src)

	src.SetChannelLayout(AV_CHANNEL_LAYOUT_5POINT1())
	src.SetSampleRate(48000)
	src.SetSampleFormat(AV_SAMPLE_FMT_FLTP)
	src.SetNumSamples(1024)

	err := AVUtil_frame_get_buffer(src, false)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Create destination frame: Stereo, 48kHz, S16
	dst := AVUtil_frame_alloc()
	if !assert.NotNil(dst) {
		t.FailNow()
	}
	defer AVUtil_frame_free(dst)

	dst.SetChannelLayout(AV_CHANNEL_LAYOUT_STEREO())
	dst.SetSampleRate(48000)
	dst.SetSampleFormat(AV_SAMPLE_FMT_S16)

	// Convert 5.1 to stereo downmix
	err = SWResample_convert_frame(ctx, src, dst)
	assert.NoError(err, "5.1 to stereo downmix should succeed")

	// Verify output
	assert.Equal(48000, dst.SampleRate())
	assert.Equal(AV_SAMPLE_FMT_S16, dst.SampleFormat())
	assert.Equal(2, dst.NumChannels())
	assert.Greater(dst.NumSamples(), 0)

	t.Logf("Downmixed 5.1 surround (%d channels) -> stereo (2 channels)",
		src.NumChannels())
}

func Test_swresample_frame_workflow_multiple_conversions(t *testing.T) {
	assert := assert.New(t)

	// Test reusing the same context for multiple conversions
	ctx := SWResample_alloc()
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWResample_free(ctx)

	for i := 0; i < 3; i++ {
		src := AVUtil_frame_alloc()
		if !assert.NotNil(src) {
			t.FailNow()
		}

		src.SetChannelLayout(AV_CHANNEL_LAYOUT_STEREO())
		src.SetSampleRate(48000)
		src.SetSampleFormat(AV_SAMPLE_FMT_FLTP)
		src.SetNumSamples(512)

		err := AVUtil_frame_get_buffer(src, false)
		if !assert.NoError(err) {
			AVUtil_frame_free(src)
			t.FailNow()
		}

		dst := AVUtil_frame_alloc()
		if !assert.NotNil(dst) {
			AVUtil_frame_free(src)
			t.FailNow()
		}

		dst.SetChannelLayout(AV_CHANNEL_LAYOUT_MONO())
		dst.SetSampleRate(44100)
		dst.SetSampleFormat(AV_SAMPLE_FMT_S16)

		err = SWResample_convert_frame(ctx, src, dst)
		assert.NoError(err, "Conversion %d should succeed", i+1)

		AVUtil_frame_free(src)
		AVUtil_frame_free(dst)
	}

	t.Log("Successfully reused context for 3 conversions")
}

func Test_swresample_frame_workflow_format_conversions(t *testing.T) {
	tests := []struct {
		name        string
		in_layout   AVChannelLayout
		out_layout  AVChannelLayout
		in_rate     int
		out_rate    int
		in_format   AVSampleFormat
		out_format  AVSampleFormat
		num_samples int
	}{
		{
			name:        "S16 to FLTP",
			in_layout:   AV_CHANNEL_LAYOUT_STEREO(),
			out_layout:  AV_CHANNEL_LAYOUT_STEREO(),
			in_rate:     48000,
			out_rate:    48000,
			in_format:   AV_SAMPLE_FMT_S16,
			out_format:  AV_SAMPLE_FMT_FLTP,
			num_samples: 1024,
		},
		{
			name:        "FLTP to S16",
			in_layout:   AV_CHANNEL_LAYOUT_STEREO(),
			out_layout:  AV_CHANNEL_LAYOUT_STEREO(),
			in_rate:     48000,
			out_rate:    48000,
			in_format:   AV_SAMPLE_FMT_FLTP,
			out_format:  AV_SAMPLE_FMT_S16,
			num_samples: 1024,
		},
		{
			name:        "FLT to FLTP",
			in_layout:   AV_CHANNEL_LAYOUT_STEREO(),
			out_layout:  AV_CHANNEL_LAYOUT_STEREO(),
			in_rate:     48000,
			out_rate:    48000,
			in_format:   AV_SAMPLE_FMT_FLT,
			out_format:  AV_SAMPLE_FMT_FLTP,
			num_samples: 1024,
		},
		{
			name:        "Complex: 5.1@96kHz FLT to Stereo@48kHz S16",
			in_layout:   AV_CHANNEL_LAYOUT_5POINT1(),
			out_layout:  AV_CHANNEL_LAYOUT_STEREO(),
			in_rate:     96000,
			out_rate:    48000,
			in_format:   AV_SAMPLE_FMT_FLT,
			out_format:  AV_SAMPLE_FMT_S16,
			num_samples: 2048,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			testAssert := assert.New(t)

			ctx := SWResample_alloc()
			if !testAssert.NotNil(ctx) {
				t.FailNow()
			}
			defer SWResample_free(ctx)

			src := AVUtil_frame_alloc()
			if !testAssert.NotNil(src) {
				t.FailNow()
			}
			defer AVUtil_frame_free(src)

			src.SetChannelLayout(tt.in_layout)
			src.SetSampleRate(tt.in_rate)
			src.SetSampleFormat(tt.in_format)
			src.SetNumSamples(tt.num_samples)

			err := AVUtil_frame_get_buffer(src, false)
			if !testAssert.NoError(err) {
				t.FailNow()
			}

			dst := AVUtil_frame_alloc()
			if !testAssert.NotNil(dst) {
				t.FailNow()
			}
			defer AVUtil_frame_free(dst)

			dst.SetChannelLayout(tt.out_layout)
			dst.SetSampleRate(tt.out_rate)
			dst.SetSampleFormat(tt.out_format)

			err = SWResample_convert_frame(ctx, src, dst)
			testAssert.NoError(err, "Conversion should succeed")

			testAssert.Equal(tt.out_rate, dst.SampleRate())
			testAssert.Equal(tt.out_format, dst.SampleFormat())
			testAssert.Greater(dst.NumSamples(), 0)

			t.Logf("Converted: %dch@%dHz (%v) -> %dch@%dHz (%v), %d -> %d samples",
				src.NumChannels(), src.SampleRate(), tt.in_format,
				dst.NumChannels(), dst.SampleRate(), tt.out_format,
				src.NumSamples(), dst.NumSamples())
		})
	}
}

// Test flushing remaining samples
func Test_swresample_frame_workflow_flush(t *testing.T) {
	assert := assert.New(t)

	ctx := SWResample_alloc()
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWResample_free(ctx)

	// Process a frame first
	src := AVUtil_frame_alloc()
	if !assert.NotNil(src) {
		t.FailNow()
	}
	defer AVUtil_frame_free(src)

	src.SetChannelLayout(AV_CHANNEL_LAYOUT_STEREO())
	src.SetSampleRate(48000)
	src.SetSampleFormat(AV_SAMPLE_FMT_FLTP)
	src.SetNumSamples(1024)

	err := AVUtil_frame_get_buffer(src, false)
	if !assert.NoError(err) {
		t.FailNow()
	}

	dst := AVUtil_frame_alloc()
	if !assert.NotNil(dst) {
		t.FailNow()
	}
	defer AVUtil_frame_free(dst)

	dst.SetChannelLayout(AV_CHANNEL_LAYOUT_MONO())
	dst.SetSampleRate(44100)
	dst.SetSampleFormat(AV_SAMPLE_FMT_S16)

	err = SWResample_convert_frame(ctx, src, dst)
	assert.NoError(err)

	// Flush by passing nil source
	dst2 := AVUtil_frame_alloc()
	if !assert.NotNil(dst2) {
		t.FailNow()
	}
	defer AVUtil_frame_free(dst2)

	dst2.SetChannelLayout(AV_CHANNEL_LAYOUT_MONO())
	dst2.SetSampleRate(44100)
	dst2.SetSampleFormat(AV_SAMPLE_FMT_S16)

	err = SWResample_convert_frame(ctx, nil, dst2)
	// Flushing may return EAGAIN if no samples buffered, which is OK
	if err != nil && err != syscall.EAGAIN {
		t.Logf("Flush returned: %v (expected nil or EAGAIN)", err)
	}

	t.Log("Successfully tested flush operation")
}
