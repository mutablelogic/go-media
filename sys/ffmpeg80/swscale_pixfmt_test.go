package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test YUV to RGB conversion
func Test_swscale_yuv_to_rgb(t *testing.T) {
	assert := assert.New(t)

	ctx := SWScale_get_context(
		320, 240, AV_PIX_FMT_YUV420P,
		320, 240, AV_PIX_FMT_RGB24,
		SWS_BILINEAR, nil, nil, nil,
	)
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWScale_free_context(ctx)

	src := AVUtil_frame_alloc()
	defer AVUtil_frame_free(src)
	dst := AVUtil_frame_alloc()
	defer AVUtil_frame_free(dst)

	src.SetWidth(320)
	src.SetHeight(240)
	src.SetPixFmt(AV_PIX_FMT_YUV420P)

	dst.SetWidth(320)
	dst.SetHeight(240)
	dst.SetPixFmt(AV_PIX_FMT_RGB24)

	if err := AVUtil_frame_get_buffer(src, false); !assert.NoError(err) {
		t.FailNow()
	}
	if err := AVUtil_frame_get_buffer(dst, false); !assert.NoError(err) {
		t.FailNow()
	}

	err := SWScale_scale_frame(ctx, dst, src, false)
	assert.NoError(err, "YUV420P to RGB24 conversion should succeed")

	t.Log("Successfully converted YUV420P to RGB24")
}

// Test RGB to YUV conversion
func Test_swscale_rgb_to_yuv(t *testing.T) {
	assert := assert.New(t)

	ctx := SWScale_get_context(
		640, 480, AV_PIX_FMT_RGB24,
		640, 480, AV_PIX_FMT_YUV420P,
		SWS_BILINEAR, nil, nil, nil,
	)
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWScale_free_context(ctx)

	src := AVUtil_frame_alloc()
	defer AVUtil_frame_free(src)
	dst := AVUtil_frame_alloc()
	defer AVUtil_frame_free(dst)

	src.SetWidth(640)
	src.SetHeight(480)
	src.SetPixFmt(AV_PIX_FMT_RGB24)

	dst.SetWidth(640)
	dst.SetHeight(480)
	dst.SetPixFmt(AV_PIX_FMT_YUV420P)

	if err := AVUtil_frame_get_buffer(src, false); !assert.NoError(err) {
		t.FailNow()
	}
	if err := AVUtil_frame_get_buffer(dst, false); !assert.NoError(err) {
		t.FailNow()
	}

	err := SWScale_scale_frame(ctx, dst, src, false)
	assert.NoError(err, "RGB24 to YUV420P conversion should succeed")

	t.Log("Successfully converted RGB24 to YUV420P")
}

// Test YUV422 to YUV420 conversion
func Test_swscale_yuv422_to_yuv420(t *testing.T) {
	assert := assert.New(t)

	ctx := SWScale_get_context(
		1920, 1080, AV_PIX_FMT_YUV422P,
		1920, 1080, AV_PIX_FMT_YUV420P,
		SWS_BILINEAR, nil, nil, nil,
	)
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWScale_free_context(ctx)

	src := AVUtil_frame_alloc()
	defer AVUtil_frame_free(src)
	dst := AVUtil_frame_alloc()
	defer AVUtil_frame_free(dst)

	src.SetWidth(1920)
	src.SetHeight(1080)
	src.SetPixFmt(AV_PIX_FMT_YUV422P)

	dst.SetWidth(1920)
	dst.SetHeight(1080)
	dst.SetPixFmt(AV_PIX_FMT_YUV420P)

	if err := AVUtil_frame_get_buffer(src, false); !assert.NoError(err) {
		t.FailNow()
	}
	if err := AVUtil_frame_get_buffer(dst, false); !assert.NoError(err) {
		t.FailNow()
	}

	err := SWScale_scale_frame(ctx, dst, src, false)
	assert.NoError(err, "YUV422P to YUV420P conversion should succeed")

	t.Log("Successfully converted YUV422P to YUV420P")
}

// Test upscaling (small to large)
func Test_swscale_upscale(t *testing.T) {
	assert := assert.New(t)

	// Test 4x upscale
	ctx := SWScale_get_context(
		320, 240, AV_PIX_FMT_YUV420P,
		1280, 960, AV_PIX_FMT_YUV420P,
		SWS_BICUBIC, nil, nil, nil,
	)
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWScale_free_context(ctx)

	src := AVUtil_frame_alloc()
	defer AVUtil_frame_free(src)
	dst := AVUtil_frame_alloc()
	defer AVUtil_frame_free(dst)

	src.SetWidth(320)
	src.SetHeight(240)
	src.SetPixFmt(AV_PIX_FMT_YUV420P)

	dst.SetWidth(1280)
	dst.SetHeight(960)
	dst.SetPixFmt(AV_PIX_FMT_YUV420P)

	if err := AVUtil_frame_get_buffer(src, false); !assert.NoError(err) {
		t.FailNow()
	}
	if err := AVUtil_frame_get_buffer(dst, false); !assert.NoError(err) {
		t.FailNow()
	}

	err := SWScale_scale_frame(ctx, dst, src, false)
	assert.NoError(err, "4x upscaling should succeed")

	t.Log("Successfully upscaled from 320x240 to 1280x960")
}

// Test downscaling (large to small)
func Test_swscale_downscale(t *testing.T) {
	assert := assert.New(t)

	// Test 4x downscale
	ctx := SWScale_get_context(
		1920, 1080, AV_PIX_FMT_YUV420P,
		480, 270, AV_PIX_FMT_YUV420P,
		SWS_AREA, nil, nil, nil, // AREA is good for downscaling
	)
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWScale_free_context(ctx)

	src := AVUtil_frame_alloc()
	defer AVUtil_frame_free(src)
	dst := AVUtil_frame_alloc()
	defer AVUtil_frame_free(dst)

	src.SetWidth(1920)
	src.SetHeight(1080)
	src.SetPixFmt(AV_PIX_FMT_YUV420P)

	dst.SetWidth(480)
	dst.SetHeight(270)
	dst.SetPixFmt(AV_PIX_FMT_YUV420P)

	if err := AVUtil_frame_get_buffer(src, false); !assert.NoError(err) {
		t.FailNow()
	}
	if err := AVUtil_frame_get_buffer(dst, false); !assert.NoError(err) {
		t.FailNow()
	}

	err := SWScale_scale_frame(ctx, dst, src, false)
	assert.NoError(err, "4x downscaling should succeed")

	t.Log("Successfully downscaled from 1920x1080 to 480x270")
}

// Test odd dimensions
func Test_swscale_odd_dimensions(t *testing.T) {
	assert := assert.New(t)

	// Test with odd width/height
	ctx := SWScale_get_context(
		321, 241, AV_PIX_FMT_YUV420P,
		641, 481, AV_PIX_FMT_YUV420P,
		SWS_BILINEAR, nil, nil, nil,
	)
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWScale_free_context(ctx)

	src := AVUtil_frame_alloc()
	defer AVUtil_frame_free(src)
	dst := AVUtil_frame_alloc()
	defer AVUtil_frame_free(dst)

	src.SetWidth(321)
	src.SetHeight(241)
	src.SetPixFmt(AV_PIX_FMT_YUV420P)

	dst.SetWidth(641)
	dst.SetHeight(481)
	dst.SetPixFmt(AV_PIX_FMT_YUV420P)

	if err := AVUtil_frame_get_buffer(src, false); !assert.NoError(err) {
		t.FailNow()
	}
	if err := AVUtil_frame_get_buffer(dst, false); !assert.NoError(err) {
		t.FailNow()
	}

	err := SWScale_scale_frame(ctx, dst, src, false)
	assert.NoError(err, "Scaling with odd dimensions should succeed")

	t.Log("Successfully scaled odd dimensions 321x241 to 641x481")
}

// Test very small dimensions
func Test_swscale_small_dimensions(t *testing.T) {
	assert := assert.New(t)

	// Test 16x16 (smallest reasonable size)
	ctx := SWScale_get_context(
		16, 16, AV_PIX_FMT_YUV420P,
		32, 32, AV_PIX_FMT_YUV420P,
		SWS_POINT, nil, nil, nil,
	)
	if !assert.NotNil(ctx) {
		t.FailNow()
	}
	defer SWScale_free_context(ctx)

	src := AVUtil_frame_alloc()
	defer AVUtil_frame_free(src)
	dst := AVUtil_frame_alloc()
	defer AVUtil_frame_free(dst)

	src.SetWidth(16)
	src.SetHeight(16)
	src.SetPixFmt(AV_PIX_FMT_YUV420P)

	dst.SetWidth(32)
	dst.SetHeight(32)
	dst.SetPixFmt(AV_PIX_FMT_YUV420P)

	if err := AVUtil_frame_get_buffer(src, false); !assert.NoError(err) {
		t.FailNow()
	}
	if err := AVUtil_frame_get_buffer(dst, false); !assert.NoError(err) {
		t.FailNow()
	}

	err := SWScale_scale_frame(ctx, dst, src, false)
	assert.NoError(err, "Scaling small dimensions should succeed")

	t.Log("Successfully scaled 16x16 to 32x32")
}

// Test different scaling algorithms
func Test_swscale_algorithms(t *testing.T) {
	algorithms := []struct {
		name string
		flag SWSFlag
	}{
		{"FAST_BILINEAR", SWS_FAST_BILINEAR},
		{"BILINEAR", SWS_BILINEAR},
		{"BICUBIC", SWS_BICUBIC},
		{"POINT", SWS_POINT},
		{"AREA", SWS_AREA},
		{"LANCZOS", SWS_LANCZOS},
	}

	for _, alg := range algorithms {
		t.Run(alg.name, func(t *testing.T) {
			assert := assert.New(t)

			ctx := SWScale_get_context(
				640, 480, AV_PIX_FMT_YUV420P,
				1280, 720, AV_PIX_FMT_YUV420P,
				alg.flag, nil, nil, nil,
			)
			if !assert.NotNil(ctx) {
				t.FailNow()
			}
			defer SWScale_free_context(ctx)

			src := AVUtil_frame_alloc()
			defer AVUtil_frame_free(src)
			dst := AVUtil_frame_alloc()
			defer AVUtil_frame_free(dst)

			src.SetWidth(640)
			src.SetHeight(480)
			src.SetPixFmt(AV_PIX_FMT_YUV420P)

			dst.SetWidth(1280)
			dst.SetHeight(720)
			dst.SetPixFmt(AV_PIX_FMT_YUV420P)

			if err := AVUtil_frame_get_buffer(src, false); !assert.NoError(err) {
				t.FailNow()
			}
			if err := AVUtil_frame_get_buffer(dst, false); !assert.NoError(err) {
				t.FailNow()
			}

			err := SWScale_scale_frame(ctx, dst, src, false)
			assert.NoError(err, "Scaling with %s should succeed", alg.name)

			t.Logf("Successfully scaled with %s algorithm", alg.name)
		})
	}
}
