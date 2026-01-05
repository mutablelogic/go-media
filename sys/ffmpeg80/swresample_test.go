package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_swresample_context_string(t *testing.T) {
	assert := assert.New(t)

	// Test nil context
	var ctx *SWRContext
	assert.Equal("<nil>", ctx.String(), "nil context should return \"<nil>\"")

	// Test valid context
	ctx = SWResample_alloc()
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWResample_free(ctx)

	str := ctx.String()
	assert.Equal("<SWRContext>", str, "Context String() should return \"<SWRContext>\"")

	t.Logf("Context string representation: %s", str)
}

func Test_swresample_alloc_free(t *testing.T) {
	assert := assert.New(t)

	ctx := SWResample_alloc()
	assert.NotNil(ctx, "Alloc should return non-nil context")

	// Context should not be initialized yet
	assert.False(SWResample_is_initialized(ctx), "Context should not be initialized without init")

	SWResample_free(ctx)
}

func Test_swresample_init_close(t *testing.T) {
	assert := assert.New(t)

	ctx := SWResample_alloc()
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWResample_free(ctx)

	// Set options before init
	in_layout := AV_CHANNEL_LAYOUT_STEREO()
	out_layout := AV_CHANNEL_LAYOUT_MONO()

	err := SWResample_set_opts(ctx, out_layout, AV_SAMPLE_FMT_S16, 44100, in_layout, AV_SAMPLE_FMT_FLTP, 48000)
	assert.NoError(err, "Set opts should succeed")

	// Initialize
	err = SWResample_init(ctx)
	assert.NoError(err, "Init should succeed")
	assert.True(SWResample_is_initialized(ctx), "Context should be initialized after init")

	// Close
	SWResample_close(ctx)
	assert.False(SWResample_is_initialized(ctx), "Context should not be initialized after close")
}

func Test_swresample_stereo_to_mono(t *testing.T) {
	assert := assert.New(t)

	ctx := SWResample_alloc()
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWResample_free(ctx)

	// Stereo 48kHz float -> Mono 44.1kHz S16
	in_layout := AV_CHANNEL_LAYOUT_STEREO()
	out_layout := AV_CHANNEL_LAYOUT_MONO()

	err := SWResample_set_opts(ctx, out_layout, AV_SAMPLE_FMT_S16, 44100, in_layout, AV_SAMPLE_FMT_FLTP, 48000)
	if !assert.NoError(err) {
		t.FailNow()
	}

	err = SWResample_init(ctx)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.True(SWResample_is_initialized(ctx))
	t.Log("Successfully configured stereo to mono resampling")
}

func Test_swresample_mono_to_stereo(t *testing.T) {
	assert := assert.New(t)

	ctx := SWResample_alloc()
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWResample_free(ctx)

	// Mono 44.1kHz S16 -> Stereo 48kHz float
	in_layout := AV_CHANNEL_LAYOUT_MONO()
	out_layout := AV_CHANNEL_LAYOUT_STEREO()

	err := SWResample_set_opts(ctx, out_layout, AV_SAMPLE_FMT_FLTP, 48000, in_layout, AV_SAMPLE_FMT_S16, 44100)
	if !assert.NoError(err) {
		t.FailNow()
	}

	err = SWResample_init(ctx)
	if !assert.NoError(err) {
		t.FailNow()
	}

	assert.True(SWResample_is_initialized(ctx))
	t.Log("Successfully configured mono to stereo resampling")
}

func Test_swresample_sample_rate_conversion(t *testing.T) {
	tests := []struct {
		name        string
		in_rate     int
		out_rate    int
		in_layout   AVChannelLayout
		out_layout  AVChannelLayout
		in_format   AVSampleFormat
		out_format  AVSampleFormat
	}{
		{
			name:       "44.1kHz to 48kHz",
			in_rate:    44100,
			out_rate:   48000,
			in_layout:  AV_CHANNEL_LAYOUT_STEREO(),
			out_layout: AV_CHANNEL_LAYOUT_STEREO(),
			in_format:  AV_SAMPLE_FMT_S16,
			out_format: AV_SAMPLE_FMT_S16,
		},
		{
			name:       "48kHz to 44.1kHz",
			in_rate:    48000,
			out_rate:   44100,
			in_layout:  AV_CHANNEL_LAYOUT_STEREO(),
			out_layout: AV_CHANNEL_LAYOUT_STEREO(),
			in_format:  AV_SAMPLE_FMT_FLTP,
			out_format: AV_SAMPLE_FMT_FLTP,
		},
		{
			name:       "22.05kHz to 48kHz upsampling",
			in_rate:    22050,
			out_rate:   48000,
			in_layout:  AV_CHANNEL_LAYOUT_MONO(),
			out_layout: AV_CHANNEL_LAYOUT_MONO(),
			in_format:  AV_SAMPLE_FMT_S16,
			out_format: AV_SAMPLE_FMT_S16,
		},
		{
			name:       "96kHz to 48kHz downsampling",
			in_rate:    96000,
			out_rate:   48000,
			in_layout:  AV_CHANNEL_LAYOUT_STEREO(),
			out_layout: AV_CHANNEL_LAYOUT_STEREO(),
			in_format:  AV_SAMPLE_FMT_FLT,
			out_format: AV_SAMPLE_FMT_FLT,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			testAssert := assert.New(t)

			ctx := SWResample_alloc()
			if !testAssert.NotNil(ctx) {
				t.FailNow()
			}
			defer SWResample_free(ctx)

			err := SWResample_set_opts(ctx, tt.out_layout, tt.out_format, tt.out_rate, tt.in_layout, tt.in_format, tt.in_rate)
			if !testAssert.NoError(err) {
				t.FailNow()
			}

			err = SWResample_init(ctx)
			testAssert.NoError(err, "Init should succeed for %s", tt.name)
			testAssert.True(SWResample_is_initialized(ctx))

			t.Logf("Successfully configured %s", tt.name)
		})
	}
}

func Test_swresample_convert_frame(t *testing.T) {
	assert := assert.New(t)

	ctx := SWResample_alloc()
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWResample_free(ctx)

	// Create input and output frames
	src := AVUtil_frame_alloc()
	if !assert.NotNil(src) {
		t.FailNow()
	}
	defer AVUtil_frame_free(src)

	dst := AVUtil_frame_alloc()
	if !assert.NotNil(dst) {
		t.FailNow()
	}
	defer AVUtil_frame_free(dst)

	// Configure source frame: stereo, 48kHz, 1024 samples
	src.SetSampleRate(48000)
	src.SetNumSamples(1024)
	src.SetSampleFormat(AV_SAMPLE_FMT_FLTP)
	src.SetChannelLayout(AV_CHANNEL_LAYOUT_STEREO())

	// Allocate buffers for source
	err := AVUtil_frame_get_buffer(src, false)
	if !assert.NoError(err) {
		t.FailNow()
	}

	// Configure destination frame: mono, 44.1kHz
	dst.SetSampleRate(44100)
	dst.SetSampleFormat(AV_SAMPLE_FMT_S16)
	dst.SetChannelLayout(AV_CHANNEL_LAYOUT_MONO())

	// Convert using convert_frame (handles config automatically)
	err = SWResample_convert_frame(ctx, src, dst)
	assert.NoError(err, "Convert frame should succeed")

	assert.True(SWResample_is_initialized(ctx), "Context should be initialized after convert_frame")

	t.Log("Successfully converted stereo 48kHz -> mono 44.1kHz using convert_frame")
}
